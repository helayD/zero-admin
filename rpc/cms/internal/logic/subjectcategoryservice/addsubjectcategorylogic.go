package subjectcategoryservicelogic

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

// AddSubjectCategoryLogic 添加专题分类
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type AddSubjectCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddSubjectCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSubjectCategoryLogic {
	return &AddSubjectCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddSubjectCategory 添加专题分类
func (l *AddSubjectCategoryLogic) AddSubjectCategory(in *cmsclient.AddSubjectCategoryReq) (*cmsclient.AddSubjectCategoryResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}

	var dupCount int64
	err = pkgscope.ApplyGovernanceScope(
		l.svcCtx.DB.WithContext(l.ctx).Model(&model.CmsSubjectCategory{}).Where("name = ?", in.Name),
		currentScope,
		"",
	).Count(&dupCount).Error
	if err != nil {
		return nil, errors.New("添加专题分类失败")
	}
	if dupCount > 0 {
		return nil, errors.New(fmt.Sprintf("专题分类名称：%s,已存在", in.Name))
	}

	item := &model.CmsSubjectCategory{
		Name:         in.Name,         // 专题分类名称
		Icon:         in.Icon,         // 分类图标
		SubjectCount: in.SubjectCount, // 专题数量
		ShowStatus:   in.ShowStatus,   // 显示状态：0->不显示；1->显示
		Sort:         in.Sort,         // 排序
		CreateBy:     in.CreateBy,     // 创建者
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := query.Use(tx).CmsSubjectCategory.WithContext(l.ctx).Create(item); err != nil {
			return err
		}
		return logiccommon.ApplySubjectCategoryScope(l.ctx, tx, item.ID, currentScope)
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加专题分类失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加专题分类失败")
	}

	return &cmsclient.AddSubjectCategoryResp{}, nil
}
