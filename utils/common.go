package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/snappy"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// If 三元表达式
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

type QueryFilter struct {
	Name string      `form:"Name" json:"Name"`
	Op   string      `form:"Op" json:"Op"`
	Val  interface{} `form:"Val" json:"Val"`
}

var SignMap = map[string]string{"eq": "=", "ne": "!=", "in_": "in", "notin_": "not in", "ilike": "like", "like": "like", "ge": ">=", "le": "<="}

func (q QueryFilter) Handle() string {
	val := q.Val
	name := q.Name
	if q.Op == "like" {
		val = "%" + q.Val.(string) + "%"
	} else if q.Op == "like" {
		val = "'%" + q.Val.(string) + "%'"
	} else if q.Op == "ilike" {
		val = "'%" + strings.ToUpper(q.Val.(string)) + "%'"
		name = "upper(" + q.Name + ")"
	} else if q.Op == "in_" || q.Op == "notin_" {
		if reflect.TypeOf(q.Val).Kind() == reflect.Slice {
			var valSlice []string
			for i := 0; i < reflect.ValueOf(q.Val).Len(); i++ {
				singleVal := reflect.ValueOf(q.Val).Index(i).Interface()
				if reflect.TypeOf(singleVal).Kind() == reflect.Float64 {
					valSlice = append(valSlice, strconv.Itoa(int(singleVal.(float64))))
				} else if reflect.TypeOf(singleVal).Kind() == reflect.String {
					valSlice = append(valSlice, "'"+singleVal.(string)+"'")
				}
			}
			val = "(" + strings.Join(valSlice, ",") + ")"
			return name + " " + SignMap[q.Op] + " " + val.(string) + " "
		}
	}
	return name + " " + SignMap[q.Op] + " '" + val.(string) + "' "
}

// RemoveRepByLoop 通过两重循环过滤重复元素
func RemoveRepByLoop(slc []int) []int {
	var result []int // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

// RemoveRepByMap 通过map主键唯一的特性过滤重复元素
func RemoveRepByMap(slc []int) []int {
	var result []int
	tempMap := map[int]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// RemoveRep 元素去重
func RemoveRep(slc []int) []int {
	if len(slc) < 1024 {
		// 切片长度小于1024的时候，循环来过滤
		return RemoveRepByLoop(slc)
	} else {
		// 大于的时候，通过map来过滤
		return RemoveRepByMap(slc)
	}
}

// Maximum 计算最大值
func Maximum(l []int) (max int) {
	max = l[0]
	for _, v := range l {
		if v > max {
			max = v
		}
	}
	return
}

// Minimum 计算最小值
func Minimum(l []int) (min int) {
	min = l[0]
	for _, v := range l {
		if v < min {
			min = v
		}
	}
	return
}

// XMLEscape xml转义
func XMLEscape(xmlStr string) string {
	return strings.Replace(strings.Replace(xmlStr, "<", "&lt;", -1), ">", "&gt;", -1)
}

func XMLUnEscape(xmlStr string) string {
	return strings.Replace(strings.Replace(xmlStr, "&lt;", "<", -1), "&gt;", ">", -1)
}

// Contain 判断array，Slice或者Map内是否包含某元素
func Contain(obj interface{}, array interface{}) bool {
	targetValue := reflect.ValueOf(array)
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

// ContainStr 判断字符串切片内是否包含某元素
func ContainStr(val string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == val {
			return true
		}
	}
	return false
}

// ContainInt 判断整型切片内是否包含某元素
func ContainInt(val int, slice []int) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == val {
			return true
		}
	}
	return false
}

// StringBuilder 字符串拼接函数
func StringBuilder(p []string) string {
	var b strings.Builder
	l := len(p)
	for i := 0; i < l; i++ {
		b.WriteString(p[i])
	}
	return b.String()
}

// AnchorToDatetime anchor转化为时间
func AnchorToDatetime(anchor string) (lastChangeNumber int, anchorTime time.Time, err error) {
	paramList := strings.Split(anchor, ",")
	lastChangeNumber, err = strconv.Atoi(paramList[0])
	anchorTime, err = time.Parse("2006-01-02 15:04:05", paramList[1])
	return
}

// DatetimeToAnchor 时间转anchor
func DatetimeToAnchor(dt time.Time, lastChangeNumber int) (anchor string) {
	dtStr := dt.Format("2006-01-02 15:04:05.000")
	numberStr := fmt.Sprintf("%d,", lastChangeNumber)
	return StringBuilder([]string{numberStr, dtStr})
}

// SliceDistinct 跨类型切片去重函数
func SliceDistinct(slc interface{}) (ret []interface{}, err error) {
	if reflect.TypeOf(slc).Kind() != reflect.Slice {
		return ret, errors.New("no slice type data used in Func SliceDistinct")
	}
	rValues := reflect.ValueOf(slc)
	tempMap := make(map[interface{}]bool)
	for i := 0; i < rValues.Len(); i++ {
		value := rValues.Index(i).Interface()
		if !tempMap[value] {
			ret = append(ret, rValues.Index(i).Interface())
			tempMap[value] = true
		}
	}
	return ret, nil
}

// DelIntItem 删除切片中的整型元素
func DelIntItem(vs []int, s int) []int {
	for i := 0; i < len(vs); i++ {
		if s == vs[i] {
			vs = append(vs[:i], vs[i+1:]...)
			i = i - 1
		}
	}
	return vs
}

// DelStringItem 删除切片中的字符串元素
func DelStringItem(vs []string, s string) []string {
	for i := 0; i < len(vs); i++ {
		if s == vs[i] {
			vs = append(vs[:i], vs[i+1:]...)
			i = i - 1
		}
	}
	return vs
}

// StructMakeUp 多个结构体或者Map组合成为一个字典
func StructMakeUp(args ...interface{}) map[string]interface{} {
	var data = make(map[string]interface{})
	for _, arg := range args {
		if reflect.TypeOf(arg).Kind() == reflect.Struct {
			for i := 0; i < reflect.ValueOf(arg).NumField(); i++ {
				data[reflect.TypeOf(arg).Field(i).Name] = reflect.ValueOf(arg).Field(i).Interface()
			}
		} else if reflect.TypeOf(arg).Kind() == reflect.Map {
			for k, v := range arg.(map[string]interface{}) {
				data[k] = v
			}
		}
	}
	return data
}

// GetPCName 获取当前服务器的名称
func GetPCName() string {
	pcName, _ := os.Hostname()
	return pcName
}

// GetFuncName 获取方法名(异步函数callback时使用)
func GetFuncName(f interface{}) string {
	funcPath := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	funcModules := strings.Split(funcPath, "/")
	funcName := funcModules[len(funcModules)-1]
	return funcName
}

// ToInt number类型的interface转换为标准int
func ToInt(i interface{}) int {
	if i == nil {
		return 0
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Int64:
		return int(i.(int64))
	case reflect.Int32:
		return int(i.(int32))
	case reflect.Int16:
		return int(i.(int16))
	case reflect.Int8:
		return int(i.(int8))
	case reflect.Float64:
		return int(i.(float64))
	case reflect.Float32:
		return int(i.(float32))
	default:
		return i.(int)
	}
}

// SliceIntToString int切片转换为string切片。大量的sql in操作时需要用到
func SliceIntToString(src []int) (drc []string) {
	for _, i := range src {
		drc = append(drc, strconv.Itoa(i))
	}
	return
}

// IntToBool int --> bool
func IntToBool(param int) bool {
	res := false
	if param > 0 {
		res = true
	}
	return res
}

func GetUnEscapeJsonEncoder() (*json.Encoder, *bytes.Buffer) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	return jsonEncoder, bf
}

// IntMapStr int silce >> string slice
func IntMapStr(intSlice []int) []string {
	var strSlice []string
	for _, i := range intSlice {
		strSlice = append(strSlice, strconv.Itoa(i))
	}
	return strSlice
}

// StrMapInt string silce >> int slice
func StrMapInt(intSlice []string) []int {
	var strSlice []int
	for _, i := range intSlice {
		val, _ := strconv.Atoi(i)
		strSlice = append(strSlice, val)
	}
	return strSlice
}

// MD5V md5加密
func MD5V(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

// SnappyEncode snappy
func SnappyEncode(data string) string {
	return base64.StdEncoding.EncodeToString(snappy.Encode(nil, []byte(data)))
}

func SnappyDecode(data interface{}) string {
	decodeString, _ := base64.StdEncoding.DecodeString(reflect.ValueOf(data).String())
	result, _ := snappy.Decode(nil, decodeString)
	return string(result)
}

// GetPageParams 获取分页信息
func GetPageParams(c *gin.Context) (pageNum, pageSize int) {
	pageNum, _ = strconv.Atoi(c.DefaultQuery("PageNumber", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("PageSize", "5"))
	return
}

// GetUserInfo 获取登录用户信息
func GetUserInfo(c *gin.Context) (userID int, userEmail string) {
	userID = c.GetInt("UID")
	userEmail = c.GetString("Email")
	return
}

// GenSqlStr 生成数据库查询可以使用的sql格式
func GenSqlStr(param []interface{}) string {
	/*
		[]int{1,2,3} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		if val, ok := param[i].(int); ok {
			paramList = append(paramList, "'"+strconv.Itoa(val)+"'")
		} else if val, ok := param[i].(string); ok {
			paramList = append(paramList, "'"+val+"'")
		}
	}
	return strings.Join(paramList, ",")
}

func GenSqlStrInt(param []int) string {
	/*
		[]int{1,2,3} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		paramList = append(paramList, "'"+strconv.Itoa(param[i])+"'")
	}
	return strings.Join(paramList, ",")
}

func GenSqlStrString(param []string) string {
	/*
		[]string{"1","2","3"} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		paramList = append(paramList, "'"+param[i]+"'")
	}
	return strings.Join(paramList, ",")
}

// Interface2String 报表导出使用，将interface全部转为string
func Interface2String(param interface{}) string {
	if param == nil {
		return ""
	}
	switch param.(type) {
	case string:
		return param.(string)
	case int:
		return strconv.Itoa(param.(int))
	case time.Time:
		return param.(time.Time).Add(time.Hour * 8).Format("2006-01-02 15:04:05")
	case float64:
		return strconv.Itoa(int(param.(float64)))
	case int8:
		return strconv.Itoa(int(param.(int8)))
	case int32:
		return strconv.Itoa(int(param.(int32)))
	case int64:
		return strconv.Itoa(int(param.(int64)))
	default:
		return ""
	}
}

// MinutesToTime 分钟数转化为时分秒
func MinutesToTime(minutes int) string {
	if minutes == 0 {
		return ""
	}
	seconds := minutes * 60

	m := seconds / 60
	s := seconds % 60

	h := m / 60
	m = m % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// Shuffle 随机打乱顺序
func Shuffle(slice []map[string]interface{}) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(slice) > 0 {
		n := len(slice)
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
		slice = slice[:n-1]
	}
}
