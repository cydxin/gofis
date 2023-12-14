package model

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis/v8"
	"gofish/game/common"
)

var redisGameCfg *redis.Client

func initRedisGameCfg() {
	redisGameCfg = redis.NewClient(&redis.Options{
		Addr:     common.GameConf.RedisGameCfgAddr,
		Password: common.GameConf.RedisGameCfgPass,
		DB:       0,
	})
	GameCfg, err := GetPkRoom()
	if err != nil {
		logs.Error("redis GameCfg初始化异常 err：%v", err)
	}
	for i, info := range GameCfg {
		fmt.Printf("%v%v", i, info)
	}
	logs.Debug("initRedisGameCfg 完成 \n")
}
func CloseRedis() {
	err := redisGameCfg.Close()
	if err != nil {
		return
	}
}
