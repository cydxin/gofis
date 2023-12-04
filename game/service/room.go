package service

import (
	"encoding/json"
	"fmt"
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
	Status          int //房间状态
	MaxPlayers      int
	AllPlayersReady bool
	IsReady         bool
	IsClose         bool            // 布尔类型的变量，用于表示房间是否关闭
	CloseChan       chan bool       // 通道，用于进行信号通知，比如关闭房间
	Players         map[int]*Client // 使用map存储玩家信息，以玩家ID作为键
	FishGroup       []*Fish
	mutex           sync.Mutex
}

// EnterRoom 进入房间逻辑
func EnterRoom(roomNum int, client *Client) {
	room := getOrCreateRoom(roomNum)
	room.mutex.Lock()
	defer room.mutex.Unlock()
	fmt.Print("EnterRoom \n")
	room.Players[int(client.UserInfo.UserId)] = client
	//等待handFishInit完成
	handFishInit(room)
	printFishGroupJSON(room)
	if len(room.Players) == roomNum { // 人满了 开始游戏！！！
		//delete(roomWait, roomNum) // 把当前等待房间移除
		roomWait[roomNum] = nil
		room.Status = 1 // 标记为游戏中状态
		fmt.Print("人满了 开!!! \n")
		go handRoomRun(room)
	}
}
func handRoomRun(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	// 需要让玩家发送一个ready表示可以开始加载数据  todo:改为chan传输 不做循环处理

	fmt.Print("开始判断玩家是否可以接收数据")
	for !room.AllPlayersReady {
		allReady := true
		for _, client := range room.Players {
			// 所有玩家进入游戏 加载资源后 再开始发送鱼的数据，
			//或许我可以先准备资源，而不是等玩家满了再准备
			if !client.IsReady {
				//fmt.Printf("玩家 %v 没有准备 \n", client.UserInfo.Name)
				allReady = false
				break
			}
		}
		if allReady {
			fmt.Print("玩家可以接收数据")
			room.AllPlayersReady = true
			//开始鱼数据实时状态和数据的发送
			//go handFishInit(room)      //鱼群初始化数据
			go closeRoomTimer(room, 5) //定义定时器关闭，其实可以直接放置在handFishrun
			go handFishrun(room)
			break
		}
	}
}
func getOrCreateRoom(roomNum int) *Room {
	roomMutex.Lock()
	defer roomMutex.Unlock()
	room := roomWait[roomNum]
	if room == nil {
		nextRoomID++
		roomID := nextRoomID

		// 在创建房间时初始化 CloseChan
		room = &Room{
			ID:         roomID,
			MaxPlayers: roomNum,
			IsClose:    false,
			CloseChan:  make(chan bool),
			Players:    make(map[int]*Client),
		}
		rooms[roomID] = room
		roomWait[roomNum] = room
	}
	return room
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

func closeRoom(room *Room) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	//结算 妈的 写个统一给房间发送信息 的 方法
	room.sendMsgAllPlayer()
}

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
func printFishGroupJSON(room *Room) {
	fishGroupJSON, err := json.MarshalIndent(room.FishGroup, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling FishGroup to JSON:", err)
		return
	}

	fmt.Println("FishGroup JSON:")
	fmt.Println(string(fishGroupJSON))
}
