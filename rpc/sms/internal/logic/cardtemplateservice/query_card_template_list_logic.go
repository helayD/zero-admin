package cardtemplateservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryCardTemplateListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCardTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCardTemplateListLogic {
	return &QueryCardTemplateListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCardTemplateList 查询卡片模板列表（带筛选 + 分页 + 引用规则数量回填）
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09
*/
func (l *QueryCardTemplateListLogic) QueryCardTemplateList(in *smsclient.QueryCardTemplateListReq) (*smsclient.QueryCardTemplateListResp, error) {
	platformId, tenantId, merchantId := resolveScope(in.Scope)

	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	q := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			platformId, tenantId, merchantId)
	if in.TemplateName != "" {
		q = q.Where("template_name LIKE ?", "%"+in.TemplateName+"%")
	}
	if in.TemplateCode != "" {
		q = q.Where("template_code = ?", in.TemplateCode)
	}
	if in.Rarity != "" {
		q = q.Where("rarity = ?", in.Rarity)
	}
	if in.Status >= 0 {
		q = q.Where("status = ?", in.Status)
	}
	if in.DisplayStatus >= 0 {
		q = q.Where("display_status = ?", in.DisplayStatus)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "查询卡片模板总数失败: %v", err)
		return nil, err
	}

	var rows []cardTemplateRow
	if err := q.
		Select(cardTemplateSelectColumns).
		Order("id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "查询卡片模板列表失败: %v", err)
		return nil, err
	}

	// 批量回填每个模板被多少条发卡规则引用
	templateIds := make([]int64, 0, len(rows))
	for i := range rows {
		templateIds = append(templateIds, rows[i].Id)
	}
	refCounts := make(map[int64]int32, len(templateIds))
	if len(templateIds) > 0 {
		type refRow struct {
			CardTemplateId int64 `gorm:"column:card_template_id"`
			Cnt            int64 `gorm:"column:cnt"`
		}
		var refs []refRow
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_rule").
			Select("card_template_id, COUNT(*) AS cnt").
			Where("card_template_id IN ? AND is_deleted = 0", templateIds).
			Group("card_template_id").
			Find(&refs).Error; err != nil {
			logc.Errorf(l.ctx, "回填发卡规则引用计数失败: %v", err)
			// 失败不阻塞列表，按 0 兜底
		} else {
			for _, r := range refs {
				refCounts[r.CardTemplateId] = int32(r.Cnt)
			}
		}
	}

	list := make([]*smsclient.CardTemplateData, 0, len(rows))
	for i := range rows {
		list = append(list, rows[i].toProto(refCounts[rows[i].Id]))
	}

	return &smsclient.QueryCardTemplateListResp{
		List:  list,
		Total: total,
	}, nil
}
