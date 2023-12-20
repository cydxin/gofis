package model

import (
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/logs"
	"gofish/common"
)

func GetPkRoom() ([]*common.PkRoomInfo, error) {
	var PkRoomInfo []*common.PkRoomInfo
	query := "SELECT pk_num_of_peo, room_name, duration_min, ticket, handling_fee, money, turret," +
		"ext_win_rate, init_score FROM pk_room WHERE status = 1"
	err := db.Select(&PkRoomInfo, query)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}

		// 其他数据库错误
		logs.Debug("GetPkRoom: %v \n", err)
		return nil, err
	}
	return PkRoomInfo, nil
}

func GetMatchRoom() ([]*common.RoomMatchInfo, error) {
	var RoomMatchInfo []*common.RoomMatchInfo
	query := "SELECT place1_reward,place2_reward,place3_reward, room_name, duration_min, ticket ,  turret," +
		"ext_win_rate, init_score FROM match_room WHERE status = 1"
	err := db.Select(&RoomMatchInfo, query)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}

		// 其他数据库错误
		logs.Debug("GetPkRoom: %v \n", err)
		return nil, err
	}
	return RoomMatchInfo, nil
}
