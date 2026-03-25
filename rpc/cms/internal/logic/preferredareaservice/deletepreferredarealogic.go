package preferredareaservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/cms/cmsclient"
	"github.com/feihua/zero-admin/rpc/cms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/cms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/cms/internal/svc"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeletePreferredAreaLogic 删除优选专区
/*
Author: LiuFeiHua
Date: 2025/01/23 15:24:00
*/
type DeletePreferredAreaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePreferredAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePreferredAreaLogic {
	return &DeletePreferredAreaLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeletePreferredArea 删除优选专区
func (l *DeletePreferredAreaLogic) DeletePreferredArea(in *cmsclient.DeletePreferredAreaReq) (*cmsclient.DeletePreferredAreaResp, error) {
	q := query.CmsPreferredArea
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, "")
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsurePreferredAreaScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "cms.preferred_area.delete", "", "delete preferred_area"); err != nil {
		return nil, err
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除优选专区失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除优选专区失败")
	}

	return &cmsclient.DeletePreferredAreaResp{}, nil
}
