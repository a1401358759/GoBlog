package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/service/admin"
	"goblog/utils"
	"strconv"
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
func Login(c *gin.Context) {
	var loginParams struct {
		Email    string `json:"Email" binding:"required"`
		Password string `json:"Password" binding:"required"`
	}
	resp := entity.NewResponse()
	err := c.ShouldBindJSON(&loginParams)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	uid, token, err := admin.LoginCheck(loginParams.Email, loginParams.Password)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	resp.SetData(map[string]interface{}{"Uid": uid, "Token": token, "Email": loginParams.Email})
	desc := fmt.Sprintf("用户【%s】登录当前系统", loginParams.Email)
	go admin.GenOperateRecord(admin.OperateAction.LOGIN, desc, admin.ResultStatus.SUCCESS, uid, loginParams.Email)
	response.Ok(resp, c)
	return
}

func ChangePwd(c *gin.Context) {
	var ChangePwdParams struct {
		Email       string `json:"Email" binding:"required"`
		Password    string `json:"Password" binding:"required"`
		NewPassword string `json:"NewPassword" binding:"required"`
	}
	resp := entity.NewResponse()
	err := c.ShouldBindJSON(&ChangePwdParams)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	if err := admin.ChangePwd(ChangePwdParams.Email, ChangePwdParams.Password, ChangePwdParams.NewPassword); err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	} else {
		uid, email := admin.GetUserInfo(c)
		desc := fmt.Sprintf("用户【%s】修改了系统密码", email)
		go admin.GenOperateRecord(admin.OperateAction.ChangePwd, desc, admin.ResultStatus.SUCCESS, uid, email)
		response.Ok(resp, c)
		return
	}
}

func Logout(c *gin.Context) {
	resp := entity.NewResponse()
	uid, email := admin.GetUserInfo(c)
	admin.Logout(uid)
	desc := fmt.Sprintf("用户【%s】从当前系统注销", email)
	go admin.GenOperateRecord(admin.OperateAction.LOGOUT, desc, admin.ResultStatus.SUCCESS, uid, email)
	response.Ok(resp, c)
	return
}

func Revision(c *gin.Context) {
	resp := entity.NewResponse()
	revisionID, err := strconv.Atoi(c.Param("revisionId"))
	if err != nil {
		resp.SetMsg("参数获取失败" + err.Error())
		response.Fail(resp, c)
		return
	}
	revisionDetail := admin.GetRevisionDetail(revisionID)
	resp.SetData(revisionDetail)
	response.Ok(resp, c)
	return
}
