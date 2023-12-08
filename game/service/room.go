package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math/rand"
	"sync"
	"time"
)

var nextRoomID int

var rooms = make(map[int]*Room)
var roomWait = map[int]*Room{
	1: nil,
	2: nil,
	3: nil,
	4: nil,
	5: nil,
	6: nil,
}
var roomMutex sync.Mutex

// Room 结构体定义
type Room struct {
	ID              int
	Status          int //房间的状态
	MaxPlayers      int
	AllPlayersReady bool
	IsReady         bool
	IsClose         bool             // 布尔类型的变量，用于表示房间是否关闭
	CloseChan       chan bool        // 通道，用于进行信号通知，比如关闭房间
	Players         map[int]*Client  // 使用map存储玩家信息，以玩家ID作为键
	FishGroup       map[FishId]*Fish // 使用map存储鱼，以鱼ID作为键  todo:换回切片
	robotAddTicker  *time.Ticker
	mutex           sync.Mutex
	fishMutex       sync.Mutex
}

// EnterRoom 进入房间逻辑
func EnterRoom(roomNum int, client *Client) {
	//使用roomMutex 确保getOrCreateRoom的唯一性即 获取房间时的并发是安全的
	roomMutex.Lock()
	room := getOrCreateRoom(roomNum)
	defer roomMutex.Unlock()
	room.Players[int(client.UserInfo.UserId)] = client
	client.Room = room
	room.broadcast([]interface{}{"newJoin", map[string]interface{}{"userInfo": client.UserInfo}})
	//等待handFishInit完成
	if len(room.Players) == roomNum { // 人满了 开始游戏！！！
		roomWait[roomNum] = nil
		room.Status = 1       // 标记为游戏中状态
		go handFishInit(room) //todo: 已经把锁去除 后续有问题可以加上
		room.broadcast([]interface{}{"message", map[string]interface{}{"message": "等待其他玩家资源加载就绪"}})
		go handRoomRun(room)
	}
}
func handRoomRun(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	// 需要让玩家发送一个ready表示可以开始加载数据  todo:改为chan传输 不做循环处理
	fmt.Print("等待全部玩家准备\n")
	for !room.AllPlayersReady {
		allReady := true
		for _, client := range room.Players {
			// 所有玩家进入游戏 加载资源后 再开始发送鱼的数据，
			//或许我可以先准备资源，而不是等玩家满了再准备
			if !client.IsReady {
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
			go closeRoomTimer(room, 5) //定义定时器关闭，其实可以直接放置在handFish run
			go handFishRun(room)
			break
		}
	}
}
func getOrCreateRoom(roomNum int) *Room {
	room := roomWait[roomNum]
	if room == nil {
		nextRoomID++
		roomID := nextRoomID

		// 在创建房间时初始化 CloseChan
		room = &Room{
			ID:             roomID,
			MaxPlayers:     roomNum,
			IsClose:        false,
			CloseChan:      make(chan bool),
			Players:        make(map[int]*Client),
			robotAddTicker: randomTicker(30*time.Second, 60*time.Second),
		}
		rooms[roomID] = room
		roomWait[roomNum] = room

		// 启动定时添加机器人的协程
		go func() {
			for {
				select {
				case <-room.robotAddTicker.C:
					room.mutex.Lock()
					//先判断人满了没
					if room.MaxPlayers == len(room.Players) {
						room.robotAddTicker.Stop()
						return
					}
					addRobotToRoom(room.MaxPlayers)
					room.robotAddTicker.Stop()
					room.robotAddTicker = randomTicker(30*time.Second, 60*time.Second)
					room.mutex.Unlock()
				case <-room.CloseChan:
					return
				}
			}
		}()
	}
	return room
}

// 添加机器人到房间
func addRobotToRoom(roomNum int) {
	robotClient := &Client{
		UserInfo: generateRobotUserInfo(),
	}
	// 调用 EnterRoom 函数将机器人加入房间
	EnterRoom(roomNum, robotClient)
}

// 生成机器人用户信息
func generateRobotUserInfo() *UserGameInfo {
	// 实现根据需要生成机器人用户信息的逻辑
	return &UserGameInfo{
		UserId: UserId(rand.Intn(500)),
	}
}
func closeRoomTimer(room *Room, min int) {
	fmt.Print("定时器开始了 \n")
	//定时器
	timer := time.NewTimer(time.Duration(min) * time.Minute)
	//监听定时器管道和手动关闭管道
	select {
	case <-timer.C:
		room.CloseChan <- false
	case <-room.CloseChan:
		timer.Stop()
	}
}

// 时间范围
// randomTicker 返回一个随机时间间隔的 Ticker
func randomTicker(min, max time.Duration) *time.Ticker {
	interval := time.Duration(rand.Int63n(int64(max-min)) + int64(min))
	return time.NewTicker(interval)
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
	close(room.CloseChan) //防止意外的没有关闭
}

// func (robotJoinRoom)  {
//
// }
// 使用 interface实现自定义
func (room *Room) sendMsgAllPlayer(messages ...interface{}) {
	var byteMessages []byte //定义一个byte
	for _, msg := range messages {
		switch v := msg.(type) {
		case string:
			byteMessages = append(byteMessages, []byte(v)...)
		case []byte:
			byteMessages = append(byteMessages, v...)
		default:
			byteMessages = append(byteMessages, []byte(fmt.Sprintf("%v", msg))...)
		}
	}
	// 发送消息给房间中的所有玩家
	for _, client := range room.Players {
		client.sendMsg(byteMessages)
	}
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
			if client.UserInfo.UserId > 0 {
				client.sendMsg(dataByte)
			}
		}
	}
}

func (room *Room) handRobotRun() {
	//判断是否有机器人 有的话进行机器人操作
	robList := make([]*UserGameInfo, 0)
	for _, client := range room.Players {
		if client.UserInfo.GroupId == 2 {
			robList = append(robList, client.UserInfo)
		}
	}
	if len(robList) == 0 {
		return
	}
}
