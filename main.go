package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gofish/game/common"
	"gofish/game/conf"
	_ "gofish/game/router" //直接一手触发init
	"gofish/game/service"
	"gofish/model"
	"net/http"
)

func main() {
	fmt.Println("开始运行")
	//定义初始化
	err := conf.InitConf()
	if err != nil {
		logs.Error("init conf err: %v", err)
		return
	}
	logs.Debug("读取conf.conf配置完成")

	err = conf.InitConf()
	if err != nil {
		logs.Error("初始化initSec错误： %v", err)
		return
	}
	logs.Debug("InitSec 完成")

	addr := fmt.Sprintf("%s:%d", common.GameConf.Host, common.GameConf.Port)
	fmt.Println("地址", addr)
	//格式化字符
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", common.GameConf.MysqlUser, common.GameConf.MysqlPassword, common.GameConf.MysqlAddr, common.GameConf.MysqlDb)
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		fmt.Printf("Failed to connect to the database: %v\n", err)
		return
	}

	defer db.Close()

	// 初始化全局数据库连接
	model.InitDB(db)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logs.Error("监听错误: %v", err)
	}
	service.HandleHallBroadcast()
	select {}
}
