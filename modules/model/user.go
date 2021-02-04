package model

import (
	"time"
)

type User struct {
	UserID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id"`
	Email          string    `gorm:"unique;type:varchar(256);column:email"`
	UserName       string    `gorm:"type:varchar(256);column:username;default:null"`
	Password       string    `gorm:"type:varchar(256);column:password"`
	LastLoginAt    time.Time `gorm:"column:last_login_at;default:null"`
	CurrentLoginAt time.Time `gorm:"column:current_login_at;default:null"`
	LastLoginIp    string    `gorm:"column:last_login_ip;type:char(100);default:null"`
	CurrentLoginIp string    `gorm:"column:current_login_ip;type:char(100);default:null"`
	LoginCount     int       `gorm:"default:0;column:login_count"`
	Active         bool      `gorm:"default:true;column:active"`
	ConfirmedAt    time.Time `gorm:"column:confirmed_at;default:null"`
	Cookie         string    `gorm:"column:cookie;type:varchar(1000);default:null"`
	ExpiresTime    time.Time `gorm:"default:null;column:expires_time"`
}

func (User) TableName() string {
	return "user"
}

func (User) BulkFlag() {

}

type Role struct {
	RoleID      int    `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	Name        string `gorm:"unique;column:name;type:char(80)" json:"name"`
	Description string `gorm:"column:description;type:varchar(256)" json:"description"`
}

func (Role) TableName() string {
	return "role"
}

type RolesUsers struct {
	ID     int  `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	UserID int  `gorm:"null;column:user_id" json:"user_id"`
	User   User `gorm:"ForeignKey:UserID;AssociationForeignKey:UserID;not null"`
	RoleID int  `gorm:"null;column:role_id" json:"role_id"`
	Role   Role `gorm:"ForeignKey:RoleID;AssociationForeignKey:RoleID;not null"`
}

func (RolesUsers) TableName() string {
	return "roles_users"
}
