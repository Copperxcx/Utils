package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql驱动
	"github.com/jmoiron/sqlx"          // sqlx包
)

// 操作类型，当前可取值：config add get
type Operation struct {
	Op string `json:"op"`
}

// 配置服务参数，如mysql服务器ip和端口，登录用户名和密码，使用的数据库等
type DbConfig struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	UserName string `json:"username"`
	Pwd      string `json:"pwd"`
	DbName   string `json:"dbname"`
	initOk   bool
}

func (s DbConfig) IsValid() error {
	var res string
	if s.DbName == "" {
		res += "数据库名为空！"
	}
	if s.Ip == "" {
		// TODO: 正则表达式判断为IP地址
		res += "IP地址错误！"
	}
	if s.Port == "" {
		res += "端口为空！"
	}
	if s.UserName == "" || s.Pwd == "" {
		res += "用户名或密码为空！"
	}

	if res == "" {
		return nil
	}
	return errors.New(res)
}

// 从配置对象中获取打开sql的语句
func (s DbConfig) getOpenSqlStr() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", s.UserName, s.Pwd, s.Ip, s.Port, s.DbName)
}

// 添加、更新记录结构体
type AddConfigItems struct {
	Op      string     `json:"op"`
	Records [][]string `json:"records"`
}
type UpdateConfigItems AddConfigItems
type AddLogItems AddConfigItems

// 查询、删除记录结构体
type GetConfigItems struct {
	Op      string   `json:"op"`
	Records []string `json:"records"`
}
type DelConfigItems GetConfigItems

// 配置条目
type ConfigItem struct {
	Name  string `db:"name" json:"name"`
	Value string `db:"value" json:"value"`
	Memo  string `db:"memo" json:"memo,omitempty"`
}

// 请求回复
type Reply struct {
	Op      string       `json:"op,omitempty"`
	Result  string       `json:"result"`
	Message string       `json:"message,omitempty"`
	Records []ConfigItem `json:"records,omitempty"`
}

var database *sqlx.DB
var config DbConfig

func main() {
	config.Ip = "127.0.0.1"
	config.Port = "3306"
	config.UserName = "root"
	config.Pwd = "root"
	config.DbName = "test"

	config.initOk = false

	// TODO: IP和端口做成可修改
	http.HandleFunc("/", rootHandler)
	err := http.ListenAndServe("127.0.0.1:9000", nil)
	if err != nil {
		fmt.Println("数据库服务启动失败: ", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	// fmt.Println(r.RequestURI)

	// 回复结构体，将编码为json后回复
	var rep Reply

	if r.Method != "POST" {
		// fmt.Println(r.Method)

		replyError(w, &rep, "请使用POST方式使用数据库服务！")
		return
	}

	b, _ := ioutil.ReadAll(r.Body)

	if len(b) == 0 {
		replyError(w, &rep, "请发送正确POST请求参数！")
		return
	}

	var op Operation
	err := json.Unmarshal(b, &op)

	if err != nil {
		replyError(w, &rep, "JSON解析错误: "+err.Error())
		return
	}

	var result string
	var items []ConfigItem
	switch strings.ToLower(op.Op) {
	// 设置数据库连接参数
	case "config":
		result = configDb(b)
	// 增删改查配置
	case "add_config":
		result = addConfig(b)
	case "update_config":
		result = updateConfig(b)
	case "get_config":
		items, result = getConfig(b)
	case "del_config":
		result = delConfig(b)
	case "clear_config":
		result = clearConfig(b)
	case "add_log":
		result = addLog(b)
	default:
		result = "传入'op'参数有误，不能为'" + op.Op + "'"
	}

	// 将处理结果组织为json格式回复请求
	if result == "OK" {
		if op.Op == "get_config" {
			rep.Records = items
		}
		rep.Op = op.Op
		rep.Result = result
		b, _ := json.Marshal(rep)
		w.Write(b)
	} else {
		replyError(w, &rep, result)
	}
}

// 判断是否已经连接数据库
func initValidate() string {
	if config.initOk {
		return "OK"
	}
	// err := Init(&config)
	// if err != nil {
	// 	str := fmt.Sprintf("数据库连接失败: %v", err)
	// 	return str
	// }
	return "数据库未连接，请使用'config'操作初始化数据库连接！"
}

// 添加记录的处理
func addConfig(body []byte) string {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return s
	}

	var st AddConfigItems
	err := json.Unmarshal(body, &st)
	if err != nil {
		return "JSON解析错误: " + err.Error()
	}
	// 是否有数据
	if len(st.Records) == 0 {
		return "未包含数据！"
	}

	// 插入每条数据
	for _, value := range st.Records {
		// 每条数据都应包含 key 和 value，可以再包含 memo
		var err error
		switch len(value) {
		case 2:
			_, err = database.Exec("INSERT INTO tbl_config( name, value ) VALUES(?, ?);", value[0], value[1])
		case 3:
			_, err = database.Exec("INSERT INTO tbl_config  VALUES(?, ?, ?);", value[0], value[1], value[2])
			// 数据字段不正确，直接返回错误
		default:
			return "数据字段数不正确，必须为2或3个字段！"
		}
		if err != nil {
			return err.Error()
		}
	}

	return "OK"
}

// 更新配置
func updateConfig(body []byte) string {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return s
	}

	var st UpdateConfigItems
	err := json.Unmarshal(body, &st)
	if err != nil {
		return "JSON解析错误: " + err.Error()
	}
	// 是否有数据
	if len(st.Records) == 0 {
		return "未包含数据！"
	}

	// 更新每条数据
	for _, value := range st.Records {
		// 每条数据都应包含 key 和 value，可以再包含 memo, newKey
		var err error
		var key, newKey, newValue, newMemo string

		switch len(value) {
		case 2:
			key = value[0]
			newValue = value[1]
			newMemo = ""
			newKey = value[0]

		case 3:
			key = value[0]
			newValue = value[1]
			newMemo = value[2]
			newKey = value[0]

		case 4:
			key = value[0]
			newValue = value[1]
			newMemo = value[2]
			newKey = value[3]
			// 数据字段不正确，直接返回错误
		default:
			return "数据字段数不正确，必须为2或3个字段！"
		}

		_, err = database.Exec("UPDATE tbl_config SET name=?, value=?, memo=? WHERE name=?;", key, newValue, newMemo, newKey)
		if err != nil {
			return err.Error()
		}
	}

	return "OK"
}

// 删除配置
func delConfig(body []byte) string {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return s
	}

	var st DelConfigItems
	err := json.Unmarshal(body, &st)
	if err != nil {
		return "JSON解析错误: " + err.Error()
	}
	// 是否有数据
	if len(st.Records) == 0 {
		return "未包含数据！"
	}

	// 删除每条数据
	for _, key := range st.Records {
		// 删除name字段为key对应的记录
		_, err = database.Exec("DELETE FROM tbl_config WHERE name=?;", key)

		if err != nil {
			return err.Error()
		}
	}

	return "OK"
}

// 删除配置
func clearConfig(body []byte) string {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return s
	}

	_, err := database.Exec("DELETE FROM tbl_config;")
	if err != nil {
		return err.Error()
	}
	return "OK"
}

// 查询配置
func getConfig(body []byte) ([]ConfigItem, string) {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return nil, s
	}

	var st GetConfigItems
	err := json.Unmarshal(body, &st)
	if err != nil {
		return nil, "JSON解析错误: " + err.Error()
	}
	// 是否有数据
	if len(st.Records) == 0 {
		// 获取所有条目
		var items []ConfigItem
		err = database.Select(&items, "SELECT * FROM tbl_config;")
		if err != nil {
			return nil, err.Error()
		}

		return items, "OK"
	}

	// 获取指定key对应的数据
	var items []ConfigItem
	for _, key := range st.Records {
		// 删除name字段为key对应的记录
		var item []ConfigItem
		err := database.Select(&item, "SELECT * FROM tbl_config WHERE name=?;", key)

		if err != nil {

			return nil, err.Error()
		} else {
			// 只能查找到1条记录，其余都不正确
			if len(item) != 1 {
				// 未获取到的值的处理
				var i ConfigItem
				i.Name = "***" + key + "***"
				i.Value = "***Key dosn't exists.***"
				items = append(items, i)
			} else {
				items = append(items, item[0])
			}
		}
	}

	return items, "OK"
}

// 回复请求错误
func replyError(w http.ResponseWriter, rep *Reply, message string) {
	rep.Result = "ERROR"
	rep.Message = message
	b, _ := json.Marshal(rep)
	w.Write(b)
}

// 初始化数据库连接
func Init(conf *DbConfig) error {

	str := conf.getOpenSqlStr()

	// fmt.Println(str)

	xdb, err := sqlx.Open("mysql", str)
	if err != nil {
		return err
	}

	err = xdb.Ping()
	if err != nil {
		return err
	}

	// 成功打开，赋值变量
	database = xdb
	config = *conf
	config.initOk = true

	return nil
}

// 设置数据库连接参数
func configDb(body []byte) string {

	var conf DbConfig
	err := json.Unmarshal(body, &conf)
	if err != nil {
		return "JSON数据解析错误！ " + err.Error()
	}

	// config.initOk = false
	// database = nil

	// 对未提供的参数，使用原先参数
	// IP地址
	if conf.Ip == "" {
		conf.Ip = config.Ip
	}
	// 端口
	if conf.Port == "" {
		conf.Port = config.Port
	}
	// 用户名
	if conf.UserName == "" {
		conf.UserName = config.UserName
	}
	// 密码
	if conf.Pwd == "" {
		conf.Pwd = config.Pwd
	}
	// 数据库名称
	if conf.DbName == "" {
		conf.DbName = config.DbName
	}

	// 如果不是以db_project_开头，则加上前缀
	if strings.Index(conf.DbName, "db_project_") != 0 {
		conf.DbName = "db_project_" + conf.DbName
	}

	err = Init(&conf)
	if err != nil {
		return err.Error()
	}
	// fmt.Println(database)
	// fmt.Println("OKOKOK")
	return "OK"
	// if database.IsValid() {
	// 	return "OK"
	// } else {
	// 	config.initOk = false
	// 	database = nil
	// 	return "数据库不存在！"
	// }
}

// 添加日志条目
func addLog(body []byte) string {
	// 判断是否已经连接数据库
	s := initValidate()
	if s != "OK" {
		return s
	}

	var st AddLogItems
	err := json.Unmarshal(body, &st)
	if err != nil {
		return "JSON解析错误: " + err.Error()
	}
	// 是否有数据
	if len(st.Records) == 0 {
		return "未包含数据！"
	}

	// 插入每条数据
	for _, value := range st.Records {
		// 每条数据都应包含 type detail，也可再包含 ts，ts必须包含3位小数
		var err error
		switch len(value) {
		case 2:
			value[0] = strings.ToUpper(value[0])
			_, err = database.Exec("INSERT INTO tbl_log( type, detail ) VALUES(?, ?);", value[0], value[1])
		case 3:
			value[0] = strings.ToUpper(value[0])
			_, err = database.Exec("INSERT INTO tbl_log( type, detail, ts)  VALUES(?, ?, ?);", value[0], value[1], value[2])
			// 数据字段不正确，直接返回错误
		default:
			return "数据字段数不正确，必须为2或3个字段！"
		}
		if err != nil {
			return err.Error()
		}
	}

	return "OK"
}
