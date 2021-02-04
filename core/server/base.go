package server

import (
	"fmt"
	"goblog/core/global"
	"goblog/routers"
	"time"

	"go.uber.org/zap"
)

type server interface {
	ListenAndServe() error
}

func RunServer() error {
	Router := routers.Routers()
	address := fmt.Sprintf(":%d", global.GConfig.System.Addr)
	s := initServer(address, Router)
	// 保证文本顺序输出
	time.Sleep(10 * time.Microsecond)
	global.GLog.Info("server run success on ", zap.String("address", address))
	err := s.ListenAndServe()
	//global.G_LOG.Error(s.ListenAndServe().Error())
	return err
}
