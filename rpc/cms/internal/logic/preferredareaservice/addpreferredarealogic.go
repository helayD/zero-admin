package preferredareaservicelogic

import (
	"context"
	"errors"
	"fmt"

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

// AddPreferredAreaLogic 添加优选专区
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type AddPreferredAreaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddPreferredAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPreferredAreaLogic {
	return &AddPreferredAreaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddPreferredArea 添加优选专区
func (l *AddPreferredAreaLogic) AddPreferredArea(in *cmsclient.AddPreferredAreaReq) (*cmsclient.AddPreferredAreaResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}

	var dupCount int64
	err = pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.CmsPreferredArea{}).Where("name = ?", in.Name),
		currentScope,
		"",
	).Count(&dupCount).Error
	if err != nil {
		return nil, errors.New("添加优选专区失败")
	}
	if dupCount > 0 {
		return nil, errors.New(fmt.Sprintf("优选专区名称：%s,已存在", in.Name))
	}

	area := &model.CmsPreferredArea{
		Name:       in.Name,       // 专区名称
		SubTitle:   in.SubTitle,   // 子标题
		Pic:        in.Pic,        // 展示图片
		Sort:       in.Sort,       // 排序
		ShowStatus: in.ShowStatus, // 显示状态：0->不显示；1->显示
		CreateBy:   in.CreateBy,   // 创建者
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := query.Use(tx).CmsPreferredArea.WithContext(l.ctx).Create(area); err != nil {
			return err
		}
		return logiccommon.ApplyPreferredAreaScope(l.ctx, tx, area.ID, currentScope)
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加优选专区失败,参数:%+v,异常:%s", area, err.Error())
		return nil, errors.New("添加优选专区失败")
	}

	return &cmsclient.AddPreferredAreaResp{}, nil
}
