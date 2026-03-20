package common

import (
	"context"

	"gorm.io/gorm"
)

// IsAdmin 判断是不是超级管理员（通过 sys_role.is_admin 字段判断）
func IsAdmin(ctx context.Context, userId int64, db *gorm.DB) bool {
	sql := `select count(1) from sys_user_role sur
			left join sys_role sr on sur.role_id = sr.id
			where sur.user_id = ? and sr.is_admin = 1`
	var count int64

	db.WithContext(ctx).Raw(sql, userId).Scan(&count)
	return count > 0
}
