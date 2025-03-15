package infra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/semaphore"
)

// LimitedRequestor 是一个基于 Redis 的请求合并和缓存器，附带限流的能力。
// 如果有并发请求，会等待第一个请求出结果，避免并发运行实际的逻辑。
// 结果会被 github.com/bytedance/sonic 序列、反序列化。
type LimitedRequestor[Q any, R any] struct {
	Sem              *semaphore.Weighted // 可选，用于限流
	TaskKeyGenerator func(query Q) string
	Action           func(query Q) (*R, error)
}

const FLAG_LEN = 1
const FLAG_PENDING = "P"
const FLAG_RESULT = "R"
const FLAG_ERROR = "E"
const FLAG_CANCELLED = "C" // 发请求的那个进程放弃了执行，其他等待者可以拿来执行权

const MAX_PENDING_DURATION = time.Minute * 20
const MAX_ERROR_DURATION = time.Minute * 5
const MAX_RESULT_DURATION = time.Hour * 1

func (r *LimitedRequestor[Q, R]) Request(ctx context.Context, query Q) (*R, error) {
	if r.TaskKeyGenerator == nil || r.Action == nil {
		return nil, errors.New("LimitedRequestor: TaskKeyGenerator or Action is nil")
	}

	rdbKey := r.TaskKeyGenerator(query)
	rdb := Rdb

	wasWaitingForActor := false
	backoff := 100
	maxBackoff := 1000

	for {
		prevValue, err := rdb.Get(ctx, rdbKey).Result()
		if err == redis.Nil {
			// Key doesn't exist, try to acquire lock
			success, err := rdb.SetNX(ctx, rdbKey, FLAG_PENDING, MAX_PENDING_DURATION).Result()
			if err != nil {
				return nil, fmt.Errorf("redis error: %w", err)
			}
			if success {
				return r.doRealRequestAndCache(ctx, rdbKey, query)
			}
			// Someone else got the lock, continue the loop
			continue
		}

		// 出现错误或者执行者取消了请求
		if strings.HasPrefix(prevValue, FLAG_ERROR) && wasWaitingForActor {
			return nil, errors.New(prevValue[FLAG_LEN:])
		}

		// 有结果
		if strings.HasPrefix(prevValue, FLAG_RESULT) {
			serialized := prevValue[FLAG_LEN:]
			result := new(R)
			if err := sonic.UnmarshalString(serialized, &result); err != nil {
				return nil, err
			}
			return result, nil
		}

		// 其他进程正在处理请求，等待
		if strings.HasPrefix(prevValue, FLAG_PENDING) {
			wasWaitingForActor = true
			if backoff >= maxBackoff {
				backoff = maxBackoff
			} else {
				backoff *= 2
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(backoff) * time.Millisecond):
				continue
			}
		}

		// 剩下的情况可以拿来执行权，发请求
		// 1. 执行者是错误，但是本教程没有在等待 （ !wasWaitingForActor ）
		// 2. 执行者取消执行
		// 3. 其他脏数据
		prev2, err := rdb.GetSet(ctx, rdbKey, FLAG_PENDING).Result()
		if err != nil && err != redis.Nil {
			return nil, fmt.Errorf("redis error: %w", err)
		}
		if prev2 == prevValue || err == redis.Nil {
			// 未发生并发，没有别人在处理请求了
			return r.doRealRequestAndCache(ctx, rdbKey, query)
		}
	}
}

// 发起实际的请求
//
// 传入的 ctx 仅用于做并发控制。一旦获得执行权，实际执行过程不会因为传入的ctx而被暂停。
func (r *LimitedRequestor[Q, R]) doRealRequestAndCache(waitCtx context.Context, rdbKey string, query Q) (*R, error) {
	// 如果真的发起了请求，则在服务器后台处理
	rdb := Rdb
	actionCtx := context.Background()
	since := time.Now()
	sem := r.Sem

	// 刚开始解析，限流处理
	if sem != nil {
		if err := sem.Acquire(waitCtx, 1); err != nil {
			// 这个请求被停掉了，不需要等待执行
			rdb.Set(actionCtx, rdbKey, FLAG_CANCELLED+"request cancelled", MAX_ERROR_DURATION)
			return nil, err
		}
		defer sem.Release(1)
	}

	// 发起请求
	result, err := r.Action(query)
	if err != nil {
		log.Printf("[LimitedRequestor][%s][%dms] error: %s", rdbKey, time.Since(since).Milliseconds(), err.Error())
		rdb.Set(actionCtx, rdbKey, FLAG_ERROR+err.Error(), MAX_ERROR_DURATION)
		return nil, err
	}

	log.Printf("[LimitedRequestor][%s][%dms] ok", rdbKey, time.Since(since).Milliseconds())

	resultSerialized, err := sonic.MarshalString(&result)
	if err != nil {
		rdb.Set(actionCtx, rdbKey, FLAG_ERROR+err.Error(), MAX_ERROR_DURATION)
		return nil, err
	}

	rdb.Set(actionCtx, rdbKey, FLAG_RESULT+resultSerialized, MAX_RESULT_DURATION)
	return result, nil
}
