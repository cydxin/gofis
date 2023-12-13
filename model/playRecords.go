package model

import (
	"database/sql"
	"fmt"
	"gofish/common"
)

func GetPkRecordsThroughUsers(limit int, userId common.UserId) (*common.PkRecord, error) {
	var pkRecord common.PkRecord
	query := "SELECT * FROM users_pk_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11    OFFSET ?"
	//query_count := "SELECT count(*) FROM users_pk_log WHERE user_id = ?"
	err := db.Get(&pkRecord, query, userId, (limit-1)*10)
	//err := db.Get(&pkRecord, query_count, userId, (limit-1)*10)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		return nil, err
	}

	return &pkRecord, nil
}

func GetMatchRecordsThroughUsers(limit int, userId common.UserId) (*common.MatchRecord, error) {
	var matchRecord common.MatchRecord
	query := "SELECT * FROM users_match_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11    OFFSET ?"

	// 使用全局数据库连接 db
	err := db.Get(&matchRecord, query, userId, (limit-1)*10)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		return nil, err
	}

	return &matchRecord, nil
}

func GetExpRecordsThroughUsers(limit int, userId common.UserId) (*common.ExpRecord, error) {
	var expRecord common.ExpRecord
	query := "SELECT * FROM users_match_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11    OFFSET ?"

	// 使用全局数据库连接 db
	err := db.Get(&expRecord, query, userId, (limit-1)*10)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		return nil, err
	}

	return &expRecord, nil
}
