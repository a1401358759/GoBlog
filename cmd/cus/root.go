package cus

import (
	"github.com/gookit/color"
	"goblog/core"
	"goblog/core/global"
	initialize "goblog/core/init"
	"os"
)

// Execute executes the root command.
func Execute() {
	run(os.Args)
}

func init() {
	// 初始化Viper
	global._ = core.Viper("config.ini")
	// 初始化zap日志库
	global.GLog = core.Zap()
}

func run(args []string) {
	if len(args) < 2 || (args[1] != "-h" && args[1] != "-m") {
		color.Error.Println("Parameter error, \nThe parameter length should be greater than 2, \nThe first parameter must be '-h' or '-m'")
	}
	if args[1] == "-h" || args[1] == "--help" {
		PrintHelp()
		return
	} else if args[1] == "-m" {
		if args[2] != "clear" && args[2] != "create_db" && args[2] != "force_clear" {
			global.GDb = initialize.Gorm()
		}
		switch args[2] {
		case "test":
			Test(args)
		case "import":
			if len(args) < 7 || (args[3] != "-f" && args[3] != "--file") || (args[5] != "-l" && args[5] != "--log") {
				color.Error.Println("Parameter error, \n Format is '-m import -f [file] -l [log_file_path]'")
				return
			}
			ImportMetaData(args[4], args[6])
		case "export":
			if len(args) < 7 || (args[3] != "-f" && args[3] != "--file") || (args[5] != "-l" && args[5] != "--log") {
				color.Error.Println("Parameter error, \n Format is '-m export -f [file] -l [log_file_path]'")
				return
			}
			ExportMetaData(args[4], args[6])
		case "clear":
			ClearData(false)
		case "force_clear":
			// 该命令不对外
			ClearData(true)
		case "movecontent":
			if len(args) < 5 || (args[3] != "-f" && args[3] != "--file") {
				color.Error.Println("parameter error")
				return
			}
			MoveContent(args[4])
		case "create_db":
			// 该命令不对外
			CreateDatabase(false)
		case "init_db":
			// 该命令不对外
			InitDatabase()
		case "init_data":
			// 该命令不对外
			initialize.Redis()
			InitData()
		case "create_user":
			if len(args) < 7 || (args[3] != "-u" && args[3] != "--user") || (args[5] != "-w" && args[5] != "--pwd") {
				color.Error.Println("parameter error")
				return
			}
			CreateUser(args[4], args[6])
		case "reset_pwd":
			if len(args) < 7 || (args[3] != "-u" && args[3] != "--user") || (args[5] != "-w" && args[5] != "--pwd") {
				color.Error.Println("parameter error")
				return
			}
			ChangePwd(args[4], args[6])
		case "delete_user":
			if len(args) < 5 || (args[3] != "-u" && args[3] != "--user") {
				color.Error.Println("parameter error")
				return
			}
			DelUser(args[4])
		case "all_user":
			AllUser()
		case "clean_all_computer":
			// 该命令不对外
			CleanAllComputer()
		case "flush_redis":
			FlushRedis()
		case "reset_import":
			ResetImport()
		case "import_procedure":
			// 该命令不对外
			ImportProcedure()
		case "refresh_isonserver":
			// 该命令不对外
			RefreshIsonserver()
		case "mysql_character_set":
			// 该命令不对外
			SetMySQLCharacter()
		default:
			color.Error.Println("parameter error")
		}
	} else {
		color.Error.Println("parameter error")
	}
}
