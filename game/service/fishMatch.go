package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"math/rand"
	"time"
)

// 房间的游戏前，或玩家加载游戏资源时就可以提前加载的方法
func handMatchFishInit(room *RoomMatchSub) {
	room.fishMutex.Lock()
	fmt.Print("进入handFishInit \n")
	// 初始化 FishGroup 切片
	//room.FishGroup = make([]*Fish, 20)//便捷删除 不再使用切片
	room.FishGroup = make(map[FishId]*Fish)
	// 为每个元素分配一个新的 Fish 对象
	//todo:多边出鱼 参考addFish
	for i := 0; i < 20; i++ {
		randomFishKindKey := rand.Intn(len(fishKinds))
		currentY := rand.Intn(maxY)
		room.FishGroup[FishId(i+1)] = &Fish{
			FishId:      FishId(i + 1),
			FishKind:    randomFishKindKey,
			OffsetX:     0,
			OffsetY:     currentY,
			CurrentX:    0,
			CurrentY:    currentY,
			ToBeDeleted: false,
		}
	}
	room.fishMutex.Unlock()
}

// handFish run 处理鱼的运动 也是主流程
func handMatchFishRun(room *RoomMatchSub) {
	logs.Debug("鱼群开始")
	buildNormalFishTicker := time.NewTicker(time.Second * 10)         //加普通鱼用定时器 TODO:鱼群 即奖励类鱼 圆阵 长方形的
	flushTimeOutFishTicker := time.NewTicker(time.Second * 5)         //清理走出屏幕的鱼和被捕捉的鱼
	UpdateFishTracksTicker := time.NewTicker(5000 * time.Millisecond) //鱼足迹
	defer buildNormalFishTicker.Stop()
	defer flushTimeOutFishTicker.Stop()
	defer UpdateFishTracksTicker.Stop()
	go room.handRobotMatchRun()
	logs.Debug("开始for select")
	for {
		select {
		case msg := <-room.RoomChan:
			logs.Debug("子房间接受到管道关闭,msg:%v", msg)
			logs.Debug("控制权外交")
			return
		case <-buildNormalFishTicker.C:
			logs.Debug("handFishRun,buildNormalFishTicker")
			addMatchFish(room, 2)
		case <-flushTimeOutFishTicker.C:
			logs.Debug("handFishRun,flushTimeOutFishTicker")
			removeMatchFish(room)
		case <-UpdateFishTracksTicker.C:
			logs.Debug("handFishRun,UpdateFishTracksTicker")
			updateMatchFishTracks(room)
		default:
			//logs.Debug(room.RoomChan)
		}
	}
}

func updateMatchFishTracks(room *RoomMatchSub) {
	room.mutex.Lock()
	//printFishGroupJSON(room)
	for _, fish := range room.FishGroup {
		// 随机生成偏移量
		if fish.CurrentX > maxX || fish.CurrentY > maxY { //走出去的鱼不管了,同时标记
			fish.ToBeDeleted = true
		} else {
			speed := fishKinds[fish.FishKind].Speed
			fish.OffsetX = rand.Intn(speed + 1)
			fish.OffsetY = speed - fish.OffsetX
			fish.CurrentX += fish.OffsetX
			fish.CurrentY += fish.OffsetY
		}
	}
	// 发送给房间内的玩家们
	logs.Debug("发送鱼迹")
	room.mutex.Unlock()
	room.broadcastFishLocation()
	return
}

func removeMatchFish(room *RoomMatchSub) {
	// 保留未标记为删除的元素
	room.fishMutex.Lock()
	defer room.fishMutex.Unlock()
	for key, fish := range room.FishGroup {
		if fish.ToBeDeleted {
			delete(room.FishGroup, key)
		}
	}
}
func addMatchFish(room *RoomMatchSub, num int) {
	//fmt.Print("添加鱼的数据 \n")
	room.fishMutex.Lock()
	defer room.fishMutex.Unlock()
	startIndex := len(room.FishGroup) + 1
	for i := 0; i <= num; i++ {
		randomFishKindKey := rand.Intn(len(fishKinds))
		newFish := &Fish{
			FishId:      FishId(startIndex + i),
			FishKind:    randomFishKindKey,
			OffsetX:     rand.Intn(maxY),
			OffsetY:     0,
			CurrentX:    rand.Intn(maxY),
			CurrentY:    0,
			ToBeDeleted: false,
		}
		room.FishGroup[FishId(startIndex+i)] = newFish

		// 随机确定起始边界
		//startEdge := rand.Intn(4)
		//print(startEdge)
		//	y1080								y1080
		//	_____________________________________x 1920
		//	|									|
		//	|									|
		//	|									|
		//	|									|
		//	|									|
		//	|									|
		//	|									|
		//  0____________________________________ x 1920

		//switch startEdge {
		//case 0: // 上边
		//	newFish := &Fish{
		//		FishId:      i + 1,
		//		FishKind:    randomFishKindKey,
		//		OffsetX:     0,
		//		OffsetY:     0,
		//		CurrentX:    OffsetX,
		//		CurrentY:    maxY,
		//		ToBeDeleted: true,
		//	}
		//	room.FishGroup = append(room.FishGroup, newFish)
		//case 1: // 右边
		//	newFish := &Fish{
		//		FishId:      i + 1,
		//		FishKind:    randomFishKindKey,
		//		OffsetX:     maxX + OffsetX,
		//		OffsetY:     OffsetY,
		//		CurrentX:    maxX + OffsetX,
		//		CurrentY:    OffsetY,1
		//		ToBeDeleted: true,
		//	}
		//	room.FishGroup = append(room.FishGroup, newFish)
		//case 2: // 下边
		//	newFish := &Fish{
		//		FishId:      i + 1,
		//		FishKind:    randomFishKindKey,
		//		OffsetX:     OffsetX,
		//		OffsetY:     maxY + OffsetY,
		//		CurrentX:    OffsetX,
		//		CurrentY:    maxY + OffsetY,
		//		ToBeDeleted: true,
		//	}
		//	room.FishGroup = append(room.FishGroup, newFish)
		//case 3: // 左边
		//	newFish := &Fish{
		//		FishId:      i + 1,
		//		FishKind:    randomFishKindKey,
		//		OffsetX:     0 - OffsetX,
		//		OffsetY:     OffsetY,
		//		CurrentX:    0 - OffsetX,
		//		CurrentY:    OffsetY,
		//		ToBeDeleted: true,
		//	}
		//	room.FishGroup = append(room.FishGroup, newFish)
		//}
	}
}
