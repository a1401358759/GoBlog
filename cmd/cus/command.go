package cus

import (
	"fmt"
	"github.com/gookit/color"
	"go/importer"
	"goblog/core/global"
	"io/ioutil"
)

func Test(args []string) {
	args1 := args[1]
	args2 := args[2]
	color.Cyan.Println("This is test")
	color.Cyan.Println("arg 1:", args1, "arg 2:", args2)

	pkg, err := importer.Default().Import("model")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, declName := range pkg.Scope().Names() {
		fmt.Println(declName)
	}

}

func ImportProcedure() {
	procedures, _ := ioutil.ReadDir("modules/procedure")
	for _, p := range procedures {
		query, _ := ioutil.ReadFile("modules/procedure/" + p.Name())
		if err := global.GDb.Exec(string(query)); err.Error != nil {
			fmt.Println(err)
		} else {
			fmt.Println("import procedure " + " success!")
		}
	}
}

func RefreshIsonserver() {
	query := "update files set IsOnServer=1;"
	if err := global.GDb.Exec(query); err.Error != nil {
		fmt.Println(err)
	} else {
		fmt.Println("refresh isonserver " + " success!" + "all files are on server!")
	}
}

func SetMySQLCharacter() {
	m := global.GConfig.Mysql
	query := "SELECT CONCAT('ALTER TABLE `', TABLE_NAME, '` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;') AS target_tables FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='" + m.Dbname + "' AND TABLE_TYPE='BASE TABLE';"
	var queryString []string
	global.GDb.Raw(query).Scan(&queryString)
	for _, item := range queryString {
		if err := global.GDb.Exec(item).Error; err != nil {
			fmt.Println(err)
		}
	}
}
