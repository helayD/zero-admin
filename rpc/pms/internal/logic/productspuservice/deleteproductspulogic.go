package productspuservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/pms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteProductSpuLogic 删除商品SPU
/*
Author: LiuFeiHua
Date: 2025/06/16 14:37:38
*/
type DeleteProductSpuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteProductSpuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteProductSpuLogic {
	return &DeleteProductSpuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteProductSpu 删除商品SPU
func (l *DeleteProductSpuLogic) DeleteProductSpu(in *pmsclient.DeleteProductSpuReq) (*pmsclient.DeleteProductSpuResp, error) {
	q := query.PmsProductSpu
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, 0)
	if err != nil {
		return nil, err
	}
	if _, err := logiccommon.EnsureProductScope(l.ctx, l.svcCtx.DB, currentScope, in.Ids, "pms.product_spu.delete", 0, "", "delete product spu"); err != nil {
		return nil, err
	}

	_, err = q.WithContext(l.ctx).Where(q.ID.In(in.Ids...)).Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除商品SPU失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除商品SPU失败")
	}

	sendProductESDelete(l.ctx, l.svcCtx, in.Ids, currentScope)

	return &pmsclient.DeleteProductSpuResp{}, nil
}
