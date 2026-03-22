package auditcenterservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"gorm.io/gorm"
)

const (
	sourceLogin    = "login_log"
	sourceOperate  = "operate_log"
	sourceSecurity = "security_event"
	sourceTenant   = "tenant_audit"
	sourceMerchant = "merchant_audit"
	maxAuditWindow = 180 * 24 * time.Hour
)

type auditCenterItem struct {
	SourceType      string
	SourceID        int64
	TraceID         string
	EventType       string
	Action          string
	Result          string
	ScopeType       string
	ScopeLabel      string
	PlatformID      int64
	TenantID        int64
	MerchantID      int64
	OperatorID      int64
	OperatorName    string
	ResourceType    string
	ResourceID      int64
	ResourceName    string
	SubjectInfo     string
	RequestSummary  string
	DetailPayload   string
	SensitiveMasked bool
	HappenedAt      time.Time
}

type auditQueryWindow struct {
	start time.Time
	end   time.Time
}

type securityEventRow struct {
	ID             int64     `gorm:"column:id"`
	TraceID        string    `gorm:"column:trace_id"`
	EventType      string    `gorm:"column:event_type"`
	Action         string    `gorm:"column:action"`
	ResourceType   string    `gorm:"column:resource_type"`
	ResourceID     int64     `gorm:"column:resource_id"`
	ScopeType      string    `gorm:"column:scope_type"`
	PlatformID     int64     `gorm:"column:platform_id"`
	TenantID       int64     `gorm:"column:tenant_id"`
	MerchantID     int64     `gorm:"column:merchant_id"`
	OperatorID     int64     `gorm:"column:operator_id"`
	OperatorName   string    `gorm:"column:operator_name"`
	RequestSummary string    `gorm:"column:request_summary"`
	Result         string    `gorm:"column:result"`
	Payload        string    `gorm:"column:payload"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

type loginLogRow struct {
	ID            int64     `gorm:"column:id"`
	LoginName     string    `gorm:"column:login_name"`
	Ipaddr        string    `gorm:"column:ipaddr"`
	LoginLocation string    `gorm:"column:login_location"`
	Status        int32     `gorm:"column:status"`
	Msg           string    `gorm:"column:msg"`
	LoginTime     time.Time `gorm:"column:login_time"`
	PlatformID    int64     `gorm:"column:platform_id"`
	TenantID      int64     `gorm:"column:tenant_id"`
	MerchantID    int64     `gorm:"column:merchant_id"`
	UserID        int64     `gorm:"column:user_id"`
}

type operateLogRow struct {
	ID              int64     `gorm:"column:id"`
	Title           string    `gorm:"column:title"`
	BusinessType    int32     `gorm:"column:business_type"`
	Method          string    `gorm:"column:method"`
	RequestMethod   string    `gorm:"column:request_method"`
	OperateName     string    `gorm:"column:operate_name"`
	OperateURL      string    `gorm:"column:operate_url"`
	OperateIP       string    `gorm:"column:operate_ip"`
	OperateLocation string    `gorm:"column:operate_location"`
	OperateParam    string    `gorm:"column:operate_param"`
	JSONResult      string    `gorm:"column:json_result"`
	Extra           string    `gorm:"column:extra"`
	Status          int32     `gorm:"column:status"`
	ErrorMsg        string    `gorm:"column:error_msg"`
	OperateTime     time.Time `gorm:"column:operate_time"`
}

type governanceExtra struct {
	TraceID        string `json:"traceId"`
	Action         string `json:"action"`
	ResourceType   string `json:"resourceType"`
	ResourceID     int64  `json:"resourceId"`
	ResourceName   string `json:"resourceName"`
	ScopeType      string `json:"scopeType"`
	PlatformID     int64  `json:"platformId"`
	TenantID       int64  `json:"tenantId"`
	MerchantID     int64  `json:"merchantId"`
	ScopeLabel     string `json:"scopeLabel"`
	OperatorID     int64  `json:"operatorId"`
	OperatorName   string `json:"operatorName"`
	Result         string `json:"result"`
	RequestSummary string `json:"requestSummary"`
}

func parseWindow(start, end string) (auditQueryWindow, error) {
	now := time.Now()
	window := auditQueryWindow{end: now, start: now.Add(-7 * 24 * time.Hour)}
	if strings.TrimSpace(start) != "" {
		v, err := parseAuditTime(start)
		if err != nil {
			return auditQueryWindow{}, err
		}
		window.start = v
	}
	if strings.TrimSpace(end) != "" {
		v, err := parseAuditTime(end)
		if err != nil {
			return auditQueryWindow{}, err
		}
		window.end = v
	}
	if window.end.Before(window.start) {
		return auditQueryWindow{}, errors.New("结束时间不能早于开始时间")
	}
	if window.end.Sub(window.start) > maxAuditWindow {
		return auditQueryWindow{}, errors.New("查询时间窗口不能超过180天")
	}
	return window, nil
}

func parseAuditTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{time.DateTime, "2006-01-02 15:04", time.RFC3339, time.DateOnly}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("非法时间格式: %s", value)
}

func queryAuditItems(ctx context.Context, svcCtx *svc.ServiceContext, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	scopeValue, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	window, err := parseWindow(in.StartTime, in.EndTime)
	if err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, 64)
	appenders := []func(context.Context, *svc.ServiceContext, pkgscope.GovernanceScope, auditQueryWindow, *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error){
		queryLoginItems,
		queryOperateItems,
		querySecurityItems,
		queryTenantItems,
		queryMerchantItems,
	}
	for _, fn := range appenders {
		rows, err := fn(ctx, svcCtx, scopeValue, window, in)
		if err != nil {
			return nil, err
		}
		items = append(items, rows...)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].HappenedAt.After(items[j].HappenedAt) })
	return items, nil
}

func queryLoginItems(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, window auditQueryWindow, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	q := svcCtx.DB.WithContext(ctx).Table("sys_login_log l").
		Select("l.id, l.login_name, l.ipaddr, l.login_location, l.status, l.msg, l.login_time, u.platform_id, u.tenant_id, u.merchant_id, u.id as user_id").
		Joins("left join sys_user u on u.user_name = l.login_name").
		Where("l.login_time >= ? and l.login_time <= ?", window.start, window.end)
	q = applyScopeWhere(q, current, "u")
	if in.OperatorName != "" {
		q = q.Where("l.login_name like ?", likeValue(in.OperatorName))
	}
	if in.Keyword != "" {
		kw := likeValue(in.Keyword)
		q = q.Where("l.login_name like ? or l.ipaddr like ? or l.msg like ?", kw, kw, kw)
	}
	if in.Result != "" {
		status := int32(1)
		if strings.Contains(strings.ToLower(in.Result), "fail") || strings.Contains(in.Result, "den") || strings.Contains(in.Result, "block") || strings.Contains(in.Result, "失败") {
			status = 0
		}
		q = q.Where("l.status = ?", status)
	}
	var rows []loginLogRow
	if err := q.Order("l.login_time desc").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, len(rows))
	for _, row := range rows {
		scopeValue := logiccommon.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)
		item := auditCenterItem{
			SourceType:      sourceLogin,
			SourceID:        row.ID,
			EventType:       "login",
			Action:          "user.login",
			Result:          mapResultFromStatus(row.Status, true),
			ScopeType:       scopeValue.ScopeType,
			ScopeLabel:      scopeValue.Label(),
			PlatformID:      scopeValue.PlatformID,
			TenantID:        scopeValue.TenantID,
			MerchantID:      scopeValue.MerchantID,
			OperatorID:      row.UserID,
			OperatorName:    row.LoginName,
			ResourceType:    "user",
			ResourceID:      row.UserID,
			SubjectInfo:     scopeValue.Label(),
			RequestSummary:  maskText(fmt.Sprintf("IP:%s 地点:%s %s", row.Ipaddr, row.LoginLocation, row.Msg), current),
			DetailPayload:   maskText(fmt.Sprintf("ip=%s\nlocation=%s\nmessage=%s", row.Ipaddr, row.LoginLocation, row.Msg), current),
			SensitiveMasked: current.ScopeType != pkgscope.SubjectTypePlatform,
			HappenedAt:      row.LoginTime,
		}
		if allowAuditItem(item, current, in) {
			items = append(items, item)
		}
	}
	return items, nil
}

func queryOperateItems(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, window auditQueryWindow, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	q := svcCtx.DB.WithContext(ctx).Table("sys_operate_log").Where("operate_time >= ? and operate_time <= ?", window.start, window.end)
	if in.OperatorName != "" {
		q = q.Where("operate_name like ?", likeValue(in.OperatorName))
	}
	if in.Keyword != "" {
		kw := likeValue(in.Keyword)
		q = q.Where("title like ? or operate_url like ? or error_msg like ?", kw, kw, kw)
	}
	var rows []operateLogRow
	if err := q.Order("operate_time desc").Limit(300).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, len(rows))
	for _, row := range rows {
		extra := parseGovernanceExtra(row.Extra)
		if extra.ScopeType == "" {
			continue
		}
		item := auditCenterItem{
			SourceType:      sourceOperate,
			SourceID:        row.ID,
			TraceID:         extra.TraceID,
			EventType:       chooseFirst(extra.Action, strconv.Itoa(int(row.BusinessType))),
			Action:          chooseFirst(extra.Action, row.Method),
			Result:          chooseFirst(extra.Result, mapResultFromStatus(row.Status, false)),
			ScopeType:       extra.ScopeType,
			ScopeLabel:      chooseFirst(extra.ScopeLabel, logiccommon.DefaultScope(extra.PlatformID, extra.TenantID, extra.MerchantID).Label()),
			PlatformID:      extra.PlatformID,
			TenantID:        extra.TenantID,
			MerchantID:      extra.MerchantID,
			OperatorID:      extra.OperatorID,
			OperatorName:    chooseFirst(extra.OperatorName, row.OperateName),
			ResourceType:    extra.ResourceType,
			ResourceID:      extra.ResourceID,
			ResourceName:    extra.ResourceName,
			SubjectInfo:     chooseFirst(extra.ScopeLabel, logiccommon.DefaultScope(extra.PlatformID, extra.TenantID, extra.MerchantID).Label()),
			RequestSummary:  maskText(chooseFirst(extra.RequestSummary, row.OperateURL), current),
			DetailPayload:   maskText(fmt.Sprintf("url=%s\nparam=%s\nresult=%s\nerror=%s\nextra=%s", row.OperateURL, row.OperateParam, row.JSONResult, row.ErrorMsg, row.Extra), current),
			SensitiveMasked: current.ScopeType != pkgscope.SubjectTypePlatform,
			HappenedAt:      row.OperateTime,
		}
		if allowAuditItem(item, current, in) {
			items = append(items, item)
		}
	}
	return items, nil
}

func querySecurityItems(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, window auditQueryWindow, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	q := svcCtx.DB.WithContext(ctx).Table("sys_security_event").Where("created_at >= ? and created_at <= ?", window.start, window.end)
	q = applyScopeWhere(q, current, "sys_security_event")
	if in.TraceId != "" {
		q = q.Where("trace_id = ?", in.TraceId)
	}
	if in.OperatorId > 0 {
		q = q.Where("operator_id = ?", in.OperatorId)
	}
	if in.OperatorName != "" {
		q = q.Where("operator_name like ?", likeValue(in.OperatorName))
	}
	if in.ResourceType != "" {
		q = q.Where("resource_type = ?", in.ResourceType)
	}
	if in.ResourceId > 0 {
		q = q.Where("resource_id = ?", in.ResourceId)
	}
	if in.EventType != "" {
		q = q.Where("event_type like ? or action like ?", likeValue(in.EventType), likeValue(in.EventType))
	}
	if in.Result != "" {
		q = q.Where("result like ?", likeValue(in.Result))
	}
	if in.Keyword != "" {
		q = q.Where("request_summary like ? or payload like ?", likeValue(in.Keyword), likeValue(in.Keyword))
	}
	var rows []securityEventRow
	if err := q.Order("created_at desc").Limit(300).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, len(rows))
	for _, row := range rows {
		item := auditCenterItem{SourceType: sourceSecurity, SourceID: row.ID, TraceID: row.TraceID, EventType: row.EventType, Action: row.Action, Result: row.Result, ScopeType: row.ScopeType, ScopeLabel: logiccommon.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID).Label(), PlatformID: row.PlatformID, TenantID: row.TenantID, MerchantID: row.MerchantID, OperatorID: row.OperatorID, OperatorName: row.OperatorName, ResourceType: row.ResourceType, ResourceID: row.ResourceID, SubjectInfo: logiccommon.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID).Label(), RequestSummary: maskText(row.RequestSummary, current), DetailPayload: maskText(row.Payload, current), SensitiveMasked: current.ScopeType != pkgscope.SubjectTypePlatform, HappenedAt: row.CreatedAt}
		if allowAuditItem(item, current, in) {
			items = append(items, item)
		}
	}
	return items, nil
}

func queryTenantItems(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, window auditQueryWindow, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	q := svcCtx.DB.WithContext(ctx).Table("sys_tenant_audit").Where("created_at >= ? and created_at <= ?", window.start, window.end)
	q = applyScopeWhere(q, current, "sys_tenant_audit")
	if in.OperatorId > 0 {
		q = q.Where("operator_id = ?", in.OperatorId)
	}
	if in.OperatorName != "" {
		q = q.Where("operator_name like ?", likeValue(in.OperatorName))
	}
	if in.Keyword != "" {
		q = q.Where("tenant_code like ? or action like ? or request_payload like ?", likeValue(in.Keyword), likeValue(in.Keyword), likeValue(in.Keyword))
	}
	var rows []tenantmodel.SysTenantAudit
	if err := q.Order("created_at desc").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, len(rows))
	for _, row := range rows {
		item := auditCenterItem{SourceType: sourceTenant, SourceID: row.ID, EventType: row.Action, Action: row.Action, Result: row.Result, ScopeType: pkgscope.SubjectTypeTenant, ScopeLabel: logiccommon.DefaultScope(row.PlatformID, row.TenantID, 0).Label(), PlatformID: row.PlatformID, TenantID: row.TenantID, OperatorID: row.OperatorID, OperatorName: row.OperatorName, ResourceType: "tenant", ResourceID: row.TenantID, ResourceName: row.TenantCode, SubjectInfo: logiccommon.DefaultScope(row.PlatformID, row.TenantID, 0).Label(), RequestSummary: maskText(row.RequestPayload, current), DetailPayload: maskText(row.EventPayload, current), SensitiveMasked: current.ScopeType != pkgscope.SubjectTypePlatform, HappenedAt: row.CreatedAt}
		if allowAuditItem(item, current, in) {
			items = append(items, item)
		}
	}
	return items, nil
}

func queryMerchantItems(ctx context.Context, svcCtx *svc.ServiceContext, current pkgscope.GovernanceScope, window auditQueryWindow, in *sysclient.QueryAuditCenterListReq) ([]auditCenterItem, error) {
	q := svcCtx.DB.WithContext(ctx).Table("sys_merchant_audit").Where("created_at >= ? and created_at <= ?", window.start, window.end)
	q = applyScopeWhere(q, current, "sys_merchant_audit")
	if in.TraceId != "" {
		q = q.Where("trace_id = ?", in.TraceId)
	}
	if in.OperatorId > 0 {
		q = q.Where("operator_id = ?", in.OperatorId)
	}
	if in.OperatorName != "" {
		q = q.Where("operator_name like ?", likeValue(in.OperatorName))
	}
	if in.Keyword != "" {
		q = q.Where("merchant_code like ? or action like ? or request_payload like ?", likeValue(in.Keyword), likeValue(in.Keyword), likeValue(in.Keyword))
	}
	var rows []merchantmodel.SysMerchantAudit
	if err := q.Order("created_at desc").Limit(200).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]auditCenterItem, 0, len(rows))
	for _, row := range rows {
		item := auditCenterItem{SourceType: sourceMerchant, SourceID: row.ID, TraceID: row.TraceID, EventType: row.Action, Action: row.Action, Result: row.Result, ScopeType: pkgscope.SubjectTypeMerchant, ScopeLabel: logiccommon.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID).Label(), PlatformID: row.PlatformID, TenantID: row.TenantID, MerchantID: row.MerchantID, OperatorID: row.OperatorID, OperatorName: row.OperatorName, ResourceType: "merchant", ResourceID: row.MerchantID, ResourceName: row.MerchantCode, SubjectInfo: logiccommon.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID).Label(), RequestSummary: maskText(row.RequestPayload, current), DetailPayload: maskText(row.EventPayload, current), SensitiveMasked: current.ScopeType != pkgscope.SubjectTypePlatform, HappenedAt: row.CreatedAt}
		if allowAuditItem(item, current, in) {
			items = append(items, item)
		}
	}
	return items, nil
}

func applyScopeWhere(db *gorm.DB, current pkgscope.GovernanceScope, alias string) *gorm.DB {
	if current.ScopeType == pkgscope.SubjectTypePlatform {
		return db
	}
	where, args := logiccommon.ScopeFilterSQL(alias, current)
	return db.Where(where, args...)
}

func allowAuditItem(item auditCenterItem, current pkgscope.GovernanceScope, in *sysclient.QueryAuditCenterListReq) bool {
	if !logiccommon.DefaultScope(item.PlatformID, item.TenantID, item.MerchantID).SameScope(current) && current.ScopeType != pkgscope.SubjectTypePlatform {
		return false
	}
	if in.TraceId != "" && item.TraceID != in.TraceId {
		return false
	}
	if in.OperatorId > 0 && item.OperatorID != in.OperatorId {
		return false
	}
	if in.ResourceType != "" && !strings.EqualFold(item.ResourceType, in.ResourceType) {
		return false
	}
	if in.ResourceId > 0 && item.ResourceID != in.ResourceId {
		return false
	}
	if in.EventType != "" && !containsAny(item.EventType+" "+item.Action, in.EventType) {
		return false
	}
	if in.Result != "" && !containsAny(item.Result, in.Result) {
		return false
	}
	if in.Keyword != "" && !containsAny(strings.Join([]string{item.OperatorName, item.ResourceName, item.SubjectInfo, item.RequestSummary, item.DetailPayload}, " "), in.Keyword) {
		return false
	}
	return true
}

func findAuditItem(items []auditCenterItem, sourceType string, sourceID int64) (auditCenterItem, error) {
	for _, item := range items {
		if item.SourceType == sourceType && item.SourceID == sourceID {
			return item, nil
		}
	}
	return auditCenterItem{}, errors.New("审计记录不存在")
}

func buildTimeline(items []auditCenterItem, current auditCenterItem) []*sysclient.AuditTimelineItem {
	related := make([]auditCenterItem, 0, 8)
	for _, item := range items {
		if current.TraceID != "" && item.TraceID == current.TraceID {
			related = append(related, item)
			continue
		}
		if current.ResourceType != "" && item.ResourceType == current.ResourceType && item.ResourceID > 0 && item.ResourceID == current.ResourceID {
			related = append(related, item)
		}
	}
	sort.Slice(related, func(i, j int) bool { return related[i].HappenedAt.Before(related[j].HappenedAt) })
	out := make([]*sysclient.AuditTimelineItem, 0, len(related))
	for _, item := range related {
		out = append(out, &sysclient.AuditTimelineItem{SourceType: item.SourceType, SourceId: item.SourceID, TraceId: item.TraceID, EventType: item.EventType, Action: item.Action, Result: item.Result, OperatorName: item.OperatorName, ResourceType: item.ResourceType, ResourceId: item.ResourceID, ResourceName: item.ResourceName, RequestSummary: item.RequestSummary, HappenedAt: item.HappenedAt.Format(time.DateTime)})
	}
	return out
}

func parseGovernanceExtra(raw string) governanceExtra {
	var extra governanceExtra
	_ = json.Unmarshal([]byte(raw), &extra)
	return extra
}

func likeValue(value string) string { return "%" + strings.TrimSpace(value) + "%" }
func chooseFirst(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func containsAny(text, sub string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(strings.TrimSpace(sub)))
}
func mapResultFromStatus(status int32, login bool) string {
	if status == 1 {
		return "success"
	}
	if login {
		return "failed"
	}
	return "error"
}
func maskText(value string, current pkgscope.GovernanceScope) string {
	if current.ScopeType == pkgscope.SubjectTypePlatform {
		return value
	}
	value = strings.ReplaceAll(value, "\n", " ")
	if len(value) > 200 {
		value = value[:200] + "..."
	}
	return value
}
