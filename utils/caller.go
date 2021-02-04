package utils

import (
	"bytes"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"goblog/core/global"
	reflect "reflect"
	"strconv"
	"strings"
)

// 存储过程调用器
func CallProc(funcName string, Params ...interface{}) (error, map[string]string) {
	m := global.GConfig.Mysql
	dsn := m.Username + ":" + m.Password + "@tcp(" + m.Path + ")/" + m.Dbname + "?" + m.Config
	db, err := sql.Open("mysql", dsn+"&multiStatements=true")

	if err != nil {
		return err, nil
	}
	defer db.Close()
	buf := new(bytes.Buffer)
	buf.WriteString("call ")
	buf.WriteString(funcName)
	buf.WriteString("(")
	var allParams []string
	var outParams []string
	for _, p := range Params {
		if reflect.TypeOf(p).Kind() == reflect.String {
			if strings.HasPrefix(p.(string), "@") {
				outParams = append(outParams, p.(string))
				allParams = append(allParams, p.(string))
			} else {
				allParams = append(allParams, "'"+p.(string)+"'")
			}
		} else if Contain(reflect.TypeOf(p).Kind(), []reflect.Kind{reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64}) {
			allParams = append(allParams, strconv.Itoa(ToInt(p)))
		} else if reflect.TypeOf(p).Kind() == reflect.Float32 {
			allParams = append(allParams, strconv.FormatFloat(float64(p.(float32)), 'E', -1, 32))
		} else if reflect.TypeOf(p).Kind() == reflect.Float32 {
			allParams = append(allParams, strconv.FormatFloat(p.(float64), 'E', -1, 64))
		}
	}
	if len(allParams) > 0 {
		buf.WriteString(strings.Join(allParams, ","))
	}
	buf.WriteString(");")
	if len(outParams) > 0 {
		buf.WriteString("select ")
		buf.WriteString(strings.Join(outParams, ","))
		buf.WriteString(";")
	}
	rows, err := db.Query(buf.String())
	if err != nil {
		return err, nil
	}
	var res = make(map[string]string)
	defer rows.Close()
	if rows.Next() {
		rowColumns, err := rows.Columns()
		if len(rowColumns) > 0 && err == nil {
			// 以interface切片为载具，承载数量为SP OUT类型参数个数的string指针，方便Scan进结果集
			var result []interface{}
			for i := 0; i < len(rowColumns); i++ {
				var res string
				result = append(result, &res)
			}
			rows.Scan(result...)
			for i, r := range result {
				res[strings.Replace(rowColumns[i], "@", "", 1)] = *r.(*string)
			}
		}
	}
	return nil, res
}
