package main

import (
	"goblog/core"
	_ "goblog/core"
	"goblog/core/global"
	initialize "goblog/core/init"
	"goblog/core/server"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

func main() {
	// 创建监听退出chan
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				// 注意： 此处信号监听在goland下不生效
				os.Exit(0)
			}
		}
	}()
	// 初始化Viper
	global.GVP = core.Viper()
	// 初始化zap日志库
	global.GLog = core.Zap()
	// 初始化gorm
	global.GDb = initialize.Gorm()
	// 初始化表
	initialize.MysqlTables(global.GDb)
	// 初始化redis服务
	initialize.Redis()

	db, _ := global.GDb.DB()
	defer func() {
		// 关闭数据库链接
		if err := db.Close(); err != nil {
			global.GLog.Error("error:", zap.Any("err", err))
		}
	}()
	// server.RunServer()
	g.Go(func() error {
		return server.RunServer()
	})
	if err := g.Wait(); err != nil {
		global.GLog.Error("error:", zap.Any("err", err))
	}
}
