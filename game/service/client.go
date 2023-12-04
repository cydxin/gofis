package service

import (
	"bytes"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	writeWait      = 100 * time.Second
	pongWait       = 6000 * time.Second
	pingPeriod     = (pongWait * 900) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //不验证origin
}

// 定义连接的结构体
type Client struct {
	conn      *websocket.Conn
	UserInfo  *UserInfo
	inChan    chan []byte
	outChan   chan []byte
	IsReady   bool
	msgChan   chan []byte
	closeChan chan bool
	isClose   bool // 通道 closeChan 是否已经关闭
}

type UserId int64

type UserInfo struct {
	UserId      UserId  `json:"userId"`
	Score       int     `json:"-"`       //对局时的积分
	Balance     float64 `json:"balance"` //余额
	Name        string  `json:"name"`
	NickName    string  `json:"nick_name"`
	Ready       bool    `json:"ready"`      // 游戏是否准备好，即游戏加载ok，所有room内的userinfo的都true后，开始给捕鱼数据  使用client的
	SeatIndex   int     `json:"seatIndex"`  // 座位，从左到右 从上到下 按进入房间顺序给
	BulletLevel int     `json:"cannonKind"` // 子弹等级
	Power       float64 `json:"power"`      // 额外概率
	Online      bool    `json:"online"`     // 离线
	//client      *Client `json:"-"`
}

type BulletId string

//	type Bullet struct {
//		UserId     UserId   `json:"userId"`
//		ChairId    int      `json:"chairId"`
//		BulletKind int      `json:"bulletKind"`
//		BulletId   BulletId `json:"bulletId"`
//		Angle      float64  `json:"angle"`
//		Sign       string   `json:"sign"`
//		LockFishId FishId   `json:"lockFishId"`
//	}
type catchFishReq struct {
	BulletId BulletId `json:"bulletId"`
	FishId   int      `json:"fishId"`
}

func (c *Client) sendMsg(msg []byte) {
	if c.UserInfo != nil {
		//logs.Debug("user [%v] send msg %v", c.UserInfo.UserId, string(msg))
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
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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
		c.conn.Close()
		if c.UserInfo != nil {
			logs.Info("用户 %v readPump断开", c.UserInfo.UserId)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil { //存在错误状态
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { //意外关闭的状态下
				if c.UserInfo != nil { //如果用户是有登录，服务器没有异常的情况下就是 用户关闭
					logs.Error("websocket userId [%v] UserInfo [%d] 意外关闭错误: %v", c.UserInfo.UserId, &c.UserInfo, err)
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

	client := &Client{conn: conn, msgChan: make(chan []byte, 100), closeChan: make(chan bool, 1), UserInfo: &UserInfo{}} //初始的客户端连接
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
		logs.Error("初始化的clent错误 : %v", err)
	} else {
		//socket.io风格的初始数据
		client.sendMsg(append([]byte{'0'}, msg...))
		client.sendMsg(append([]byte{'4', '0'}))
	}
}
