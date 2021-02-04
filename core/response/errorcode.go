package response

const (
	COMMON           = 9999
	SUCCESS          = COMMON + iota // 成功
	UNKNOWN                          // 未知错误
	FAILED                           // 失败
	ServerTooBusy                    // 服务器繁忙  # 限流
	RequestTooOften                  // 请求太频繁
	PermissionDenied                 // 权限不足
	ParamNOTEnough                   // 参数不足
	ParamError                       // 参数错误
	NotFound                         // 未找到
	NotLogin                         // 未登录
	MethodNotAllowed                 // 方法不被允许
	FormatError                      // 格式错误
	NotFinish                        // 未完成
	IDNotFound                       // 当前ID未找到
	CleanFailed                      // 清理失败
	Invalid                          // 失效
)

var ErrorMsg = map[int]string{
	SUCCESS:          "成功",
	UNKNOWN:          "未知错误",
	FAILED:           "失败",
	ServerTooBusy:    "服务器繁忙",
	RequestTooOften:  "请求太频繁",
	PermissionDenied: "权限不足",
	ParamNOTEnough:   "参数不足",
	ParamError:       "参数错误",
	NotFound:         "未找到",
	NotLogin:         "未登录",
	MethodNotAllowed: "方法不被允许",
	FormatError:      "格式错误",
	NotFinish:        "未完成",
	IDNotFound:       "当前ID未找到",
	CleanFailed:      "清理失败",
	Invalid:          "失效",
}
