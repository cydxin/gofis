package common

import "time"

// 用来定义model的数据结构

// PageList 分页数据结构
type PageList struct {
	Count int64       `json:"count" db:"count"`
	List  interface{} `json:"list"`
}

// PkRecord 用户PK记录
type PkRecord struct {
	Count        int64 `json:"count"`
	List         interface{}
	ID           uint64    `json:"id" db:"id"`
	UserId       uint64    `json:"user_id" db:"user_id"`
	TicketAmount string    `json:"ticket_amount" db:"ticket_amount"`
	Ranking      string    `json:"ranking" db:"ranking"`
	PkNumber     string    `json:"pk_number" db:"pk_number"`
	Action       string    `json:"action" db:"action"`
	RoomName     string    `json:"room_name" db:"room_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PkRecord 用户PK记录
type MatchRecord struct {
	ID           uint64    `json:"id" db:"id"`
	UserId       uint64    `json:"user_id" db:"user_id"`
	TicketAmount string    `json:"ticket_amount" db:"ticket_amount"`
	Ranking      string    `json:"ranking" db:"ranking"`
	PkNumber     string    `json:"pk_number" db:"pk_number"`
	Action       string    `json:"action" db:"action"`
	RoomName     string    `json:"room_name" db:"room_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PkRecord 用户PK记录
type ExpRecord struct {
	ID           uint64    `json:"id" db:"id"`
	UserId       uint64    `json:"user_id" db:"user_id"`
	TicketAmount string    `json:"ticket_amount" db:"ticket_amount"`
	Ranking      string    `json:"ranking" db:"ranking"`
	PkNumber     string    `json:"pk_number" db:"pk_number"`
	Action       string    `json:"action" db:"action"`
	RoomName     string    `json:"room_name" db:"room_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// PkRoomInfo pkRoom PK房间配置
type PkRoomInfo struct {
	PkNumOfPeople int     `db:"pk_num_of_peo" json:"pk_num_of_peo"`
	RoomName      string  `db:"room_name" json:"room_name"`
	DurationMin   int     `db:"duration_min" json:"duration_min"`
	Ticket        float64 `db:"ticket" json:"ticket"`
	HandlingFee   float64 `db:"handling_fee" json:"handling_fee"`
	Money         float64 `db:"money" json:"money"`
	Turret        string  `db:"turret" json:"turret"`
	ExtWinRate    float64 `db:"ext_win_rate" json:"ext_win_rate"`
	InitScore     int     `db:"init_score" json:"init_score"`
}

// RoomMatchInfo MatchRoom Match房间配置
type RoomMatchInfo struct {
	Place1Reward float64 `db:"place1_reward" json:"place1_reward"`
	Place2Reward float64 `db:"place2_reward" json:"place2_reward"`
	Place3Reward float64 `db:"place3_reward" json:"place3_reward"`
	RoomName     string  `db:"room_name" json:"room_name"`
	DurationMin  int     `db:"duration_min" json:"duration_min"`
	Ticket       float64 `db:"ticket" json:"ticket"`
	Turret       string  `db:"turret" json:"turret"`
	ExtWinRate   float64 `db:"ext_win_rate" json:"ext_win_rate"`
	InitScore    int     `db:"init_score" json:"init_score"`
}

// Response 通用响应结构体
type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
