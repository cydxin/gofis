package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/model"
)

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
		fmt.Printf("UserGameInfo:%v \n", client.UserInfo)
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

	// 设置 client.UserGameInfo 中的字段
	client.UserInfo = &UserGameInfo{
		UserId:   UserId(userInfo.ID),
		Name:     userInfo.Account,
		NickName: userInfo.Nickname,
		GroupId:  userInfo.GroupID,
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

// 其他请求
func handleRequest(reqMap map[string]interface{}, client *Client) {
	if len(reqMap) > 0 {
		act := reqMap["event"]
		switch act {
		//PK记录
		case "pkRecord":
			break
		//比赛记录
		case "matchRecord":
			break
		//体验记录
		case "expRecord":
			break
		//进入房间
		case "enterRoom":
			roomNumFloat, okRoomNum := reqMap["roomNum"].(float64)
			if !okRoomNum {
				fmt.Printf("参数错判 %v: %v\n", okRoomNum, reqMap["roomNum"])
				client.sendMsg([]byte("参数错判"))
				return
			}
			// 将 float64 转换为 int
			roomNum := int(roomNumFloat)
			EnterRoom(roomNum, client)
			break
		case "ready":
			client.IsReady = true
			//fmt.Printf("client准备 %v\n", client)

			client.sendMsg([]byte("收到了你的准备"))
			break
		case "catch_fish":
			if len(reqMap) < 2 {
				return
			}
			fmt.Printf("reqmap格式 %v\n", reqMap)
			BulletIdFloat64, errBulletId := reqMap["BulletId"].(float64)
			FishIdFloat64, errFishId := reqMap["FishId"].(float64)
			if !errBulletId || !errFishId {
				fmt.Printf("错误参数:\n errBulletId:%v ; errFishId:%v \n", errBulletId, errFishId)
			}
			bulletId := BulletId(BulletIdFloat64)
			fishId := FishId(FishIdFloat64)
			catchFishReq := catchFishReq{
				BulletId: bulletId, // 使用类型断言将值转换为 BulletId 类型
				FishId:   fishId,   // 使用类型断言将值转换为 FishId 类型
			}
			client.catchFish(catchFishReq.FishId, catchFishReq.BulletId)
			break
		}
	}
}
