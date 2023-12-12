package model

import (
	"database/sql"
	"fmt"
	"gofish/common"
)

func GetPkRecordsThroughUsers(limit int, userId common.UserId) (*common.PkRecord, error) {
	var pkRecord common.PkRecord
	query := "SELECT * FROM users_pk_log WHERE user_id = ?  ORDER BY created_at DESC   LIMIT 11     OFFSET ?"

	// 使用全局数据库连接 db
	err := db.Get(&pkRecord, query, userId, (limit-1)*10)
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
