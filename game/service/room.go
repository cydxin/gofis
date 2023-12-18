package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"gofish/model"
	"math/rand"
	"sync"
	"time"
)

var nextRoomID int

type roomName string

var rooms = make(map[int]*Room)
var roomWait = make(map[int]map[string]*Room)
var roomNames = map[int]string{
	1: "体验场",
	2: "2人PK场",
	3: "3人PK场",
	4: "4人PK场",
	5: "5人PK场",
	6: "6人PK场",
}
var roomMutex sync.Mutex

// Room 结构体定义
type Room struct {
	ID              int
	Name            roomName
	Status          int //房间的状态
	MaxPlayers      int
	AllPlayersReady bool
	IsReady         bool
	IsClose         bool             // 布尔类型的变量，用于表示房间是否关闭
	CloseChan       chan bool        // 通道，用于进行信号通知，比如关闭房间
	Players         map[int]*Client  // 使用map存储玩家信息，以玩家ID作为键
	FishGroup       map[FishId]*Fish // 使用map存储鱼，以鱼ID作为键  todo:换回切片
	MaxOdsFishId    FishId
	NextOdsFishId   FishId
	robotAddTicker  *time.Ticker
	PkRoomConfig    *common.PkRoomInfo
	maxFishIdOds    FishId //鱼最大赔率 共享给全部机器人
	nextFishIdOds   FishId //鱼次大赔率 共享给全部机器人
	randFishIdOds   FishId //鱼随机赔率 共享给全部机器人
	mutex           sync.Mutex
	fishMutex       sync.Mutex
}

// EnterRoom 进入房间逻辑
func EnterRoom(roomNum int, roomLevel string, client *Client) {
	logs.Debug("进入房间")
	//使用roomMutex 确保getOrCreateRoom的唯一性即 获取房间时的并发是安全的
	roomMutex.Lock()
	//获取等待中的房间
	room, err := getOrCreateRoom(roomNum, roomLevel)
	if err != nil {
		errorMsg := err.Error()
		marshal, _ := json.Marshal([]interface{}{"err", map[string]interface{}{"message": errorMsg}})
		client.sendMsg(marshal)
		return
	}
	defer roomMutex.Unlock()
	//数据操作
	room.Players[int(client.UserGameInfo.UserId)] = client
	client.Room = room
	client.UserGameInfo.Client = client
	client.UserGameInfo.SeatIndex = len(room.Players)
	client.UserGameInfo.Score = room.PkRoomConfig.InitScore
	client.UserGameInfo.GameConfig = &PlayerConfiguration{
		InitScore:    room.PkRoomConfig.InitScore,
		Power:        1,
		HitSpeed:     1,
		Room:         room,
		LockFishType: "",
	}
	//广播给房间内的人 谁谁谁加入了
	room.broadcast([]interface{}{"newJoin", map[string]interface{}{"userInfo": client.UserGameInfo}})
	//等待handFishInit完成
	if len(room.Players) == roomNum { // 人满了 开始游戏！！！
		room.robotAddTicker.Stop()
		roomWait[roomNum][roomLevel] = nil //弹出此等待的房间
		room.Status = 1                    // 标记为游戏中状态
		go handFishInit(room)              //todo: 已经把锁去除 后续有问题可以加上
		room.broadcast([]interface{}{"message", map[string]interface{}{"message": "等待其他玩家资源加载就绪"}})
		go handRoomRun(room)
	}
}
func handRoomRun(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	// 需要让玩家发送一个ready表示可以开始加载数据  todo:改为chan传输 不做循环处理
	logs.Debug("检测玩家准备中")
	for !room.AllPlayersReady {
		allReady := true
		for _, client := range room.Players {
			// 所有玩家进入游戏 加载资源后 再开始发送鱼的数据，
			//或许我可以先准备资源，而不是等玩家满了再准备
			if !client.IsReady && client.UserGameInfo.GroupId != 2 {
				allReady = false
				break
			}
		}
		if allReady {
			// 发送给前端
			room.broadcast([]interface{}{"message", map[string]interface{}{"message": "开始推送"}})
			room.AllPlayersReady = allReady
			//开始鱼数据实时状态和数据的发送
			//go handFishInit(room)      //鱼群初始化数据
			go closeRoomTimer(room) //定义定时器关闭
			go handFishRun(room)
			break
		}
	}
}
func getOrCreateRoom(roomNum int, roomLevel string) (*Room, error) {
	roomKey := roomNum
	roomsByNum, exists := roomWait[roomKey]

	if !exists {
		roomsByNum = make(map[string]*Room)
		roomWait[roomKey] = roomsByNum
	}

	room, exists := roomsByNum[roomLevel]
	if !exists {
		nextRoomID++
		roomID := nextRoomID
		//获取房间的配置
		roomConfig := getRoomConfig(roomNum, roomLevel)
		if roomConfig == nil {
			message := fmt.Errorf("不存在的的房间或房间已关闭")
			logs.Debug(message)
			return nil, message
		}
		// 在创建房间时初始化 CloseChan
		timeTick, _ := randomTicker(1, 5)
		room = &Room{
			ID:             roomID,
			Name:           roomName(roomNames[roomNum]),
			MaxPlayers:     roomNum,
			IsClose:        false,
			CloseChan:      make(chan bool),
			Players:        make(map[int]*Client),
			robotAddTicker: timeTick,
			PkRoomConfig:   roomConfig,
		}
		rooms[roomID] = room
		roomWait[roomNum][roomLevel] = room

		// 启动定时添加机器人的协程
		go robotRun(room)
	}
	return room, nil
}

func getRoomConfig(num int, name string) *common.PkRoomInfo {
	roomConfig, err := model.GetConfigFromRedis(num, name)
	if err != nil {
		return nil
	}
	logs.Debug("getRoomConfig的roomConfig：%v", roomConfig)
	return roomConfig

}

func closeRoomTimer(room *Room) {
	fmt.Print("定时器开始了 \n")
	//定时器
	timer := time.NewTimer(time.Duration(room.PkRoomConfig.DurationMin) * time.Second)
	val := time.Duration(room.PkRoomConfig.DurationMin) * time.Minute
	logs.Debug("房间定时val %v", val)
	gameProgressSendTime := time.NewTimer(time.Minute)
	nowProgress := time.Duration(room.PkRoomConfig.DurationMin) * time.Minute
	//监听定时器管道和手动关闭管道
	select {
	case <-timer.C:
		logs.Debug("closeRoomTimer,<-timer.C")
		room.CloseChan <- false
	case <-room.CloseChan:
		logs.Debug("closeRoomTimer,<-room.CloseChan")
		timer.Stop()
	case <-gameProgressSendTime.C:
		nowProgress -= time.Minute
		minutes := int(nowProgress.Minutes())
		seconds := int(nowProgress.Seconds())
		data := []interface{}{"timeLeft",
			map[string]interface{}{
				"message": "剩余时间",
				"data": map[string]interface{}{
					"minutes": minutes,
					"seconds": seconds,
				},
			},
		}
		room.broadcast(data)
	}
}

// 时间范围
// randomTicker 返回一个随机时间间隔的 Ticker
func randomTicker(min, max time.Duration) (*time.Ticker, time.Duration) {
	interval := time.Duration(rand.Int63n(int64(max-min))) + min*time.Second
	ticker := time.NewTicker(interval)
	return ticker, interval
} // 时间范围
// randomTicker 返回一个随机时间间隔的 Ticker
func randomFTicker(min, max float32) (*time.Ticker, time.Duration) {
	minDuration := time.Duration(min * float32(time.Second))
	maxDuration := time.Duration(max * float32(time.Second))

	interval := time.Duration(rand.Int63n(int64(maxDuration-minDuration))) + minDuration
	ticker := time.NewTicker(interval)
	return ticker, interval
}

func closeRoom(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	//结算 妈的 写个统一给房间发送信息 的 方法
	closeRoom := []interface{}{"closeRoom", map[string]interface{}{
		"message":  "时间到",
		"integral": "10000分",
		"rank":     "第一名",
	}}
	for _, client := range room.Players {
		client.IsReady = false
	}
	// 发送给前端
	room.broadcast(closeRoom)
	close(room.CloseChan) //防止意外没有关闭
}

func (room *Room) broadcastFishLocation() {
	FishReady := []interface{}{"fishLocation", room.FishGroup}
	// 发送给前端
	room.broadcast(FishReady)
}

func (room *Room) broadcastFishReady() {
	FishReady := []interface{}{"message",
		map[string]interface{}{
			"message": "鱼群就绪",
		}}
	// 发送给前端
	room.broadcast(FishReady)
}

func (room *Room) broadcast(data []interface{}) {
	if dataByte, err := json.Marshal(data); err != nil {
		logs.Error("broadcast [%v] json marshal err :%v ", data, err)
	} else {
		dataByte = append([]byte{'4', '2'}, dataByte...)
		for _, client := range room.Players {
			if client.UserGameInfo.UserId > 0 && client.UserGameInfo.GroupId == 1 {
				client.sendMsg(dataByte)
			}
		}
	}
}
