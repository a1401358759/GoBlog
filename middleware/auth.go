package middleware

import (
	"github.com/gin-gonic/gin"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/service/admin"
)

func LoginAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 登录接口特殊处理
		if c.FullPath() == "/api/v2/user/login" || c.FullPath() == "/api/v2/config/server_info" || c.FullPath() == "/api/v2/config/mapping" {
			c.Next()
		} else {
			resp := entity.NewResponse()
			token := c.Request.Header.Get("Authentication-Token")
			if token == "" {
				resp.SetCode(response.NotLogin)
				response.Fail(resp, c)
				c.Abort()
				return
			}
			params, err := admin.NewSign().ParseToken(token)
			if err != nil {
				resp.SetCode(response.NotLogin)
				response.Fail(resp, c)
				c.Abort()
				return
			}
			c.Set("Email", params.Email)
			c.Set("UID", params.UID)
			c.Next()
		}
	}
}
