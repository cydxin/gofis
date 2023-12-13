package common

import "time"

// 用来定义model的数据结构

// PkRecord 用户PK记录
type PkRecord struct {
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

// Response 通用响应结构体
type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
