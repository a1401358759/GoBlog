package model

// Author 文章作者
type Author struct {
	ID      int    `gorm:"primary_key;column:id" json:"id"`
	Name    string `gorm:"type:varchar(256);column:name;comment:姓名" json:"name"`
	Email   string `gorm:"type:varchar(128);column:email;comment:邮件;default:null" json:"email"`
	Website string `gorm:"type:varchar(128);column:website;comment:个人网站;default:null" json:"website"`
	TimeModelMiXin
}

func (Author) TableName() string {
	return "author"
}

// OwnerMessage 主人寄语
type OwnerMessage struct {
	ID      int    `gorm:"primary_key;column:id" json:"id"`
	Summary string `gorm:"type:varchar(100);column:summary;comment:简介;default:null" json:"summary"`
	Message string `gorm:"type:longtext;column:message;comment:邮件;default:null" json:"message"`
	Editor  int    `gorm:"column:website;comment:编辑器类型;default:1" json:"editor"`
	TimeModelMiXin
}

func (OwnerMessage) TableName() string {
	return "owner_message"
}
