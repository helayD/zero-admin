package subjectservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateSubjectStatusLogic 更新专题
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type UpdateSubjectStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSubjectStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSubjectStatusLogic {
	return &UpdateSubjectStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateSubjectStatus 更新专题状态
func (l *UpdateSubjectStatusLogic) UpdateSubjectStatus(in *cmsclient.UpdateSubjectStatusReq) (*cmsclient.UpdateSubjectStatusResp, error) {
	q := query.CmsSubject
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.UpdateBy)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureSubjectScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "cms.subject.status", in.UpdateBy, "update subject status"); err != nil {
		return nil, err
	}

	// 发布校验：当目标状态为1（显示/发布）时，对每个专题执行发布前置校验
	if in.ShowStatus == 1 {
		for _, id := range in.Ids {
			detail, qErr := q.WithContext(l.ctx).Where(q.ID.Eq(id)).First()
			if qErr != nil {
				logc.Errorf(l.ctx, "查询专题失败,id:%d,异常:%s", id, qErr.Error())
				return nil, fmt.Errorf("专题(ID:%d)不存在", id)
			}

			var missing []string
			if strings.TrimSpace(detail.Title) == "" {
				missing = append(missing, "标题")
			}
			if detail.CategoryID == 0 {
				missing = append(missing, "专题分类")
			}
			if strings.TrimSpace(detail.Pic) == "" {
				missing = append(missing, "专题主图")
			}
			if len(missing) > 0 {
				return nil, fmt.Errorf("专题「%s」发布校验未通过：缺少%s", detail.Title, strings.Join(missing, "、"))
			}
		}
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Update(q.ShowStatus, in.ShowStatus)

	if err != nil {
		logc.Errorf(l.ctx, "更新专题状态失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新专题状态失败")
	}

	return &cmsclient.UpdateSubjectStatusResp{}, nil
}
