package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"gofish/game/common"
	"gofish/game/conf"
	_ "gofish/game/router" //直接一手触发init
	"gofish/game/service"
	"gofish/model"
	"net/http"
)

func main() {

	logs.Debug("开始运行 \n")
	//定义初始化
	err := conf.InitConf()
	if err != nil {
		logs.Error("init conf err: %v", err)
		return
	}
	logs.Debug("读取conf.conf配置完成 \n")
	err = conf.InitSec()
	if err != nil {
		logs.Error("初始化initSec错误： %v", err)
		return
	}
	logs.Debug("InitSec 完成 \n")
	model.InitDb()
	logs.Debug("InitDb 完成 \n")

	addr := fmt.Sprintf("%s:%d", common.GameConf.Host, common.GameConf.Port)
	fmt.Println("地址", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		logs.Error("监听错误: %v", err)
	}
	service.HandleHallBroadcast()
	select {}
}
