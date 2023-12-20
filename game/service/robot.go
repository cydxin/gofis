package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"gofish/model"
	"math/rand"
	"time"
)

func robotRun(room *RoomPk) {
	logs.Debug("进入机器人检测")
	for {
		select {
		case <-room.robotAddTicker.C:
			logs.Debug("进入机器人添加")
			room.mutex.Lock()
			//先判断人满了没 人满了就结束
			if room.MaxPlayers == len(room.Players) {
				logs.Debug("结束机器人添加")
				room.robotAddTicker.Stop()
				room.mutex.Unlock()
				return
			}
			logs.Debug("机器人添加")
			room.robotAddTicker.Stop()
			room.robotAddTicker, _ = randomTicker(3, 6)
			room.mutex.Unlock()
			addRobotToRoom(room)
			if room.MaxPlayers == len(room.Players) {
				logs.Debug("结束机器人添加")
				room.robotAddTicker.Stop()
				return
			}
		case <-room.RoomChan:
			logs.Debug("房间结束 机器人停止")
			return
		}
	}
}

// 添加机器人到房间
func addRobotToRoom(room *RoomPk) {
	UserGameInfo, err := generateRobotUserInfo()
	if err == nil {
		robotClient := &Client{
			UserGameInfo: UserGameInfo,
			Room:         room,
		}
		// 调用 EnterPkRoom 函数将机器人加入房间
		EnterPkRoom(room.MaxPlayers, room.RoomConfig.RoomName, robotClient)
	}

}

// 生成机器人用户信息
func generateRobotUserInfo() (*UserGameInfo, error) {
	// 实现根据需要生成机器人用户信息的逻辑
	userInfo, err := model.GetUserByRobot()
	if err != nil {
		logs.Error("提供机器人错误 %v", err)
		return nil, err
	}

	return &UserGameInfo{
		UserId:      common.UserId(userInfo.ID),
		GroupId:     userInfo.GroupID,
		SeatIndex:   0,
		Score:       0,
		BulletLevel: 1,
		GameConfig:  nil,
		Balance:     userInfo.Balance,
		Name:        userInfo.Account,
		NickName:    userInfo.Nickname,
		Online:      1,
		Client:      nil,
	}, nil
}

// 机器人的操作配置
func (room *RoomPk) handRobotRun() {
	//判断是否有机器人 有的话进行机器人操作
	robList := make([]*UserGameInfo, 0)
	for _, client := range room.Players {
		if client.UserGameInfo.GroupId == 2 {
			robList = append(robList, client.UserGameInfo)
		}
	}
	if len(robList) == 0 {
		return
	}
	// 启动每个机器人的操作goroutine
	for _, robot := range robList {
		go func(robot *UserGameInfo) {
			//初始化运行
			//切换炮台
			switchTurretTimer, val := randomTicker(time.Duration(10), time.Duration(30))
			logs.Debug(" switchTurretTimer val:%v", val)
			//切换炮速
			gunSpeedTimer, val := randomTicker(time.Duration(3), time.Duration(8))
			logs.Debug("gunSpeedTimer val:%v", val)

			logs.Debug("机器人初始化运行")

			switchTurret(robot, switchTurretTimer)
			switchTurretSpeed(robot, gunSpeedTimer)
			switchLockFish(robot)
			go performRobotAction(robot)
			logs.Debug("第一次操作的配置和执行机器人操作")
			// 执行机器人的操作逻辑
			for {
				select {
				case <-room.RoomChan:
					logs.Debug("机器人接收到房间关闭信号，结束运行")
					return
				case <-switchTurretTimer.C:
					switchTurret(robot, switchTurretTimer)
				case <-gunSpeedTimer.C:
					switchTurretSpeed(robot, gunSpeedTimer)
				default:
				}
			}
		}(robot)
	}
}

// /机器人的动作
func performRobotAction(robot *UserGameInfo) {
	//robot.GameConfig.RoomPk.FishGroup
	//射速 * 此次时间 等于需要循环的次数
	//shot := time.NewTicker(time.Second * )
	var bulletStartingPoint string
	var bulletEndPoint string
	//每射击多少次后 切换目标
	ranDomTimes := rand.Intn(5) + 3
	hitNum := 0
	select {
	case <-robot.GameConfig.Room.RoomChan:
		return
	default:
		hitSpeed := robot.GameConfig.HitSpeed
		hitNum++
		if ranDomTimes == hitNum {
			ranDomTimes = rand.Intn(5) + 3
			hitNum = 0
			switchLockFish(robot)
		}
		for _, fish := range robot.GameConfig.Room.FishGroup { //todo:随机一下 暂时不实装按鱼的赔率
			if !fish.ToBeDeleted {
				bulletStartingPoint = "0,100"
				bulletEndPoint = fmt.Sprintf("%d,%d", fish.CurrentX, fish.CurrentY)
			}
		}
		FireBulletsMess := []interface{}{"fire_bullets",
			map[string]interface{}{
				"userId":              robot.UserId,
				"bulletId":            robot.BulletLevel,
				"bulletStartingPoint": bulletStartingPoint,
				"bulletEndPoint":      bulletEndPoint,
			},
		}
		robot.GameConfig.Room.broadcast(FireBulletsMess)
		floatHitSpeed := 1 / hitSpeed * float32(time.Second)
		time.Sleep(time.Duration(floatHitSpeed))
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

func switchTurretSpeed(robot *UserGameInfo, gunSpeedTimer *time.Ticker) {
	//上限速度： 2次射击在一秒内
	probability := rand.Intn(100) + 1 // 概率
	interval := time.Duration(rand.Int63n(int64(8-3))) + 3*time.Second
	gunSpeedTimer.Reset(interval)

	switch {
	case probability <= 50: //50的概率 上限速度
		robot.GameConfig.HitSpeed = 2
	case probability <= 85: //51 - 85 35的概率随机射速
		robot.GameConfig.HitSpeed = rand.Float32()*(2.0-0.1) + 0.1
	default:
		robot.GameConfig.HitSpeed = float32(1 / interval)
	}

}
func switchTurret(robot *UserGameInfo, switchTurretTimer *time.Ticker) {
	probability := rand.Intn(100) + 1 // 概率
	//logs.Debug(robot)
	interval := time.Duration(rand.Int63n(int64(30-10))) + 10*time.Second
	switchTurretTimer.Reset(interval)
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
