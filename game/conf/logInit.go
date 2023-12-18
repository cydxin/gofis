package conf

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"gofish/game/gcommon"
)

func conversionLogLevel(logLevel string) int {
	switch logLevel {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = gcommon.GameConf.LogPath
	config["level"] = conversionLogLevel(gcommon.GameConf.LogLevel)

	// 设置控制台输出
	config["console"] = true

	configStr, err := json.Marshal(config)
	if err != nil {
		return
	}

	if config["level"] != "7" {
		err = logs.SetLogger(logs.AdapterFile, string(configStr))
	}
	return
}

func InitSec() (err error) {
	if gcommon.GameConf.LogLevel == "debug" {
		return
	}
	err = initLogger()
	if err != nil {
		return
	}
	// 设置日志级别
	return
}
