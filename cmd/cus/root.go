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
	global.GVP = core.Viper("config.ini")
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
		case "clear":
			ClearData(false)
		case "force_clear":
			// 该命令不对外
			ClearData(true)
		case "create_db":
			// 该命令不对外
			CreateDatabase(false)
		case "init_db":
			// 该命令不对外
			InitDatabase()
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
