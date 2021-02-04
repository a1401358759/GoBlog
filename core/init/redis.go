package initialize

import (
	"goblog/core/global"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

func Redis() {
	redisConfig := global.GConfig.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		global.GLog.Error("redis connect ping failed, err:", zap.Any("err", err))
	} else {
		global.GLog.Info("redis connect ping response:", zap.String("pong", pong))
		global.GRedis = client
	}
}

func QueueClient() {
	redisConfig := global.GConfig.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       1,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		global.GLog.Error("queue redis connect ping failed, err:", zap.Any("err", err))
	} else {
		global.GLog.Info("queue redis connect ping response:", zap.String("pong", pong))
		global._ = client
	}
}
