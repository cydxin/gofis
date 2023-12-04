package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

var fishKinds = map[int]FishKind{
	0: {Speed: 11, Odds: 2, Name: "浪浪鱼"},
	1: {Speed: 10, Odds: 2, Name: "茂泽鱼"},
	2: {Speed: 8, Odds: 3, Name: "小兰鱼"},
	3: {Speed: 6, Odds: 5, Name: "佳丽鱼"},
}

type Fish struct {
	FishId      int  `json:"fish_id"`
	OffsetX     int  `json:"offset_x"`  // 鱼相对于之前X的位置
	OffsetY     int  `json:"offset_y"`  // 鱼相对于之前Y的位置
	CurrentX    int  `json:"CurrentX"`  // 鱼当前位置x
	CurrentY    int  `json:"CurrentY"`  // 鱼当前位置y
	FishKind    int  `json:"fish_kind"` //育种
	Hit         bool `json:"hit"`
	ToBeDeleted bool `json:"to_be_deleted"`
}

// 房间游戏前，或玩家加载游戏资源时就可以提前加载的方法
func handFishInit(room *Room) {
	fmt.Print("进入handFishInit \n")
	// 初始化 FishGroup 切片
	room.FishGroup = make([]*Fish, 50)

	// 为每个元素分配一个新的 Fish 对象
	for i := 0; i < 50; i++ {
		randomFishKindKey := rand.Intn(len(fishKinds))
		room.FishGroup[i] = &Fish{
			FishId:      i + 1,
			FishKind:    randomFishKindKey,
			OffsetX:     rand.Intn(maxY),
			OffsetY:     0,
			CurrentX:    rand.Intn(maxY),
			CurrentY:    0,
			ToBeDeleted: false,
		}
	}
}

// handFishrun 处理鱼的运动 也是主流程
func handFishrun(room *Room) {
	fmt.Print("鱼群开始 \n")

	//buildNormalFishTicker := time.NewTicker(time.Second * 20)          //刷普通鱼用定时器 TODO:鱼群 即奖励类鱼 圆阵 长方形的
	//flushTimeOutFishTicker := time.NewTicker(time.Second * 5)          //清理死鱼
	UpdateFishTracksTicker := time.NewTicker(10000 * time.Millisecond) //鱼足迹
	//defer buildNormalFishTicker.Stop()
	//defer flushTimeOutFishTicker.Stop()
	defer UpdateFishTracksTicker.Stop()

	for {
		select {
		case <-room.CloseChan:
			closeRoom(room)
			fmt.Printf("房价结束\n")
			return
		//case <-buildNormalFishTicker.C:
		//	fmt.Printf("驾驭\n")
		//
		//	addFish(room, 2)
		//case <-flushTimeOutFishTicker.C:
		//	fmt.Printf("删除鱼\n")
		//
		//	removeFish(room)
		case <-UpdateFishTracksTicker.C:
			// 循环处理所有鱼的位置更新
			fmt.Print("循环处理所有鱼的位置更新 \n")

			room.mutex.Lock()
			printFishGroupJSON(room)
			var fishData []interface{}
			for _, fish := range room.FishGroup {
				// 随机生成偏移量
				speed := fishKinds[fish.FishKind].Speed

				if fish.CurrentX > maxX || fish.CurrentY > maxY { //走出去的鱼不管了,同时删除
					fish.ToBeDeleted = true
					continue
				}
				fish.OffsetX = rand.Intn(speed + 1)
				fish.OffsetY = speed - fish.OffsetX
				fish.CurrentX += fish.OffsetX
				fish.CurrentY += fish.OffsetY
				// 将 Fish 对象转换为 JSON 格式的字节切片，并追加到结果切片中
				fishJSON, err := json.Marshal(fish)
				if err != nil {
					// 处理转换为 JSON 失败的情况
					fmt.Println("Error marshalling fish:", err)
					continue
				}
				fishData = append(fishData, fishJSON) //切片

			}
			room.mutex.Unlock()
			// 发送给房间内的玩家们
			room.sendMsgAllPlayer(fishData)
		default:
		}
	}
}

func removeFish(room *Room) {
	// 保留未标记为删除的元素
	room.mutex.Lock()
	defer room.mutex.Unlock()
	var newFishGroup []*Fish
	for _, fish := range room.FishGroup {
		if fish.ToBeDeleted {
			newFishGroup = append(newFishGroup, fish)
		}
	}
	room.FishGroup = newFishGroup
}
func addFish(room *Room, num int) {
	fmt.Print("添加鱼的数据 \n")
	for i := 0; i <= num; i++ {
		randomFishKindKey := rand.Intn(len(fishKinds))
		newFish := &Fish{
			FishId:      i + 1,
			FishKind:    randomFishKindKey,
			OffsetX:     rand.Intn(maxY),
			OffsetY:     0,
			CurrentX:    rand.Intn(maxY),
			CurrentY:    0,
			ToBeDeleted: false,
		}
		room.FishGroup = append(room.FishGroup, newFish)

		// 随机确定起始边界
		//startEdge := rand.Intn(4)
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
		//		CurrentY:    OffsetY,
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
