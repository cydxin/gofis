package model

import (
	"database/sql"
	"fmt"
	"gofish/common"
)

func GetPkRecordsThroughUsers(limit int, userId common.UserId) (*common.PageList, error) {
	var PkList common.PageList
	query := "SELECT * FROM users_pk_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11    OFFSET ?"
	queryCount := "SELECT count(*) as count FROM users_pk_log WHERE user_id = ?"
	errC := db.Get(&PkList.Count, queryCount, userId)
	if errC != nil {
		if errC == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		fmt.Printf("21 %v", errC)
		return nil, errC
	}
	fmt.Printf("PkList.List type: %t", PkList.List)
	var list []common.PkRecord
	err := db.Select(&list, query, userId, (limit-1)*10)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		// 其他数据库错误
		fmt.Printf("32 %v", err)
		return nil, err
	}
	PkList.List = list
	return &PkList, nil
}

func GetMatchRecordsThroughUsers(limit int, userId common.UserId) (*common.PageList, error) {
	var MatchList common.PageList
	query := "SELECT * FROM users_pk_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11    OFFSET ?"
	queryCount := "SELECT count(*) as count FROM users_pk_log WHERE user_id = ?"
	errC := db.Get(&MatchList.Count, queryCount, userId)
	if errC != nil {
		if errC == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		fmt.Printf("errC %v", errC)
		return nil, errC
	}
	var list []common.MatchRecord
	err := db.Select(&list, query, userId, (limit-1)*10)
	if err != nil {
		if err == sql.ErrNoRows {
			// 用户不存在
			return nil, fmt.Errorf("用户没有")
		}
		fmt.Printf("错误59 %v \n", err)
		// 其他数据库错误
		return nil, err
	}
	MatchList.List = list
	return &MatchList, nil
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
