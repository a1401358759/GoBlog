package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/utils"
)

func About404(c *gin.Context) {
	c.JSON(404, gin.H{
		"message": "路由错误",
	})
	return
}

type SimpleParams struct {
	Sort       string `form:"Sort"`
	PageSize   int    `form:"PageSize"`
	PageNumber int    `form:"PageNumber"`
}

type QueryParams interface {
	Handle(c *gin.Context) (*SimpleParams, []utils.QueryFilter, error)
}

func NewMethodQueryParam(method string) QueryParams {
	var methodMapping = map[string]QueryParams{
		"POST": &QueryPostParams{},
		"GET":  &QueryGetParams{},
	}
	return methodMapping[method]
}

type QueryPostParams struct {
	SimpleParams
	Filter []utils.QueryFilter `form:"Filter"`
}

func (q *QueryPostParams) Handle(c *gin.Context) (*SimpleParams, []utils.QueryFilter, error) {
	if err := c.BindJSON(q); err != nil {
		return nil, nil, err
	}
	return &q.SimpleParams, q.Filter, nil
}

type QueryGetParams struct {
	SimpleParams
	Filter string `form:"Filter"`
}

func (q *QueryGetParams) Handle(c *gin.Context) (*SimpleParams, []utils.QueryFilter, error) {
	var filter []utils.QueryFilter
	var err error
	if err = c.BindQuery(q); err != nil {
		return nil, nil, err
	}
	if len(q.Filter) > 0 {
		if err = json.Unmarshal([]byte(q.Filter), &filter); err != nil {
			return nil, nil, err
		}
	}
	return &q.SimpleParams, filter, nil
}

// request 列表 query 标准参数解析
// 因有部分特殊接口需要试用POST请求获取列表，所以需要根据请求Method不同，做不同处理
// 此处简单使用了Golang的继承，组合，多态特性和简单工厂模式.
func QueryHandle(c *gin.Context) (*SimpleParams, []utils.QueryFilter, error) {
	return NewMethodQueryParam(c.Request.Method).Handle(c)
}

// @Tags User
// @Summary 用户登录
// @Accept application/json
// @Produce  application/json
// @Success 200 {string} string "{"code":10001,"data":{},"msg":"成功"}"
// @Router /user/login [get]

// func Login(c *gin.Context) {
// 	var loginParams struct {
// 		Email    string `json:"Email" binding:"required"`
// 		Password string `json:"Password" binding:"required"`
// 	}
// 	resp := entity.NewResponse()
// 	err := c.ShouldBindJSON(&loginParams)
// 	if err != nil {
// 		resp.SetMsg(err.Error())
// 		response.Fail(resp, c)
// 		return
// 	}
// 	uid, token, err := admin.LoginCheck(loginParams.Email, loginParams.Password)
// 	if err != nil {
// 		resp.SetMsg(err.Error())
// 		response.Fail(resp, c)
// 		return
// 	}
// 	resp.SetData(map[string]interface{}{"Uid": uid, "Token": token, "Email": loginParams.Email})
// 	response.Ok(resp, c)
// 	return
// }
//
// func ChangePwd(c *gin.Context) {
// 	var ChangePwdParams struct {
// 		Email       string `json:"Email" binding:"required"`
// 		Password    string `json:"Password" binding:"required"`
// 		NewPassword string `json:"NewPassword" binding:"required"`
// 	}
// 	resp := entity.NewResponse()
// 	err := c.ShouldBindJSON(&ChangePwdParams)
// 	if err != nil {
// 		resp.SetMsg(err.Error())
// 		response.Fail(resp, c)
// 		return
// 	}
// 	if err := admin.ChangePwd(ChangePwdParams.Email, ChangePwdParams.Password, ChangePwdParams.NewPassword); err != nil {
// 		resp.SetMsg(err.Error())
// 		response.Fail(resp, c)
// 		return
// 	} else {
// 		uid, email := admin.GetUserInfo(c)
// 		response.Ok(resp, c)
// 		return
// 	}
// }

func Logout(c *gin.Context) {
	resp := entity.NewResponse()
	// uid, email := admin.GetUserInfo(c)
	// admin.Logout(uid)
	response.Ok(resp, c)
	return
}

func Test(c *gin.Context) {
	return
}
