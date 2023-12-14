package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"gofish/common"
	"net/http"
	"time"
)

const (
	writeWait      = 100 * time.Second
	pongWait       = 6000 * time.Second
	pingPeriod     = (pongWait * 900) / 10
	maxMessageSize = 512
)

// 过滤
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
var allClient = make(map[*Client]bool)

var hallBroadcast = make(chan []byte)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //直接true
}

// Client 定义连接的结构体
type Client struct {
	conn      *websocket.Conn
	UserInfo  *UserGameInfo
	Room      *Room
	inChan    chan []byte
	outChan   chan []byte
	IsReady   bool
	msgChan   chan []byte
	closeChan chan bool
	isClose   bool // 通道 closeChan 是否已经关闭
}

// PlayerConfiguration 玩家配置规范
type PlayerConfiguration struct {
	InitScore    int     `json:"-"`        //开局积分
	Power        float64 `json:"power"`    // 额外概率
	HitSpeed     float32 `json:"HitSpeed"` // 发送速度
	Room         *Room   `json:"-"`        // 发送速度
	LockFishType string  `json:"-"`        // 发送速度
}

// UserGameInfo 配置类型
type UserGameInfo struct {
	UserId      common.UserId        `json:"userId"`
	GroupId     int                  `json:"GroupId"`    //机器人标识
	SeatIndex   int                  `json:"seatIndex"`  // 座位，从左到右 从上到下 按进入房间的顺序给
	Score       int                  `json:"-"`          //对局时的积分
	BulletLevel int                  `json:"cannonKind"` // 子弹等级
	GameConfig  *PlayerConfiguration `json:"-"`          //对局时的配置 此数据不展示
	Balance     float64              `json:"balance"`    //余额
	Name        string               `json:"name"`
	NickName    string               `json:"nick_name"`
	Online      bool                 `json:"online"` // 离线
	Client      *Client              `json:"-"`
}

type BulletId int

type catchFishReq struct {
	BulletId BulletId `json:"bulletId"`
	FishId   FishId   `json:"fishId"`
}

func (c *Client) sendMsg(msg []byte) {
	if c.UserInfo != nil {
		//logs.Debug("user [%v] send msg %v", c.UserGameInfo.UserId, string(msg))
	}
	//fmt.Printf(" send msg %v\n", string(msg))
	//fmt.Printf(" send msg %v\n", msg)
	//TODO:加密消息
	c.msgChan <- msg
}

func (c *Client) writePump() {
	PingTicker := time.NewTicker(pingPeriod)
	defer func() {
		PingTicker.Stop()
		//close(c.closeChan)
		//RoomMgr.RoomLock.Lock()
		//defer RoomMgr.RoomLock.Unlock()
		//if c.Room != nil { //客户端在房间内
		//	if _, ok := RoomMgr.RoomsInfo[c.Room.RoomId]; ok { //房间没销毁
		//		c.Room.ClientReqChan <- &clientReqData{c, []string{"client_exit"}}
		//	}
		//}

		if c.UserInfo != nil {
			logs.Info("用户 %v writePump断开", c.UserInfo.UserId)
		}
	}()
	for {
		select {
		case msg := <-c.msgChan:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				logs.Error("sendMsg SetWriteDeadline err, %v", err)
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logs.Error("sendMsg NextWriter err, %v", err)
				return
			}
			if _, err = w.Write(msg); err != nil {
				logs.Error("sendMsg Write err, %v", err)
			}
			if err = w.Close(); err != nil {
				_ = c.conn.Close()
			}
		case <-PingTicker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logs.Debug("PingTicker write message err : %v", err)
				return
			}
		case <-c.closeChan:
			if err := c.conn.Close(); err != nil {
				if c.UserInfo != nil {
					logs.Info("user %v client conn close err : %v", err)
				}
			}
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() { // 协程结束后会执行的操作
		err := c.conn.Close()
		if err != nil {
			return
		}
		if c.UserInfo != nil {
			logs.Info("用户 %v readPump断开", c.UserInfo.UserId)
			c.removeFromClients()
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil { //存在错误状态
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { //意外关闭的状态下
				if c.UserInfo != nil { //如果用户是有登录，服务器没有异常的情况下就是 用户关闭
					logs.Error("websocket userId [%v] UserGameInfo [%d] 意外关闭错误: %v", c.UserInfo.UserId, &c.UserInfo, err)
				} else { //如果用户没有登录
					logs.Error("WebSocket 意外关闭错误: %v", err)
				}
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if message != nil {
			logs.Debug("读取conf.conf配置完成", string(message))
		}
		if err != nil {
			if c.UserInfo != nil {
				logs.Error("消息 unMarsha1 错误， user_id[%d] 错误:%v", c.UserInfo.UserId, err)
			} else {
				logs.Error("消息 unMarsha1 错误:%v", err)
			}
		} else {
			wsRequest(message, c)
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) //将其升级为WS请求
	if err != nil {
		logs.Error("upgrader 错误:%v", err)
		return
	}
	sid := r.Header.Get("Sec-Websocket-Key")

	client := &Client{conn: conn, msgChan: make(chan []byte, 100), closeChan: make(chan bool, 1), UserInfo: &UserGameInfo{}} //初始的客户端连接
	client.addToClients()
	logs.Debug("客户端可以连接")

	//创建读写 协程做之后的操作
	go client.readPump()  //读操作
	go client.writePump() //写操作

	if msg, err := json.Marshal(map[string]interface{}{
		"pingInterval": 25000,
		"pingTimeout":  5000,
		"sid":          sid, //不相关没事
		"upgrades":     make([]int, 0),
	}); err != nil {
		logs.Error("初始化的client错误 : %v", err)
	} else {
		//socket.io风格的初始数据
		client.sendMsg(append([]byte{'0'}, msg...))
		client.sendMsg(append([]byte{'4', '0'}))
	}
}

func (c *Client) addToClients() {
	allClient[c] = true
}

func (c *Client) removeFromClients() {
	if _, ok := allClient[c]; ok {
		delete(allClient, c)
	}
}

func HandleHallBroadcast() {
	for {
		select {
		case msg := <-hallBroadcast:
			for client := range allClient {
				client.sendMsg(msg)
			}
		}
	}
}

func (c *Client) ModelSuccess(Data interface{}) {
	// 封装响应数据
	response := common.Response{
		Status: "success",
		Data:   Data,
	}
	// 使用encoding/json包进行JSON序列化
	responseData, err := json.Marshal(response)
	if err != nil {
		sprintf := fmt.Sprintf("错误的JSON序列化：%v \n", err)
		c.sendMsg([]byte(sprintf))
		return
	}
	// 发送JSON数据给前端
	c.sendMsg(responseData)

}
