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
	color.Cyan.Println("   1. 导入更新包     -m import -f [file] -l [log_file_path]")
	color.Cyan.Println("   2. 导出更新包     -m export -f [file] -l [log_file_path]")
	color.Cyan.Println("   3. 清空数据       -m clear")
	color.Cyan.Println("   4. 上传更新数据   -m movecontent -f [file]")
	color.Cyan.Println("   5. 创建数据库     -m create_db")
	color.Cyan.Println("   6. 初始化数据库   -m init_data")
	color.Cyan.Println("   7. 新建用户       -m create_user -u [email] -w [password]")
	color.Cyan.Println("   8. 重置用户密码   -m reset_pwd -u [email] -w [password]")
	color.Cyan.Println("   9. 删除用户       -m delete_user -u [email]")
	color.Cyan.Println("  10. 查看用户       -m all_user")
	color.Cyan.Println("  11. 清空所有计算机 -m clean_all_computer")
	color.Cyan.Println("  12. 刷新redis缓存  -m flush_redis")
	color.Cyan.Println("  13. 重置导入状态   -m reset_import")
	color.Cyan.Println("")
	color.BgCyan.Println("--------------------------------------------------------------")
	color.Cyan.Println("")
}
