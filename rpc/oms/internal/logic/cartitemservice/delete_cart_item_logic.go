package cartitemservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteCartItemLogic 删除购物车
/*
Author: LiuFeiHua
Date: 2024/6/12 9:53
*/
type DeleteCartItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCartItemLogic {
	return &DeleteCartItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteCartItem 删除/清空购物车
// MEDIUM-8: Where 条件强制包含 MemberID，防止跨会员越权删除
// 采用物理删除：避免与加购 upsert 的唯一键 uk_member_sku_status(member_id, product_sku_id, delete_status)
// 冲突——软删除会留下墓碑记录(delete_status=1)，同一 SKU 第二次删除时触发 Duplicate entry。
// 业务层所有查询均只读 delete_status=0 的记录，墓碑数据无人使用，故直接物理删除。
func (l *DeleteCartItemLogic) DeleteCartItem(in *omsclient.DeleteCartItemReq) (*omsclient.CartItemResp, error) {
	item := query.OmsCartItem
	q := item.WithContext(l.ctx).Where(item.MemberID.Eq(in.MemberId))
	if len(in.Ids) > 0 {
		q = q.Where(item.ID.In(in.Ids...))
	}
	_, err := q.Delete()

	if err != nil {
		logc.Errorf(l.ctx, "删除购物车失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除购物车失败")
	}

	return &omsclient.CartItemResp{}, nil
}
