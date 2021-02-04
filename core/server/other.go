// +build !windows

package server

import (
	"goblog/core/global"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func initServer(address string, router *gin.Engine) server {
	s := endless.NewServer(address, router)
	s.ReadHeaderTimeout = time.Duration(global.GConfig.System.Timeout) * time.Second
	s.WriteTimeout = time.Duration(global.GConfig.System.Timeout) * time.Second
	s.MaxHeaderBytes = 1 << 20
	return s
}
