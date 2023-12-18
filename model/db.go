package model

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"gofish/game/gcommon"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var db *sqlx.DB

// init 函数会在包初始化时调用
func InitDb() {
	// 初始化数据库连接
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", gcommon.GameConf.MysqlUser, gcommon.GameConf.MysqlPassword, gcommon.GameConf.MysqlAddr, gcommon.GameConf.MysqlDb)
	database, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		fmt.Printf("Failed to connect to the database: %v\n", err)
		return
	}
	db = database

	initRedisGameCfg()
	// 设置信号处理，确保在程序退出时关闭数据库连接
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		CloseDB() // 关闭数据库连接
		CloseRedis()
		os.Exit(1)
	}()
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("mysql 关闭错误 %v\n", err)
		}
	}
}
