package service

import (
	"fmt"
	"go.uber.org/zap"
	"goblog/core/global"
)

// CacheSetter 缓存设置再封装。当设置失败时，即新Cache没有成功更新进redis，旧Cache需要保证清理干净
func CacheSetter(method, cacheKey string, args ...interface{}) {
	var params []interface{}
	var setErr error
	params = append(params, method)
	params = append(params, cacheKey)
	params = append(params, args...)
	if method == "HMSet" {
		if err := global.GRedis.HMSet(cacheKey, args[0].(map[string]interface{})).Err(); err != nil {
			setErr = err
		}
	} else {
		if err := global.GRedis.Do(params...).Err(); err != nil {
			setErr = err
		}
	}
	if setErr != nil {
		global.GLog.Error(fmt.Sprintf("Set redis cache [%s] methd [%s] error,", cacheKey, method), zap.Any("err", setErr))
		global.GRedis.Del(cacheKey)
	}
}
