package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"gofish/model"
)

type eventHandler func(reqMap map[string]interface{}, client *Client)

var eventHandlers = map[string]eventHandler{
	"pkRecord":    handlePKRecord,
	"matchRecord": handleMatchRecord,
	"expRecord":   handleExpRecord,
	"enterRoom":   handleEnterRoom,
	"ready":       handleReady,
	"touchFish":   handleTouchFish,
	"fireBullets": handleFireBullets,
}

func wsRequest(req []byte, client *Client) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("wsRequest panic:%v ", r)
		}
	}()

	if req[0] == 52 && req[1] == 50 { //ASCII
		reqMap := make(map[string]interface{}) //搞个map类型
		err := json.Unmarshal(req[2:], &reqMap)
		if err != nil {
			client.sendMsg([]byte("转换错误，数据格式不对"))
			return
		}
		event, okEvent := reqMap["event"].(string)
		if !okEvent {
			client.sendMsg([]byte("没有定义事件"))
			return
		}
		if len(reqMap) < 1 {
			client.sendMsg([]byte("参数长度错误"))
			return
		}
		fmt.Printf("reqMap:%v \n", reqMap)
		if client.UserInfo.UserId == 0 { //ws内无userinfo 说明还是没登录过的,要强制登录
			if event == "login" {
				handleLoginRequest(reqMap, client)
				fmt.Printf("UserGameInfo:%v \n", client.UserInfo)
			} else {
				client.sendMsg([]byte("尚未登录，请登录"))
				logs.Error("未定义的event %v", reqMap["event"])
			}
		} else {
			handleRequest(reqMap, client)
		}
	} else {
		logs.Error("invalid message %v", req)
	}
}

// 处理登录请求
func handleLoginRequest(reqMap map[string]interface{}, client *Client) {
	// 从 reqMap 中获取用户名和密码
	username, okUsername := reqMap["account"].(string)
	password, okPassword := reqMap["password"].(string)
	if !okUsername || !okPassword {
		client.sendMsg([]byte("参数错判"))
		return
	}
	// 查询用户数据
	userInfo, err := model.GetUserByCredentials(username, password)
	if err != nil {
		errStr := fmt.Sprintf("登录失败：%v", err)
		client.sendMsg([]byte(errStr))
		logs.Error("Login failed: %v", err)
		return
	}
	PlayerConfig := &PlayerConfiguration{
		InitScore: 0,
		Power:     0,
		HitSpeed:  0,
	}
	// 设置 client.UserGameInfo 中的字段
	client.UserInfo = &UserGameInfo{
		UserId:     common.UserId(userInfo.ID),
		Name:       userInfo.Account,
		NickName:   userInfo.Nickname,
		GroupId:    userInfo.GroupID,
		GameConfig: PlayerConfig,
	}
	client.IsReady = false
	// TODO: 发送登录成功的消息给客户端
	userInfoJSON, err := json.Marshal(userInfo)
	if err != nil {
		fmt.Println("转换为 JSON 出错:", err)
		return
	}

	userInfoJSON = append([]byte{'4', '1'}, userInfoJSON...)
	client.sendMsg(userInfoJSON)
}

// 处理其他请求
func handleRequest(reqMap map[string]interface{}, client *Client) {
	if len(reqMap) > 0 {
		act, ok := reqMap["event"].(string)
		if !ok {
			fmt.Printf("无效的事件：%v\n", reqMap["event"])
			return
		}

		// 从映射中获取事件处理函数
		handler, exists := eventHandlers[act]
		if !exists {
			fmt.Printf("未知的事件：%v\n", act)
			return
		}
		// 调用事件处理函数
		handler(reqMap, client)
	}
}
