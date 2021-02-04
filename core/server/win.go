// +build windows

package server

import (
	"goblog/core/global"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func initServer(address string, router *gin.Engine) server {
	return &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    time.Duration(global.GConfig.System.Timeout) * time.Second,
		WriteTimeout:   time.Duration(global.GConfig.System.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
