package service

import (
	"fmt"
	"gofish/model"
)

func handlePKRecord(reqMap map[string]interface{}, client *Client) {
	Page := 1
	if PageFloat, okPage := reqMap["page"].(float64); !okPage {
		sprintf := fmt.Sprintf("错误的page值：%v，	获取到的参数类型 %t \n", reqMap["page"], reqMap["page"])
		client.sendMsg([]byte(sprintf))
		return
	} else {
		Page = int(PageFloat)
	}
	pkRecord, err := model.GetPkRecordsThroughUsers(Page, client.UserGameInfo.UserId)
	if err != nil {
		sprintf := fmt.Sprintf("err：%v \n", err)
		client.sendMsg([]byte(sprintf))
		// 处理错误
		return
	}
	client.ModelSuccess(pkRecord)
}
func handleMatchRecord(reqMap map[string]interface{}, client *Client) {
	page, okPage := reqMap["page"].(float64)
	if !okPage {
		sprintf := fmt.Sprintf("错误的page值: %v,page得到的类型 %t \n", reqMap["page"], reqMap["page"])
		client.sendMsg([]byte(sprintf))
		return
	}
	matchRecord, err := model.GetMatchRecordsThroughUsers(int(page), client.UserGameInfo.UserId)
	if err != nil {
		sprintf := fmt.Sprintf("数据错误 :%v \n", err)
		client.sendMsg([]byte(sprintf))
		return
	}
	client.ModelSuccess(matchRecord)
}
func handleExpRecord(reqMap map[string]interface{}, client *Client) {
	page, okPage := reqMap["page"].(float64)
	if !okPage {
		sprintf := fmt.Sprintf("错误的page值: %v,page得到的类型 %t \n", reqMap["page"], reqMap["page"])
		client.sendMsg([]byte(sprintf))
		return
	}
	matchRecord, err := model.GetMatchRecordsThroughUsers(int(page), client.UserGameInfo.UserId)
	if err != nil {
		sprintf := fmt.Sprintf("数据错误 :%v \n", err)
		client.sendMsg([]byte(sprintf))
		return
	}
	client.ModelSuccess(matchRecord)
}

func handleEnterRoom(reqMap map[string]interface{}, client *Client) {
	roomNumFloat, okRoomNum := reqMap["roomNum"].(float64)
	roomLevel, okRoomLevel := reqMap["roomLevel"].(string)
	if !okRoomNum {
		fmt.Printf("参数错判,roomNum的类型 %t： , %v \n", reqMap["roomNum"], reqMap["roomNum"])
		client.sendMsg([]byte("参数错判"))
		return
	}
	if !okRoomLevel {
		fmt.Printf("参数错判,okroomLevel的类型 %t： , %v \n", reqMap["okRoomLevel"], reqMap["okRoomLevel"])
		client.sendMsg([]byte("参数错判"))
		return
	}
	// 将 float64 转换为 int
	roomNum := int(roomNumFloat)
	EnterRoom(roomNum, roomLevel, client)
}

func handleReady(reqMap map[string]interface{}, client *Client) {
	client.IsReady = true
	client.sendMsg([]byte("收到了你的准备"))
}
func handleTouchFish(reqMap map[string]interface{}, client *Client) {
	fmt.Printf("reqmap格式 %v\n", reqMap)
	BulletIdFloat64, errBulletId := reqMap["BulletId"].(int)
	FishIdFloat64, errFishId := reqMap["FishId"].(int)
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
}
func handleFireBullets(reqMap map[string]interface{}, client *Client) {
	fmt.Printf("reqmap格式 %v\n", reqMap)
	bulletIdFloat64, errBulletId := reqMap["BulletId"].(int)      //子弹等级
	bulletStartingPoint := reqMap["bulletStartingPoint"].(string) //子弹起点 xy
	bulletEndPoint := reqMap["bulletEndPoint"].(string)           //子弹终点 xy
	if !errBulletId {
		fmt.Printf("错误参数:\n errBulletId:%v ; errStartingPoint:%v ; errEndPoint:%v \n", errBulletId, bulletStartingPoint, bulletEndPoint)
	}
	bulletId := BulletId(bulletIdFloat64)
	FireBulletsResult := []interface{}{"fire_bullets",
		map[string]interface{}{
			"userId":              client.UserGameInfo.UserId,
			"bulletId":            bulletId,
			"bulletStartingPoint": bulletStartingPoint,
			"bulletEndPoint":      bulletEndPoint,
		},
	}
	client.Room.broadcast(FireBulletsResult)
}
