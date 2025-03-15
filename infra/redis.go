package infra

import (
	"github.com/redis/go-redis/v9"
	"lyonbot.github.com/my_app/misc"
)

var Rdb = redis.NewClient(&redis.Options{
	Addr: misc.Getenv("REDIS_ADDR", "localhost:6379"),
})
