package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const maxX = 1920
const maxY = 1080

// FishKind 结构体定义
type FishKind struct {
	Speed int    // 鱼的速度
	Odds  int    // 赔率
	Name  string // 鱼的名称
}

// 先不定义kindID 直接int
var fishKinds = map[int]FishKind{
	0: {Speed: 110, Odds: 2, Name: "浪浪鱼"},
	1: {Speed: 100, Odds: 2, Name: "茂泽鱼"},
	2: {Speed: 80, Odds: 3, Name: "小兰鱼"},
	3: {Speed: 60, Odds: 5, Name: "佳丽鱼"},
	4: {Speed: 60, Odds: 5, Name: "佳丽鱼"},
	5: {Speed: 60, Odds: 5, Name: "佳丽鱼"},
	6: {Speed: 60, Odds: 5, Name: "佳丽鱼"},
	7: {Speed: 60, Odds: 5, Name: "佳丽鱼"},
}

type FishId int

type Fish struct {
	FishId      FishId `json:"fish_id"`
	OffsetX     int    `json:"offset_x"`  // 鱼相对于之前X的位置
	OffsetY     int    `json:"offset_y"`  // 鱼相对于之前Y的位置
	CurrentX    int    `json:"CurrentX"`  // 鱼当前位置x
	CurrentY    int    `json:"CurrentY"`  // 鱼当前位置y
	FishKind    int    `json:"fish_kind"` // 育种
	Hit         bool   `json:"hit"`
	ToBeDeleted bool   `json:"to_be_deleted"`
	Mutex       sync.Mutex
}

// 房间的游戏前，或玩家加载游戏资源时就可以提前加载的方法
func handFishInit(room *Room) {
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
func handFishRun(room *Room) {
	fmt.Print("鱼群开始 \n")

	buildNormalFishTicker := time.NewTicker(time.Second * 20)        //加普通鱼用定时器 TODO:鱼群 即奖励类鱼 圆阵 长方形的
	flushTimeOutFishTicker := time.NewTicker(time.Second * 5)        //清理走出屏幕的鱼和被捕捉的鱼
	UpdateFishTracksTicker := time.NewTicker(500 * time.Millisecond) //鱼足迹
	defer buildNormalFishTicker.Stop()
	defer flushTimeOutFishTicker.Stop()
	defer UpdateFishTracksTicker.Stop()
	//关闭 > 清理 > 加鱼 > 鱼轨迹   后续可调整
	go room.handRobotRun()
	for {
		select {
		case <-room.CloseChan:
			closeRoom(room)
			fmt.Printf("房间结束\n")
			return
		case <-buildNormalFishTicker.C:
			//fmt.Printf("驾驭\n")
			addFish(room, 2)
		case <-flushTimeOutFishTicker.C:
			//fmt.Printf("删除鱼\n")
			removeFish(room)
		case <-UpdateFishTracksTicker.C:
			// 循环处理所有鱼的位置更新
			//fmt.Print("循环处理所有鱼的位置更新 \n")
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
			room.broadcastFishLocation()
			room.mutex.Unlock()
		default:
		}
	}
}

func removeFish(room *Room) {
	// 保留未标记为删除的元素
	room.fishMutex.Lock()
	defer room.fishMutex.Unlock()
	for key, fish := range room.FishGroup {
		if fish.ToBeDeleted {
			delete(room.FishGroup, key)
		}
	}
}

func addFish(room *Room, num int) {
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
func (f *Fish) hitFish(bulletId BulletId) bool {
	// 赔的分
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	odds := fishKinds[f.FishKind].Odds
	// 使用随机数范围控制捕获率
	i := odds / int(bulletId)
	a := rand.Intn(int(i) + 1)
	b := 0
	// 是否命中
	hitCondition := a == b && !f.Hit && !f.ToBeDeleted //被别人锁定 以及
	if hitCondition {
		f.Hit = true
		f.ToBeDeleted = true
		return true
	}
	return false
}
func (c *Client) catchFish(fishId FishId, bulletId BulletId) {
	//计算概率
	//已使用毫秒触发尝试，同时发送不会出现都返回true，考虑到实际情况更少，不做锁处理
	c.UserInfo.Score -= int(bulletId)
	if c.Room.FishGroup[fishId].hitFish(bulletId) {
		Score := fishKinds[c.Room.FishGroup[fishId].FishKind].Odds
		c.UserInfo.Score += Score
		catchResult := []interface{}{"catch_fish_reply",
			map[string]interface{}{
				"userId":   c.UserInfo.UserId,
				"integral": Score,
			}}
		c.Room.broadcast(catchResult)
	}
}
