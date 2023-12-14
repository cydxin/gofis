package conf

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"gofish/game/common"
)

func InitConf() (err error) {
	conf, err := config.NewConfig("ini", "./common/conf/conf.conf")
	if err != nil {
		fmt.Println("没有配置文件:", err)
		return
	}
	common.GameConf.Host = conf.String("host")
	if common.GameConf.Host == "" {
		return fmt.Errorf("监听地址为空")
	}
	common.GameConf.Port, err = conf.Int("port")
	if err != nil {
		fmt.Println("没有配置文件:", err)
		return
	}
	common.GameConf.LogPath = conf.String("log_path")
	if common.GameConf.LogPath == "" {
		return fmt.Errorf("log路径为空")
	}
	common.GameConf.LogLevel = conf.String("log_level")
	if common.GameConf.LogLevel == "" {
		return fmt.Errorf("日志等级没配置")
	}
	common.GameConf.MysqlAddr = conf.String("mysql_addr")
	if common.GameConf.MysqlAddr == "" {
		return fmt.Errorf("mysql_addr为空")
	}
	common.GameConf.MysqlUser = conf.String("mysql_user")
	if common.GameConf.MysqlUser == "" {
		return fmt.Errorf("mysql_user为空")
	}
	common.GameConf.MysqlDb = conf.String("mysql_db")
	if common.GameConf.MysqlDb == "" {
		return fmt.Errorf("mysql_db为空")
	}
	common.GameConf.MysqlPassword = conf.String("mysql_password")
	if common.GameConf.MysqlPassword == "" {
		return fmt.Errorf("mysql_password账户为空")
	}

	common.GameConf.RedisGameCfgAddr = conf.String("redis_game_cfg_addr")
	if common.GameConf.MysqlPassword == "" {
		return fmt.Errorf("redis_game_cfg_addr账户为空")
	}
	common.GameConf.RedisGameCfgPass = conf.String("redis_game_cfg_pass")
	if common.GameConf.MysqlPassword == "" {
		return fmt.Errorf("redis_game_cfg_pass账户为空")
	}

	return
}
