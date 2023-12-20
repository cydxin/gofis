package service

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"gofish/model"
	"sort"
	"sync"
	"time"
)

var nextPkRoomMatchID int

type RoomMatchName string

var RoomMatchs = make(map[int]*RoomMatch)
var RoomMatchWait = make(map[int]*RoomMatch)
var RoomMatchNames = map[int]string{
	20:  "20人快速赛",
	50:  "50人争霸赛",
	100: "100人大奖赛",
}
var RoomMatchMutex sync.Mutex

// RoomMatchPk 结构体定义
type RoomMatch struct {
	ID              int
	Name            RoomMatchName
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
	RoomMatchConfig *common.RoomMatchInfo
	Type            string
	maxFishIdOds    FishId //鱼最大赔率 共享给全部机器人
	nextFishIdOds   FishId //鱼次大赔率 共享给全部机器人
	randFishIdOds   FishId //鱼随机赔率 共享给全部机器人
	mutex           sync.Mutex
	fishMutex       sync.Mutex
}

// 子RoomMatch 结构体定义
type RoomMatchSub struct {
	RoomChan        chan bool        // 通道，用于进行信号通知，比如关闭房间
	Players         []*Client        // 使用map存储玩家信息，以玩家ID作为键
	FishGroup       map[FishId]*Fish // 使用map存储鱼，以鱼ID作为键  todo:换回切片
	MaxOdsFishId    FishId
	NextOdsFishId   FishId
	robotAddTicker  *time.Ticker
	RoomMatchConfig *common.RoomMatchInfo
	Type            string
	maxFishIdOds    FishId //鱼最大赔率 共享给全部机器人
	nextFishIdOds   FishId //鱼次大赔率 共享给全部机器人
	randFishIdOds   FishId //鱼随机赔率 共享给全部机器人
	mutex           sync.Mutex
	fishMutex       sync.Mutex
}

// 子房间发送鱼的定位
func (room *RoomMatchSub) broadcastFishLocation() {
	FishReady := []interface{}{"fishLocation", room.FishGroup}
	// 发送给前端
	room.broadcast(FishReady)
}

// 子房间消息
func (room *RoomMatchSub) broadcast(data []interface{}) {
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

// EnterPkRoomMatch 进入房间逻辑
func EnterRoomMatch(RoomMatchNum int, client *Client) {
	//使用RoomMatchMutex 确保getOrCreateRoomMatch的唯一性即 获取房间时的并发是安全的
	RoomMatchMutex.Lock()
	var RoomMatch *RoomMatch // 声明一个 RoomMatchPk 指针
	if client.RoomMatch != nil {
		logs.Debug("机器人进入房间")
		RoomMatch = client.RoomMatch
	} else {
		//获取等待中的房间
		r, err := getOrCreateRoomMatch(RoomMatchNum)
		if err != nil {
			RoomMatchMutex.Unlock()
			errorMsg := err.Error()
			marshal, _ := json.Marshal([]interface{}{"err", map[string]interface{}{"message": errorMsg}})
			client.sendMsg(marshal)
			return
		}
		RoomMatch = r
	}

	logs.Debug("RoomMatchPkMutex.Unlock()")
	RoomMatchMutex.Unlock()
	logs.Debug("RoomMatch.mutex.Lock()")
	logs.Debug(RoomMatch)
	RoomMatch.mutex.Lock()
	defer func() {
		RoomMatch.mutex.Unlock()
		logs.Debug("RoomMatch.mutex.Unlock()")
	}()
	//数据操作
	logs.Debug("数据操作")
	RoomMatch.Players[int(client.UserGameInfo.UserId)] = client
	client.RoomMatch = RoomMatch
	client.UserGameInfo.Client = client
	client.UserGameInfo.SeatIndex = len(RoomMatch.Players)
	client.UserGameInfo.Score = RoomMatch.RoomMatchConfig.InitScore
	logs.Debug("初始配置数据 优先机器人使用")

	//初始配置数据 优先机器人使用
	client.UserGameInfo.Score = RoomMatch.RoomMatchConfig.InitScore
	client.UserGameInfo.GameConfig = &PlayerConfiguration{
		InitScore: RoomMatch.RoomMatchConfig.InitScore, //初始积分
		Power:     1,
		RoomMatch: RoomMatch,
	}

	logs.Debug("标记用户状态为房间中")
	//标记用户状态为房间中
	client.UserGameInfo.setOnline(2) // 标记用户状态为游戏中
	//广播给房间内的人 谁谁谁加入了
	RoomMatch.broadcast([]interface{}{"newJoin", map[string]interface{}{"userInfo": client.UserGameInfo}})
	//等待handFishInit完成
	if len(RoomMatch.Players) == RoomMatchNum { // 人满了 开始游戏！！！
		logs.Debug("人满了")
		RoomMatch.robotAddTicker.Stop()
		RoomMatchWait[RoomMatchNum] = nil // 弹出此等待的房间
		RoomMatch.Status = 1              // 标记为游戏中状态
		//client.UserGameInfo.setOnline(3)  // 标记用户状态为游戏中
		RoomMatch.broadcast([]interface{}{"message", map[string]interface{}{"message": "等待其他玩家资源加载就绪"}})
		go handRoomMatchRun(RoomMatch)
	}
}

// 比赛房的操作
func handRoomMatchRun(RoomMatch *RoomMatch) {
	RoomMatch.mutex.Lock()
	defer RoomMatch.mutex.Unlock()
	// 需要让玩家发送一个ready表示可以开始加载数据  todo:改为chan传输 不做循环处理
	logs.Debug("检测玩家准备中接收数据")
	logs.Debug(RoomMatch.AllPlayersReady)

	for !RoomMatch.AllPlayersReady {
		allReady := true
		for _, client := range RoomMatch.Players {
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
			RoomMatch.broadcast([]interface{}{"message", map[string]interface{}{"message": "开始推送"}})
			RoomMatch.AllPlayersReady = allReady
			//开始鱼数据实时状态和数据的发送
			//go handFishInit(RoomMatch)      //鱼群初始化数据
			//直接拆成四人组
			roomSubData := RoomMatch.intoSubRooms()
			chanSub := make(chan bool)
			for _, groupClient := range roomSubData {
				roomSub := &RoomMatchSub{
					RoomChan:        chanSub,
					Players:         groupClient,
					FishGroup:       nil,
					robotAddTicker:  RoomMatch.robotAddTicker,
					RoomMatchConfig: RoomMatch.RoomMatchConfig,
					Type:            "MatchSub",
					mutex:           sync.Mutex{},
					fishMutex:       sync.Mutex{},
				}
				go handMatchFishInit(roomSub)
				go handMatchFishRun(roomSub)
			}
			go closeRoomMatchTimer(RoomMatch, chanSub) //定义定时器外部关闭
			return
		}
		logs.Debug("检测准备")
		time.Sleep(time.Second * 2)
	}
}

// 拆分组
func (r *RoomMatch) intoSubRooms() [][]*Client {
	var groupedPlayers [][]*Client
	var currentGroup []*Client
	players := r.Players
	for _, player := range players {
		currentGroup = append(currentGroup, player)

		// 如果当前组达到四人或所有玩家遍历完成，将当前组加入分组列表
		if len(currentGroup) == 4 || player == players[len(players)-1] {
			groupedPlayers = append(groupedPlayers, currentGroup)
			currentGroup = []*Client{}
		}
	}
	return groupedPlayers
}

// 外部加锁，不能不加
func getOrCreateRoomMatch(RoomMatchNum int) (*RoomMatch, error) {
	logs.Debug("获取房间")
	room, exists := RoomMatchWait[RoomMatchNum]
	if !exists || room == nil {
		nextPkRoomMatchID++
		RoomMatchID := nextPkRoomMatchID
		//获取房间的配置
		RoomMatchConfig := getRoomMatchConfig(RoomMatchNum)
		if RoomMatchConfig == nil {
			message := fmt.Errorf("不存在的的房间或房间已关闭")
			logs.Debug(message)
			return nil, message
		}
		// 在创建房间时初始化 RoomMatchChan
		timeTick, _ := randomTicker(1, 5)
		room = &RoomMatch{
			ID:              RoomMatchID,
			Name:            RoomMatchName(RoomMatchNames[RoomMatchNum]),
			MaxPlayers:      RoomMatchNum,
			IsClose:         false,
			RoomChan:        make(chan bool),
			Players:         make(map[int]*Client),
			robotAddTicker:  timeTick,
			RoomMatchConfig: RoomMatchConfig,
			Type:            "Match",
		}
		RoomMatchs[RoomMatchID] = room
		RoomMatchWait[RoomMatchNum] = room
		// 启动定时添加机器人的协程
		go robotMatchRun(room)
	}
	logs.Debug("返回房间数据")
	return room, nil
}

func getRoomMatchConfig(num int) *common.RoomMatchInfo {
	RoomMatchConfig, err := model.GetMatchConfigFromRedis(num, "match")
	if err != nil {
		return nil
	}
	logs.Debug("getRoomMatchConfig的RoomMatchConfig：%v", RoomMatchConfig)
	return RoomMatchConfig

}

func closeRoomMatchTimer(RoomMatch *RoomMatch, chanSub chan bool) {
	fmt.Print("定时器开始了 \n")
	//定时器
	timer := time.NewTimer(time.Duration(RoomMatch.RoomMatchConfig.DurationMin) * time.Second)
	val := time.Duration(RoomMatch.RoomMatchConfig.DurationMin) * time.Minute
	logs.Debug("房间定时val %v", val)
	gameProgressSendTime := time.NewTimer(time.Minute)
	nowProgress := time.Duration(RoomMatch.RoomMatchConfig.DurationMin) * time.Minute
	//监听定时器管道和手动关闭管道
	for {
		select {
		case <-timer.C:
			logs.Debug("关闭RoomMatchChan")
			close(RoomMatch.RoomChan)
			close(chanSub)
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
			RoomMatch.broadcast(data)
		default:
		}
	}

}

// 房间结束
func (RoomMatch *RoomMatch) closeRoomMatch() {
	// 对所有玩家按照积分进行排序
	var players []*Client
	for _, player := range RoomMatch.Players {
		players = append(players, player)
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].UserGameInfo.Score > players[j].UserGameInfo.Score
	})
	//用户修改数据
	var usersToUpdate []model.UserCloseRoomSetData
	//RoomMatch.AllPlayersReady = false
	var balanceAdd int
	for i, client := range players {
		// 确定名次信息
		rankString := fmt.Sprintf("第%d名", i+1)
		// 计算获得积分
		score := fmt.Sprintf("无奖励")
		balanceAdd = 0

		if i == 0 {
			logs.Debug(client.UserGameInfo.Balance)
			client.UserGameInfo.Balance += RoomMatch.RoomMatchConfig.Place1Reward
			logs.Debug(client.UserGameInfo.Balance)
			score = fmt.Sprintf("奖励积分：%v", RoomMatch.RoomMatchConfig.Place1Reward)
			balanceAdd = int(RoomMatch.RoomMatchConfig.Place1Reward)
		}
		if i == 1 {
			logs.Debug(client.UserGameInfo.Balance)
			client.UserGameInfo.Balance += RoomMatch.RoomMatchConfig.Place2Reward
			logs.Debug(client.UserGameInfo.Balance)
			score = fmt.Sprintf("奖励积分：%v", RoomMatch.RoomMatchConfig.Place2Reward)
			balanceAdd = int(RoomMatch.RoomMatchConfig.Place2Reward)
		}
		if i == 2 {
			logs.Debug(client.UserGameInfo.Balance)
			client.UserGameInfo.Balance += RoomMatch.RoomMatchConfig.Place3Reward
			logs.Debug(client.UserGameInfo.Balance)
			score = fmt.Sprintf("奖励积分：%v", RoomMatch.RoomMatchConfig.Place3Reward)
			balanceAdd = int(RoomMatch.RoomMatchConfig.Place3Reward)
		}
		// 发送结算信息
		closeRoomMatch := []interface{}{"closePkRoomMatch", map[string]interface{}{
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
		client.broadcast(closeRoomMatch)
		client.RoomMatch = nil
	}
	model.SetUserCloseRoomData(usersToUpdate)
}

// 广播鱼群数据
func (RoomMatch *RoomMatch) broadcastFishLocation() {
	FishReady := []interface{}{"fishLocation", RoomMatch.FishGroup}
	// 发送给前端
	RoomMatch.broadcast(FishReady)
}

// 广播数据准备就绪消息
func (RoomMatch *RoomMatch) broadcastFishReady() {
	FishReady := []interface{}{"message",
		map[string]interface{}{
			"message": "鱼群就绪",
		}}
	// 发送给前端
	RoomMatch.broadcast(FishReady)
}

// 自定义广播
func (RoomMatch *RoomMatch) broadcast(data []interface{}) {
	if dataByte, err := json.Marshal(data); err != nil {
		logs.Error("broadcast [%v] json marshal err :%v ", data, err)
	} else {
		dataByte = append([]byte{'4', '2'}, dataByte...)
		for _, client := range RoomMatch.Players {
			if client.UserGameInfo.UserId > 0 && client.UserGameInfo.GroupId == 1 {
				client.sendMsg(dataByte)
			}
		}
	}
}
