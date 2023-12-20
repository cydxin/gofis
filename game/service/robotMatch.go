package service

import (
	"github.com/astaxie/beego/logs"
	"time"
)

func robotMatchRun(room *RoomMatch) {
	logs.Debug("进入机器人检测")
	val := time.Duration(1)
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
			room.robotAddTicker, val = randomTicker(3, 6)
			logs.Debug("机器人下次定时长:%v", val)
			room.mutex.Unlock()
			addRobotToMatchRoom(room)
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
func addRobotToMatchRoom(room *RoomMatch) {
	UserGameInfo, err := generateRobotUserInfo()
	if err == nil {
		robotClient := &Client{
			UserGameInfo: UserGameInfo,
			RoomMatch:    room,
		}
		// 调用 EnterPkRoom 函数将机器人加入房间
		EnterRoomMatch(room.MaxPlayers, robotClient)
	}
}
func (room *RoomMatchSub) handRobotMatchRun() {
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
