package model

import (
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v4"
	"log"
	"time"
)

// UserInfo 表示用户信息
type UserInfo struct {
	ID                     uint64    `json:"id" db:"id"`
	GroupID                int       `json:"group_id" db:"group_id"`
	Account                string    `json:"account" db:"account"`
	Nickname               string    `json:"nickname" db:"nickname"`
	Password               string    `json:"password" db:"password"`
	Phone                  string    `json:"phone" db:"phone"`
	Avatar                 string    `json:"avatar" db:"avatar"`
	Gender                 int       `json:"gender" db:"gender"`
	Balance                float64   `json:"balance" db:"balance"`
	MatchFrequency         int       `json:"match_frequency" db:"match_frequency"`
	PKFrequency            int       `json:"pk_frequency" db:"pk_frequency"`
	MatchFee               int       `json:"match_fee" db:"match_fee"`
	PKFee                  int       `json:"pk_fee" db:"pk_fee"`
	MatchMoney             int       `json:"match_money" db:"match_money"`
	PKMoney                int       `json:"pk_money" db:"pk_money"`
	ExperienceTotal        int       `json:"experience_total" db:"experience_total"`
	WinningRate20          int       `json:"winning_rate_20" db:"winning_rate_20"`
	WinningRate50          int       `json:"winning_rate_50" db:"winning_rate_50"`
	WinningRate100         int       `json:"winning_rate_100" db:"winning_rate_100"`
	FirstParticipationTime time.Time `json:"first_participation_time" db:"first_participation_time"`
	LastParticipationTime  time.Time `json:"last_participation_time" db:"last_participation_time"`
	LoginDay               int       `json:"login_day" db:"login_day"`
	MaxLoginDay            int       `json:"max_login_day" db:"max_login_day"`
	ParticipationDay       int       `json:"participation_day" db:"participation_day"`
	MaxParticipationDay    int       `json:"max_participation_day" db:"max_participation_day"`
	Status                 int       `json:"status" db:"status"`
	IsOnline               int       `json:"online_status" db:"online_status"`
	DeletedAt              null.Time `json:"deleted_at" db:"deleted_at"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type UserCloseRoomSetData struct {
	ID       common.UserId `db:"id"`
	IsOnline int           `db:"online_status"`
	PKMoney  int           `db:"pk_money"`
	Balance  float64       `db:"balance"`
}

func GetUserByCredentials(username, password string) (*UserInfo, error) {
	var userInfo UserInfo
	query := "SELECT * FROM users WHERE phone = ?  and group_id = 1 LIMIT 1 "

	// 使用全局数据库连接 db
	err := db.Get(&userInfo, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		return nil, err
	}

	// 在需要验证的地方
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	return &userInfo, nil
}
func GetUserByRobot() (*UserInfo, error) {
	var userInfo UserInfo
	query := "SELECT * FROM users WHERE group_id = 2 AND online_status < 2 LIMIT 1 "

	// 使用全局数据库连接 db
	err := db.Get(&userInfo, query)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		return nil, err
	}
	return &userInfo, nil
}
func SetUserOnline(u interface{}) {
	// 要更新的用户信息
	// 执行UPDATE语句

	query := "UPDATE users SET online_status = :online_status  WHERE id = :userId"
	_, err := db.NamedExec(query, u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Update successful")
}

func SetUserOfflineAll() {
	// 要更新的用户信息
	// 执行UPDATE语句

	query := "UPDATE users SET online_status = 0"
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Update successful")
}
func SetUserCloseRoomData(u []UserCloseRoomSetData) {
	for _, user := range u {
		query := "UPDATE users SET online_status = :online_status, pk_money = pk_money + :pk_money, balance = :balance WHERE id = :id"
		logs.Debug(user)
		_, err := db.NamedExec(query, user)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Update successful")
}
