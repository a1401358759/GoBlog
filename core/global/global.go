package global

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"goblog/config"
	"gorm.io/gorm"
)

var (
	GDb     *gorm.DB
	GRedis  *redis.Client
	GConfig config.Server
	GVP     *viper.Viper
	GLog    *zap.Logger
)
