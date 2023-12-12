package model

import "github.com/jmoiron/sqlx"

var db *sqlx.DB

// InitDB 初始化全局数据库连接
func InitDB(database *sqlx.DB) {
	db = database
}
