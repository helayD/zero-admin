package preferredareaservicelogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/model"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdatePreferredAreaLogic 更新优选专区
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type UpdatePreferredAreaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePreferredAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePreferredAreaLogic {
	return &UpdatePreferredAreaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdatePreferredArea 更新优选专区
func (l *UpdatePreferredAreaLogic) UpdatePreferredArea(in *cmsclient.UpdatePreferredAreaReq) (*cmsclient.UpdatePreferredAreaResp, error) {
	q := query.CmsPreferredArea.WithContext(l.ctx)
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsurePreferredAreaScope(l.ctx, l.svcCtx.DB, currentScope, []int64{in.Id}, "cms.preferred_area.update", in.UpdateBy, fmt.Sprintf("preferredAreaId=%d", in.Id)); err != nil {
		return nil, err
	}

	area, err := q.Where(query.CmsPreferredArea.ID.Eq(in.Id)).First()
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, errors.New("优选专区不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询优选专区异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询优选专区异常")
	}

	var dupCount int64
	err = pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.CmsPreferredArea{}).Where("id != ? AND name = ?", in.Id, in.Name),
		currentScope,
		"",
	).Count(&dupCount).Error
	if err != nil {
		return nil, errors.New("更新优选专区失败")
	}
	if dupCount > 0 {
		return nil, errors.New(fmt.Sprintf("优选专区名称：%s,已存在", in.Name))
	}

	now := time.Now()
	item := &model.CmsPreferredArea{
		ID:         in.Id,           // 主键ID
		Name:       in.Name,         // 专区名称
		SubTitle:   in.SubTitle,     // 子标题
		Pic:        in.Pic,          // 展示图片
		Sort:       in.Sort,         // 排序
		ShowStatus: in.ShowStatus,   // 显示状态：0->不显示；1->显示
		CreateBy:   area.CreateBy,   // 创建者
		CreateTime: area.CreateTime, // 创建时间
		UpdateBy:   in.UpdateBy,     // 更新者
		UpdateTime: &now,            // 更新时间
	}

	// 2.优选专区存在时,则直接更新优选专区
	err = l.svcCtx.DB.Model(&model.CmsPreferredArea{}).WithContext(l.ctx).Where(query.CmsPreferredArea.ID.Eq(in.Id)).Save(item).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新优选专区失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("更新优选专区失败")
	}

	return &cmsclient.UpdatePreferredAreaResp{}, nil
}
