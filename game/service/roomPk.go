package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"gofish/model"
	"math/rand"
	"sort"
	"sync"
	"time"
)

var nextPkRoomID int

type roomPkName string

var roomsPk = make(map[int]*RoomPk)
var roomPkWait = make(map[int]map[string]*RoomPk)
var roomPkNames = map[int]string{
	1: "体验场",
	2: "2人PK场",
	3: "3人PK场",
	4: "4人PK场",
	5: "5人PK场",
	6: "6人PK场",
}
var roomPkMutex sync.Mutex

// RoomPk 结构体定义
type RoomPk struct {
	ID              int
	Name            roomPkName
	Status          int //房间的状态
	MaxPlayers      int
	AllPlayersReady bool
	IsReady         bool
	IsClose         bool             // 布尔类型的变量，用于表示房间是否关闭
	RoomChan        chan bool        // 通道，用于进行信号通知，比如关闭房间
	Players         map[int]*Client  // 使用map存储玩家信息，以玩家ID作为键
	FishGroup       map[FishId]*Fish // 使用map存储鱼，以鱼ID作为键  todo:换回切片
	MaxOdsFishId    FishId
	NextOdsFishId   FishId
	robotAddTicker  *time.Ticker
	RoomConfig      *common.PkRoomInfo
	Type            string
	maxFishIdOds    FishId //鱼最大赔率 共享给全部机器人
	nextFishIdOds   FishId //鱼次大赔率 共享给全部机器人
	randFishIdOds   FishId //鱼随机赔率 共享给全部机器人
	mutex           sync.Mutex
	fishMutex       sync.Mutex
}

func (room *RoomPk) handFishInit() {

}

// EnterPkRoom 进入房间逻辑
func EnterPkRoom(roomNum int, roomLevel string, client *Client) {
	//使用roomMutex 确保getOrCreateRoom的唯一性即 获取房间时的并发是安全的
	roomPkMutex.Lock()
	var room *RoomPk // 声明一个 RoomPk 指针
	if client.Room != nil {
		logs.Debug("机器人进入房间")
		room = client.Room
	} else {
		//获取等待中的房间
		r, err := getOrCreatePkRoom(roomNum, roomLevel)
		if err != nil {
			roomPkMutex.Unlock()
			errorMsg := err.Error()
			marshal, _ := json.Marshal([]interface{}{"err", map[string]interface{}{"message": errorMsg}})
			client.sendMsg(marshal)
			return
		}
		room = r
	}

	logs.Debug("roomPkMutex.Unlock()")
	roomPkMutex.Unlock()
	logs.Debug("room.mutex.Lock()")
	logs.Debug(room)
	room.mutex.Lock()
	defer func() {
		room.mutex.Unlock()
		logs.Debug("room.mutex.Unlock()")

	}()
	//数据操作
	logs.Debug("数据操作")
	room.Players[int(client.UserGameInfo.UserId)] = client
	client.Room = room
	client.UserGameInfo.Client = client
	client.UserGameInfo.SeatIndex = len(room.Players)
	client.UserGameInfo.Score = room.RoomConfig.InitScore
	logs.Debug("初始配置数据 优先机器人使用")

	//初始配置数据 优先机器人使用
	client.UserGameInfo.Score = room.RoomConfig.InitScore
	client.UserGameInfo.GameConfig = &PlayerConfiguration{
		InitScore: room.RoomConfig.InitScore, //初始积分
		Power:     1,
		RoomPk:    room,
	}

	logs.Debug("标记用户状态为房间中")
	//标记用户状态为房间中
	client.UserGameInfo.setOnline(2) // 标记用户状态为游戏中
	//广播给房间内的人 谁谁谁加入了
	room.broadcast([]interface{}{"newJoin", map[string]interface{}{"userInfo": client.UserGameInfo}})
	//等待handFishInit完成
	if len(room.Players) == roomNum { // 人满了 开始游戏！！！
		logs.Debug("人满了")
		room.robotAddTicker.Stop()
		roomPkWait[roomNum][roomLevel] = nil // 弹出此等待的房间
		room.Status = 1                      // 标记为游戏中状态
		client.UserGameInfo.setOnline(3)     // 标记用户状态为游戏中

		go handFishInit(room) //todo: 已经把锁去除 后续有问题可以加上
		room.broadcast([]interface{}{"message", map[string]interface{}{"message": "等待其他玩家资源加载就绪"}})
		go handPkRoomRun(room)
	}
}

func handPkRoomRun(room *RoomPk) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	// 需要让玩家发送一个ready表示可以开始加载数据  todo:改为chan传输 不做循环处理
	logs.Debug("检测玩家准备中")
	logs.Debug(room.AllPlayersReady)
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
			logs.Debug("玩家都就绪")
			// 发送给前端
			room.broadcast([]interface{}{"message", map[string]interface{}{"message": "开始推送"}})
			room.AllPlayersReady = allReady
			//开始鱼数据实时状态和数据的发送
			//go handFishInit(room)      //鱼群初始化数据
			go closePkRoomTimer(room) //定义定时器关闭
			go handFishRun(room)
			return
		}
		logs.Debug("检测准备")
		time.Sleep(time.Second * 2)
	}
}

// 外部加锁，不能不加
func getOrCreatePkRoom(roomNum int, roomLevel string) (*RoomPk, error) {
	logs.Debug("获取房间")
	roomKey := roomNum
	roomsByNum, exists := roomPkWait[roomKey]

	if !exists {
		roomsByNum = make(map[string]*RoomPk)
		roomPkWait[roomKey] = roomsByNum
	}

	room, exists := roomsByNum[roomLevel]
	if !exists || room == nil {
		nextPkRoomID++
		roomID := nextPkRoomID
		//获取房间的配置
		roomConfig := getRoomPkConfig(roomNum, roomLevel)
		if roomConfig == nil {
			message := fmt.Errorf("不存在的的房间或房间已关闭")
			logs.Debug(message)
			return nil, message
		}
		// 在创建房间时初始化 RoomChan
		timeTick, _ := randomTicker(1, 5)
		room = &RoomPk{
			ID:             roomID,
			Name:           roomPkName(roomPkNames[roomNum]),
			MaxPlayers:     roomNum,
			IsClose:        false,
			RoomChan:       make(chan bool),
			Players:        make(map[int]*Client),
			robotAddTicker: timeTick,
			RoomConfig:     roomConfig,
			Type:           "Pk",
		}
		roomsPk[roomID] = room
		roomPkWait[roomNum][roomLevel] = room
		// 启动定时添加机器人的协程
		go robotRun(room)
	}
	logs.Debug("返回房间数据")
	return room, nil
}

func getRoomPkConfig(num int, name string) *common.PkRoomInfo {
	roomConfig, err := model.GetConfigFromRedis(num, name)
	if err != nil {
		return nil
	}
	logs.Debug("getRoomConfig的roomConfig：%v", roomConfig)
	return roomConfig

}

func closePkRoomTimer(room *RoomPk) {
	fmt.Print("定时器开始了 \n")
	//定时器
	timer := time.NewTimer(time.Duration(room.RoomConfig.DurationMin) * time.Second)
	val := time.Duration(room.RoomConfig.DurationMin) * time.Minute
	logs.Debug("房间定时val %v", val)
	gameProgressSendTime := time.NewTimer(time.Minute)
	nowProgress := time.Duration(room.RoomConfig.DurationMin) * time.Minute
	//监听定时器管道和手动关闭管道
	for {
		select {
		case <-timer.C:
			logs.Debug("关闭RoomChan")
			close(room.RoomChan)
			return
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
		default:
		}
	}

}

// 房间结束
func (room *RoomPk) closePkRoom() {
	// 对所有玩家按照积分进行排序
	var players []*Client
	for _, player := range room.Players {
		players = append(players, player)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].UserGameInfo.Score > players[j].UserGameInfo.Score
	})
	//用户修改数据
	var usersToUpdate []model.UserCloseRoomSetData
	//room.AllPlayersReady = false
	var balanceAdd int
	for i, client := range players {
		// 确定名次信息
		rankString := fmt.Sprintf("第%d名", i+1)
		// 计算获得积分
		score := fmt.Sprintf("无奖励")
		balanceAdd = 0

		if i == 0 {
			logs.Debug(client.UserGameInfo.Balance)
			client.UserGameInfo.Balance += room.RoomConfig.Money
			logs.Debug(client.UserGameInfo.Balance)
			score = fmt.Sprintf("奖励积分：%v", room.RoomConfig.Money)
			balanceAdd = int(room.RoomConfig.Money)
		}
		// 发送结算信息
		closeRoom := []interface{}{"closePkRoom", map[string]interface{}{
			"message": "时间结束",
			"score":   score, // 更新后的积分
			"rank":    rankString,
		}}

		data := model.UserCloseRoomSetData{
			ID:       client.UserGameInfo.UserId,
			IsOnline: 1,
			PKMoney:  balanceAdd,
			Balance:  client.UserGameInfo.Balance,
		}
		client.UserGameInfo.Online = 1
		usersToUpdate = append(usersToUpdate, data)
		client.IsReady = false
		client.broadcast(closeRoom)
		client.Room = nil
	}
	model.SetUserCloseRoomData(usersToUpdate)
}

// 返回一个随机时间间隔的 Ticker
func randomTicker(min, max time.Duration) (*time.Ticker, time.Duration) {
	interval := time.Duration(rand.Int63n(int64(max-min))) + min*time.Second
	ticker := time.NewTicker(interval)
	return ticker, interval
}

// 广播鱼群数据
func (room *RoomPk) broadcastFishLocation() {
	FishReady := []interface{}{"fishLocation", room.FishGroup}
	// 发送给前端
	room.broadcast(FishReady)
}

// 广播数据准备就绪消息
func (room *RoomPk) broadcastFishReady() {
	FishReady := []interface{}{"message",
		map[string]interface{}{
			"message": "鱼群就绪",
		}}
	// 发送给前端
	room.broadcast(FishReady)
}

// 自定义广播
func (room *RoomPk) broadcast(data []interface{}) {
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
