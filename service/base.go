package service

import (
	"goblog/core/global"
	"goblog/modules/model"
)

const (
	SyncModeAuto    = "Auto"    // 自动
	SyncModeReplica = "Replica" // 副本
	SyncModeMulti   = "Multi"   // 双源
	SyncModeCommand = "Command" // 命令行
	SyncModeFile    = "File"    // 文件
)

const (
	SyncTypeManual = iota
	SyncTypeAuto
	SyncTypeCommand
	SyncTypeApi
)

var SyncTypeMap = map[int]string{
	SyncTypeManual: "手动同步",
	SyncTypeAuto:   "自动同步",
}

var SyncStatusMap = map[int]string{
	0: "同步失败",
	1: "同步成功",
	2: "同步取消",
}

var SyncModeMap = map[string]string{
	SyncModeAuto:    "自治",
	SyncModeReplica: "副本",
	SyncModeMulti:   "多源",
	SyncModeCommand: "命令行",
	SyncModeFile:    "文件导入",
}

var SyncScope = map[int]string{
	1: "更新", // 现版本，获取更新 = 先获取产品分类，再更新 = 获取All
	2: "产品分类",
}

// 获取ServerConfig
func GetServerConfig() *model.ServerConfiguration {
	var sc model.ServerConfiguration
	global.GDb.First(&sc)
	return &sc
}
