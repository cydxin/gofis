package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
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
	MaxOdsFishId    FishId
	NextOdsFishId   FishId
	robotAddTicker  *time.Ticker
	maxFishIdOds    FishId //鱼最大赔率 共享给全部机器人
	nextFishIdOds   FishId //鱼次大赔率 共享给全部机器人
	randFishIdOds   FishId //鱼随机赔率 共享给全部机器人
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
	client.UserInfo.Client = client
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
			if !client.IsReady && client.UserInfo.GroupId != 2 {
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
		timeTick, _ := randomTicker(30*time.Second, 60*time.Second)
		room = &Room{
			ID:             roomID,
			MaxPlayers:     roomNum,
			IsClose:        false,
			CloseChan:      make(chan bool),
			Players:        make(map[int]*Client),
			robotAddTicker: timeTick,
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
					room.robotAddTicker, _ = randomTicker(30*time.Second, 60*time.Second)
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
		UserId: common.UserId(rand.Intn(500)),
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
func randomTicker(min, max time.Duration) (*time.Ticker, time.Duration) {
	interval := time.Duration(rand.Int63n(int64(max-min)) + int64(min))
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
	close(room.CloseChan) //防止意外的没有关闭
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
			if client.UserInfo.UserId > 0 && client.UserInfo.GroupId == 1 {
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
	// 启动每个机器人的操作goroutine
	for _, robot := range robList {
		go func(robot *UserGameInfo) {
			//初始化运行
			switchTurret(robot)
			switchTurretSpeed(robot)
			switchLockFish(robot)
			go performRobotAction(robot)
			//切换炮台
			switchTurretTimer, _ := randomTicker(10, 30)
			//切换炮速
			gunSpeedTimer, _ := randomTicker(3, 8)
			//切换瞄准单位  使用射速判定时间 1.0攻速为例子，一秒射一下， 三到十颗子弹就是 3到十秒
			switchLockFishTimer, _ := randomTicker(time.Duration(3*robot.GameConfig.HitSpeed), time.Duration(8*robot.GameConfig.HitSpeed))
			// 执行机器人的操作逻辑
			for {
				select {
				case <-room.CloseChan:
					return
				case <-switchTurretTimer.C:
					switchTurret(robot)
				case <-gunSpeedTimer.C:
					switchTurretSpeed(robot)
				case <-switchLockFishTimer.C:
					switchLockFish(robot)
				default:
				}
			}
		}(robot)
	}
}

func performRobotAction(robot *UserGameInfo) {
	//robot.GameConfig.Room.FishGroup
	//射速 * 此次时间 等于需要循环的次数
	//shot := time.NewTicker(time.Second * )
	var bulletStartingPoint string
	var bulletEndPoint string
	select {
	case <-robot.GameConfig.Room.CloseChan:
		return
	default:
		for _, fish := range robot.GameConfig.Room.FishGroup {
			if !fish.ToBeDeleted {
				bulletStartingPoint = "0,0"
				bulletEndPoint = fmt.Sprintf("%d,%d", fish.CurrentX, fish.CurrentY)
			}
		}
		FireBulletsResult := []interface{}{"fire_bullets",
			map[string]interface{}{
				"userId":              robot.UserId,
				"bulletId":            robot.BulletLevel,
				"bulletStartingPoint": bulletStartingPoint,
				"bulletEndPoint":      bulletEndPoint,
			},
		}
		robot.GameConfig.Room.broadcast(FireBulletsResult)
		time.Sleep(time.Second * (1 / time.Duration(robot.GameConfig.HitSpeed)))
	}
}

func switchLockFish(robot *UserGameInfo) {
	probability := rand.Intn(100) + 1 // 概率
	switch {
	case probability <= 50: //50的概率 赔率最高的鱼
		robot.GameConfig.LockFishType = "max"
	case probability <= 80: //51 - 80 30赔率次高的鱼
		robot.GameConfig.LockFishType = "next"
	default: //81-10 随机目标
		robot.GameConfig.LockFishType = "rand"
	}
}

func switchTurretSpeed(robot *UserGameInfo) {
	//上限速度： 2次射击在一秒内
	probability := rand.Intn(100) + 1 // 概率
	switch {
	case probability <= 50: //50的概率 上限速度
		robot.GameConfig.HitSpeed = 2
	case probability <= 85: //51 - 85 35的概率随机射速
		robot.GameConfig.HitSpeed = rand.Float32()*(2.0-0.1) + 0.1
	default:
		robot.GameConfig.HitSpeed = 0
	}
}
func switchTurret(robot *UserGameInfo) {
	probability := rand.Intn(100) + 1 // 概率
	if robot.Score > robot.GameConfig.InitScore {
		switch {
		case probability >= 81: //81 - 100 20的概率2 1 随机一个
			//10的概率二分炮或一分
			robot.BulletLevel = rand.Intn(2) + 1
		case probability >= 51: //51 - 80 30的概率 4 5 随机一个
			//15的概率四分炮或五分
			robot.BulletLevel = rand.Intn(2) + 4
		case probability <= 50: //1 - 50 50的概率 3
			//50的概率三分炮
			robot.BulletLevel = 3
		}
	} else {
		switch {
		case probability <= 50: //1 - 50 50的概率 1
			//50的概率一分炮
			robot.BulletLevel = 1
		case probability <= 90: //51 - 90 40的概率2 3  随机一个
			//20的概率二或三分炮
			robot.BulletLevel = rand.Intn(2) + 2
		case probability >= 91: //91 - 100 10的概率4 5 随机一个
			//5的概率四分炮或五分
			robot.BulletLevel = rand.Intn(2) + 4
		}
	}
	switchTurretResult := []interface{}{"switch_turret",
		map[string]interface{}{
			"user_id":      robot.UserId,
			"bullet_level": robot.BulletLevel,
		},
	}
	robot.GameConfig.Room.broadcast(switchTurretResult)
}
