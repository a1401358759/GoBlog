package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type VersionCheckMixin struct {
	VersionID uint `json:"version_id" gorm:"column:version_id;not null;default:1"`
}

type TimeModelMiXin struct {
	CreatedTime time.Time `gorm:"autoCreateTime;column:created_time;comment:创建时间" json:"created_time"`
	LastUpdate  time.Time `gorm:"autoCreateTime;autoUpdateTime;column:last_update;comment:最后更新时间" json:"last_update"`
}

func (VersionCheckMixin) BulkFlag() {}
