package admin

import (
	"go.uber.org/zap"
	"goblog/core/global"
	"gorm.io/gorm"
)

// StatusStatsOfPerComputer 单个计算机安装更新状态的统计
func StatusStatsOfPerComputer() {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		return nil
	})
	if err != nil {
		global.GLog.Error("StatusStatsOfPerComputer", zap.Any("err", err))
	}
}
