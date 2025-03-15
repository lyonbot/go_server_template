package misc

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type MemoTask[Q any, R any] struct {
	ToKey         func(args Q) string
	Execute       func(args Q) (*R, error)
	CacheDuration time.Duration

	lru *lru.Cache[string, *MemoTaskCacheItem[R]]
}

type MemoTaskCacheItem[R any] struct {
	cond        *sync.Cond
	value       *R
	err         error
	expireTimer *time.Timer
	expired     bool
}

func MakeMemoTask[Q any, R any](lruCount int, cacheDuration time.Duration, toKey func(args Q) string, execute func(args Q) (*R, error)) *MemoTask[Q, R] {
	l, _ := lru.New[string, *MemoTaskCacheItem[R]](lruCount)
	return &MemoTask[Q, R]{
		ToKey:         toKey,
		Execute:       execute,
		CacheDuration: cacheDuration,
		lru:           l,
	}
}

func (t *MemoTask[Q, R]) ExecuteWithCache(args Q) (*R, error) {

	key := t.ToKey(args)

	cacheItem := &MemoTaskCacheItem[R]{
		cond: sync.NewCond(&sync.Mutex{}),
	}
	cacheItem.cond.L.Lock()
	defer cacheItem.cond.L.Unlock()

	if prev, ok, _ := t.lru.PeekOrAdd(key, cacheItem); ok {
		t.lru.Get(key)
		if prev.cond != nil {
			prev.cond.Wait()
		}
		if !prev.expired && prev.expireTimer != nil {
			// 如果 expireTimer 已经触发并执行完毕，调用 Reset 会导致 panic，需要检查 timer 状态
			prev.expireTimer.Reset(t.CacheDuration)
		}
		return prev.value, prev.err
	}

	// do the work
	result, err := t.Execute(args)
	cacheItem.value = result
	cacheItem.err = err

	cleanup := func() {
		cacheItem.expired = true
		t.lru.Remove(key)
	}

	if err != nil {
		cleanup()
	} else {
		cacheItem.expireTimer = time.AfterFunc(t.CacheDuration, cleanup)
	}

	cacheItem.cond.Broadcast()
	cacheItem.cond = nil

	return result, err
}

func (t *MemoTask[Q, R]) ClearCache() {
	t.lru.Purge()
}

func (t *MemoTask[Q, R]) ClearCacheSingle(args Q) {
	key := t.ToKey(args)
	t.lru.Remove(key)
}
