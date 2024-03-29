package helpers

import "github.com/go-redis/redis"

var RedisClient *redis.Client

// 初始化redis
func InitRedis() {
	var err error
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err != nil || RedisClient == nil {
		panic("init redis failed!")
	}
	_, err = RedisClient.Ping().Result()
	if err != nil {
		panic("redis ping failed")
	}
}

func CloseRedis() {
	_ = RedisClient.Close()
}
