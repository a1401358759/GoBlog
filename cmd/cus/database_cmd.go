package cus

import (
	"database/sql"
	"fmt"
	"github.com/gookit/color"
	"goblog/core/global"
	initialize "goblog/core/init"
)

func ClearData(force bool) {
	var input string
	if force == false {
		color.Warn.Println("This operation will clean all data in the Database, enter [yes] to confirm your operation: ")
		fmt.Scanln(&input)
	} else {
		input = "no"
	}

	if input == "yes" || input == "no" {
		// 创建数据库
		CreateDatabase(true)
		// 初始化gorm
		global.GDb = initialize.Gorm()
		// 初始化表
		initialize.MysqlTables(global.GDb)
		color.Warn.Println("You have cleared the database successfully")
	}
}

func CreateDatabase(delete bool) {
	switch global.GConfig.System.DbType {
	case "mysql":
		CreateMySQLDatabase(delete)
	default:
		CreateMySQLDatabase(delete)
	}
}

func InitDatabase() {
	// 初始化gorm
	global.GDb = initialize.Gorm()
	// 初始化表
	initialize.MysqlTables(global.GDb)
}

func CreateMySQLDatabase(delete bool) {
	m := global.GConfig.Mysql
	dsn := m.Username + ":" + m.Password + "@tcp(" + m.Path + ")/"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 删除数据库
	if delete {
		_, err = db.Exec("DROP DATABASE IF EXISTS `" + m.Dbname + "`;")
		if err != nil {
			panic(err)
		}
		// 初始化redis服务
		initialize.Redis()
		global.GRedis.FlushAll()
	}

	// 重新创建数据库
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + m.Dbname + "` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;")
	if err != nil {
		panic(err)
	}
}
