package response

import (
	"net/http"
	"reflect"

	"goblog/modules/entity"
	commonResponse "goblog/modules/entity"

	"github.com/gin-gonic/gin"
)

func RespHandler(response *commonResponse.Response, c *gin.Context) {
	c.JSON(http.StatusOK, response)
}

func XmlResult(data interface{}, c *gin.Context) {
	code := http.StatusOK
	if reflect.TypeOf(data) == reflect.TypeOf(entity.FaultEnvelope{}) {
		code = http.StatusInternalServerError
	}
	c.XML(code, data)
}

func Ok(response *commonResponse.Response, c *gin.Context) {
	if response.Code == 0 {
		response.SetCode(SUCCESS)
	}
	if response.Msg == "" {
		response.SetMsg(ErrorMsg[response.Code])
	}
	RespHandler(response, c)
}

func Fail(response *commonResponse.Response, c *gin.Context) {
	if response.Code == 0 {
		response.SetCode(FAILED)
	}
	if response.Msg == "" {
		response.SetMsg(ErrorMsg[response.Code])
	}
	RespHandler(response, c)
}
