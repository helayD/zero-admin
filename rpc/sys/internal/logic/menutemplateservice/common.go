package menutemplateservicelogic

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/time_util"
	"gorm.io/gorm"
)

type menuTemplateRow struct {
	ID         int64      `gorm:"column:id"`
	Name       string     `gorm:"column:name"`
	ScopeType  string     `gorm:"column:scope_type"`
	PlatformID int64      `gorm:"column:platform_id"`
	Status     int32      `gorm:"column:status"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
}

func (menuTemplateRow) TableName() string {
	return "sys_menu_template"
}

type menuTemplateListRow struct {
	ID         int64      `gorm:"column:id"`
	Name       string     `gorm:"column:name"`
	ScopeType  string     `gorm:"column:scope_type"`
	PlatformID int64      `gorm:"column:platform_id"`
	Status     int32      `gorm:"column:status"`
	Remark     string     `gorm:"column:remark"`
	CreateBy   string     `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   string     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	MenuCount  int64      `gorm:"column:menu_count"`
}

type menuTemplateItemRow struct {
	ID         int64 `gorm:"column:id"`
	TemplateID int64 `gorm:"column:template_id"`
	MenuID     int64 `gorm:"column:menu_id"`
}

func (menuTemplateItemRow) TableName() string {
	return "sys_menu_template_item"
}

func normalizeTemplateScope(scopeType string) (string, error) {
	scopeType = strings.TrimSpace(scopeType)
	switch scopeType {
	case "tenant", "merchant":
		return scopeType, nil
	default:
		return "", errors.New("菜单模板作用域仅支持 tenant 或 merchant")
	}
}

func normalizeTemplateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("模板名称不能为空")
	}

	return name, nil
}

func normalizePlatformID(platformID int64) int64 {
	if platformID <= 0 {
		return 1
	}

	return platformID
}

func normalizePage(pageNum, pageSize int64) (int64, int64) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	return pageNum, pageSize
}

func normalizeMenuIDs(menuIDs []int64) []int64 {
	result := make([]int64, 0, len(menuIDs))
	seen := make(map[int64]struct{}, len(menuIDs))

	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		if _, ok := seen[menuID]; ok {
			continue
		}
		seen[menuID] = struct{}{}
		result = append(result, menuID)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func formatNullableTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return time_util.TimeToString(value)
}

func loadTemplateMenuIDs(ctx context.Context, db *gorm.DB, templateID int64) ([]int64, error) {
	menuIDs := make([]int64, 0)
	err := db.WithContext(ctx).
		Table("sys_menu_template_item").
		Distinct("menu_id").
		Where("template_id = ?", templateID).
		Order("menu_id ASC").
		Pluck("menu_id", &menuIDs).Error
	if err != nil {
		return nil, err
	}

	return menuIDs, nil
}

func replaceTemplateMenuIDs(ctx context.Context, tx *gorm.DB, templateID int64, menuIDs []int64) error {
	if err := tx.WithContext(ctx).
		Table("sys_menu_template_item").
		Where("template_id = ?", templateID).
		Delete(&menuTemplateItemRow{}).Error; err != nil {
		return err
	}

	if len(menuIDs) == 0 {
		return nil
	}

	rows := make([]menuTemplateItemRow, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		rows = append(rows, menuTemplateItemRow{
			TemplateID: templateID,
			MenuID:     menuID,
		})
	}

	return tx.WithContext(ctx).Table("sys_menu_template_item").Create(&rows).Error
}
