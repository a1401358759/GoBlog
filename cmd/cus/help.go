package cus

import (
	"github.com/gookit/color"
)

func PrintHelp() {
	color.Cyan.Println("")
	color.BgCyan.Println("--------------------------------------------------------------")
	color.Cyan.Println("")
	color.Warn.Println("  Please enter the method in CUSToolKit by using -m")
	color.Cyan.Println("")
	color.Cyan.Println("   1. 清空数据       -m clear")
	color.Cyan.Println("   2. 创建数据库     -m create_db")
	color.Cyan.Println("")
	color.BgCyan.Println("--------------------------------------------------------------")
	color.Cyan.Println("")
}
