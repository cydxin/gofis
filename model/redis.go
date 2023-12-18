package model

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/go-redis/redis/v8"
	"gofish/common"
	"gofish/game/gcommon"
	"log"
)

var redisGameCfg *redis.Client

func initRedisGameCfg() {
	redisGameCfg = redis.NewClient(&redis.Options{
		Addr:     gcommon.GameConf.RedisGameCfgAddr,
		Password: gcommon.GameConf.RedisGameCfgPass,
		DB:       0,
	})
	GameCfg, err := GetPkRoom()
	if err != nil {
		logs.Error("redis GameCfg初始化异常 err：%v", err)
	}
	for _, info := range GameCfg {
		err := setPkConfigToRedis(info)
		if err != nil {
			logs.Debug("err: %v", err)
			return
		}
	}

	GameCfg = nil

	logs.Debug("initRedisGameCfg 完成 \n")
}

// 将PK场次配置信息存储到 Redis 中
func setPkConfigToRedis(info *common.PkRoomInfo) error {
	//logs.Debug("info: %v", info)

	key := fmt.Sprintf("%d&{%s}", info.PkNumOfPeople, info.RoomName)
	infoJson, err := json.Marshal(info)
	if err != nil {
		logs.Debug("存储配置信息到 Redis 失败：%v\n", err)
		return err
	}
	//logs.Debug("infoJson: %v", string(infoJson))

	err = redisGameCfg.Set(redisGameCfg.Context(), key, infoJson, 0).Err()
	if err != nil {
		logs.Debug("存储配置信息到 Redis 失败：%v\n", err)
		return err
	}
	return nil
}

// GetConfigFromRedis 从 Redis 中获取配置信息
func GetConfigFromRedis(pkNumOfPeo int, roomName string) (*common.PkRoomInfo, error) {
	key := fmt.Sprintf("%d&{%s}", pkNumOfPeo, roomName)
	logs.Debug("GetConfigFromRedis的key:%v", key)
	val, err := redisGameCfg.Get(redisGameCfg.Context(), key).Bytes()
	if err != nil {
		log.Printf("从 Redis 获取配置key:%v 信息失败：%v \n", key, err)
		return nil, err
	}
	logs.Debug("读取的值，redisGameCfg：%v", string(val))

	var config common.PkRoomInfo
	err = json.Unmarshal(val, &config)
	if err != nil {
		log.Printf("解析配置信息失败：%v\n", err)
		return nil, err
	}

	return &config, nil
}

func CloseRedis() {
	err := redisGameCfg.Close()
	if err != nil {
		return
	}
}
