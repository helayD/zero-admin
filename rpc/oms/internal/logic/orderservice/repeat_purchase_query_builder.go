package orderservicelogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"gorm.io/gorm"
)

type repeatPurchaseFilter struct {
	Scope        pkgscope.GovernanceScope
	StartTime    time.Time
	EndTime      time.Time
	Channel      string
	ActivityType string
	ActivityID   int64
}

type repeatPurchaseQualifiedOrderRow struct {
	OrderID       int64     `gorm:"column:order_id"`
	MemberID      int64     `gorm:"column:member_id"`
	PayTime       time.Time `gorm:"column:pay_time"`
	PayAmount     float64   `gorm:"column:pay_amount"`
	SourceType    int32     `gorm:"column:source_type"`
	ActivityType  string    `gorm:"column:activity_type"`
	ActivityID    int64     `gorm:"column:activity_id"`
	PlatformID    int64     `gorm:"column:platform_id"`
	TenantID      int64     `gorm:"column:tenant_id"`
	MerchantID    int64     `gorm:"column:merchant_id"`
}

type repeatPurchaseQueryBuilder struct {
	ctx context.Context
	db  *gorm.DB
}

func newRepeatPurchaseQueryBuilder(ctx context.Context, svcCtx *svc.ServiceContext) *repeatPurchaseQueryBuilder {
	return &repeatPurchaseQueryBuilder{
		ctx: ctx,
		db:  svcCtx.DB,
	}
}

func buildRepeatPurchaseQualifiedOrderQuery(ctx context.Context, db *gorm.DB, filter repeatPurchaseFilter) (*gorm.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库连接不能为空")
	}
	if _, err := operatefunnel.RepeatPurchaseScopeOrderColumns(filter.Scope); err != nil {
		return nil, err
	}
	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.StartTime.Before(filter.EndTime) {
		return nil, fmt.Errorf("复购分析时间范围非法")
	}

	priorExistsSQL, err := buildRepeatPurchasePriorOrderExistsSQL(filter.Scope, "curr", "prev")
	if err != nil {
		return nil, err
	}

	query := db.WithContext(ctx).
		Table("oms_order_main curr").
		Select(strings.Join([]string{
			"curr.id AS order_id",
			"curr.user_id AS member_id",
			"curr.pay_time",
			"curr.pay_amount",
			"curr.source_type",
			"curr.activity_type",
			"curr.activity_id",
			"curr.platform_id",
			"curr.tenant_id",
			"curr.merchant_id",
		}, ", "))
	query = applyRepeatPurchaseCurrentOrderFilters(query, filter, "curr")
	query = query.Where("EXISTS (" + priorExistsSQL + ")")
	return query, nil
}

func applyRepeatPurchaseCurrentOrderFilters(query *gorm.DB, filter repeatPurchaseFilter, alias string) *gorm.DB {
	scopeSQL, scopeArgs := pkgscope.ScopeFilterSQL(alias, filter.Scope)
	query = query.Where(alias + ".is_deleted = ?", 0)
	query = query.Where(alias + ".pay_time IS NOT NULL")
	query = query.Where(alias+".order_status IN ?", [][]int32{{2, 3, 4, 7}}[0])
	query = query.Where(alias+".pay_time >= ? AND "+alias+".pay_time < ?", filter.StartTime, filter.EndTime)
	query = query.Where(scopeSQL, scopeArgs...)

	if channel, ok := operatefunnel.OptionalChannel(filter.Channel); ok {
		query = query.Where(repeatPurchaseOrderChannelCaseSQL(alias)+" = ?", channel)
	}
	if activityType, ok := operatefunnel.OptionalActivityType(filter.ActivityType); ok {
		if activityType == operatefunnel.ActivityNone {
			query = query.Where("COALESCE(NULLIF("+alias+".activity_type, ''), 'none') = 'none'")
		} else {
			query = query.Where(alias+".activity_type = ?", activityType)
			if filter.ActivityID > 0 {
				query = query.Where(alias+".activity_id = ?", filter.ActivityID)
			}
		}
	}
	return query
}

func buildRepeatPurchasePriorOrderExistsSQL(scope pkgscope.GovernanceScope, currentAlias, previousAlias string) (string, error) {
	columns, err := operatefunnel.RepeatPurchaseScopeOrderColumns(scope)
	if err != nil {
		return "", err
	}
	clauses := []string{
		"SELECT 1 FROM oms_order_main " + previousAlias,
		"WHERE " + previousAlias + ".user_id = " + currentAlias + ".user_id",
		"AND " + previousAlias + ".is_deleted = 0",
		"AND " + previousAlias + ".pay_time IS NOT NULL",
		"AND " + previousAlias + ".order_status IN (2,3,4,7)",
		"AND " + previousAlias + ".pay_time >= DATE_SUB(" + currentAlias + ".pay_time, INTERVAL 180 DAY)",
		"AND " + previousAlias + ".pay_time < " + currentAlias + ".pay_time",
	}
	for _, column := range columns {
		clauses = append(clauses, "AND "+previousAlias+"."+column+" = "+currentAlias+"."+column)
	}
	return strings.Join(clauses, " "), nil
}

func repeatPurchaseOrderChannelCaseSQL(alias string) string {
	return "CASE " + alias + ".source_type WHEN 1 THEN 'app' WHEN 2 THEN 'pc' WHEN 3 THEN 'mini_program' ELSE 'unknown' END"
}
