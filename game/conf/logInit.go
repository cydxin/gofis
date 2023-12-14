package conf

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/game/common"
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
	config["filename"] = common.GameConf.LogPath
	config["level"] = conversionLogLevel(common.GameConf.LogLevel)

	// 设置控制台输出
	config["console"] = true

	configStr, err := json.Marshal(config)
	if err != nil {
		return
	}

	err = logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func InitSec() (err error) {
	err = initLogger()
	if err != nil {
		return
	}

	// 设置日志级别
	logs.SetLevel(conversionLogLevel(common.GameConf.LogLevel))

	// 输出当前日志级别
	fmt.Printf("Current log level: %s\n", common.GameConf.LogLevel)

	return
}
