package conf

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"gofish/game/gcommon"
)

func InitConf() (err error) {
	conf, err := config.NewConfig("ini", "./common/conf/conf.conf")
	if err != nil {
		fmt.Println("没有配置文件:", err)
		return
	}
	gcommon.GameConf.Host = conf.String("host")
	if gcommon.GameConf.Host == "" {
		return fmt.Errorf("监听地址为空")
	}

	gcommon.GameConf.Port, err = conf.Int("port")
	if err != nil {
		fmt.Println("没有配置文件:", err)
		return
	}
	gcommon.GameConf.LogPath = conf.String("log_path")
	if gcommon.GameConf.LogPath == "" {
		return fmt.Errorf("log路径为空")
	}
	gcommon.GameConf.LogLevel = conf.String("log_level")
	if gcommon.GameConf.LogLevel == "" {
		return fmt.Errorf("日志等级没配置")
	}
	gcommon.GameConf.MysqlAddr = conf.String("mysql_addr")
	if gcommon.GameConf.MysqlAddr == "" {
		return fmt.Errorf("mysql_addr为空")
	}
	gcommon.GameConf.MysqlUser = conf.String("mysql_user")
	if gcommon.GameConf.MysqlUser == "" {
		return fmt.Errorf("mysql_user为空")
	}
	gcommon.GameConf.MysqlDb = conf.String("mysql_db")
	if gcommon.GameConf.MysqlDb == "" {
		return fmt.Errorf("mysql_db为空")
	}
	gcommon.GameConf.MysqlPassword = conf.String("mysql_password")
	if gcommon.GameConf.MysqlPassword == "" {
		return fmt.Errorf("mysql_password账户为空")
	}

	gcommon.GameConf.RedisGameCfgAddr = conf.String("redis_game_cfg_addr")
	if gcommon.GameConf.MysqlPassword == "" {
		return fmt.Errorf("redis_game_cfg_addr账户为空")
	}
	gcommon.GameConf.RedisGameCfgPass = conf.String("redis_game_cfg_pass")
	if gcommon.GameConf.MysqlPassword == "" {
		return fmt.Errorf("redis_game_cfg_pass账户为空")
	}

	return
}
