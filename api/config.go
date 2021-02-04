package api

import (
	"github.com/gin-gonic/gin"
	"goblog/core/global"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/service"
	"goblog/service/admin"
)

// HomePage 首页统计数据
func HomePage(c *gin.Context) {
	resp := entity.NewResponse()
	db := global.GDb
	// 数量统计
	computerCount, groupsCount, dssCount, updatesCount := admin.GetHomeCountInfo(db)
	// MSRCSeverity 严重程度
	MSRCSeverity := admin.GetMSRCSeverityPie(db)
	// 计算机概览，所有计算机对所有更新的安装状态的统计
	computerUpdateStats := admin.ComputerUpdateStatsPie(db, computerCount, updatesCount)
	// 更新状态
	updateClassify := admin.GetUpdateClassifyPie(db)

	var countStats = map[string]int64{
		"computer_count": computerCount,
		"update_count":   updatesCount,
		"group_count":    groupsCount - 3,
		"dss_count":      dssCount,
	}

	data := map[string]interface{}{
		"count_stats":        countStats,
		"msrc_severity":      MSRCSeverity,
		"all_computer_stats": computerUpdateStats,
		"update_classify":    updateClassify,
	}
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// 获取上游服务器配置
func ParentUssConfigs(c *gin.Context) {
	resp := entity.NewResponse()
	uss := admin.GetParentUssConfigs()
	resp.SetData(uss)
	response.Ok(resp, c)
	return
}

// 修改上游服务器配置
func EditParentUssConfig(c *gin.Context) {
	resp := entity.NewResponse()
	var params = make([]map[string]interface{}, 2)
	var err error
	if err = c.BindJSON(&params); err == nil {
		err = admin.EditParentUssConfig(params)
	}
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	response.Ok(resp, c)
	return
}

// 获取同步定时任务信息
func GetSyncSchedule(c *gin.Context) {
	resp := entity.NewResponse()
	data, err := admin.GetSyncSchedule()
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// 设置同步定时任务信息
func SetSyncSchedule(c *gin.Context) {
	resp := entity.NewResponse()
	var syncSchedule map[string]interface{}
	var err error
	err = c.BindJSON(&syncSchedule)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	err = admin.SetSyncSchedule(syncSchedule)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	response.Ok(resp, c)
	return
}

// 设置是否开启Express
func SetExpressEnable(c *gin.Context) {
	resp := entity.NewResponse()
	var expressEnable map[string]interface{}
	var err error
	err = c.BindJSON(&expressEnable)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	err = admin.SetExpressEnable(expressEnable)
	if err != nil {
		resp.SetMsg(err.Error())
		response.Fail(resp, c)
		return
	}
	response.Ok(resp, c)
	return
}

// 获取服务器设置
func GetServerConfig(c *gin.Context) {
	resp := entity.NewResponse()
	data := service.GetServerConfig()
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// 获取操作系统
func GetOsVersion(c *gin.Context) {
	resp := entity.NewResponse()
	isLiveStats := c.DefaultQuery("is_live_stats", "false")
	data := admin.GetOsVersion(isLiveStats)
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

func GetServerInfo(c *gin.Context) {
	resp := entity.NewResponse()
	data := admin.GetServerInfo()
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

func GetPending(c *gin.Context) {
	resp := entity.NewResponse()
	data := admin.GetPending()
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// 获取revision个数
func GetRevisionCount(c *gin.Context) {
	resp := entity.NewResponse()
	data := admin.GetRevisionCount()
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

//获取前端所需数据的map
func GetAllMap(c *gin.Context) {
	resp := entity.NewResponse()
	data := admin.GetAllMap()
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

func VersionCheck(c *gin.Context) {
	resp := entity.NewResponse()
	platform := c.Query("platform")
	data, err := admin.VersionCheck(platform)
	if err != nil {
		resp.SetMsg("版本检查失败.")
		response.Fail(resp, c)
		return
	}
	resp.SetData(data)
	response.Ok(resp, c)
	return
}

func Version(c *gin.Context) {
	resp := entity.NewResponse()
	data, err := admin.Version()
	if err != nil {
		resp.SetMsg("版本检查失败.")
		response.Fail(resp, c)
		return
	}
	resp.SetData(data)
	response.Ok(resp, c)
	return
}
