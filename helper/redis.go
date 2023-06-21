package helper

import (
	"github.com/go-redis/redis/v8"
)

type RedisClients struct {
	Bitcoin  *redis.Client
	Litecoin *redis.Client
	Dogecoin *redis.Client
}

func InitRedisClients() RedisClients {
	return RedisClients{
		Bitcoin: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}),
		Litecoin: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       1,
		}),
		Dogecoin: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       2,
		}),
	}
}
