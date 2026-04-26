package memberaddressservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteMemberAddressLogic 删除会员收货地址
/*
Author: LiuFeiHua
Date: 2025/05/21 10:37:06
*/
type DeleteMemberAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMemberAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMemberAddressLogic {
	return &DeleteMemberAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteMemberAddress 删除会员收货地址
func (l *DeleteMemberAddressLogic) DeleteMemberAddress(in *umsclient.DeleteMemberAddressReq) (*umsclient.DeleteMemberAddressResp, error) {
	err := query.Q.Transaction(func(tx *query.Query) error {
		q := tx.UmsMemberAddress
		items, err := q.WithContext(l.ctx).
			Where(q.ID.In(in.Ids...), q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0)).
			Find()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return errors.New("会员收货地址不存在")
		}

		deletedDefault := false
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
			if item.IsDefault == 1 {
				deletedDefault = true
			}
		}

		now := time.Now()
		if _, err = q.WithContext(l.ctx).
			Where(q.ID.In(ids...), q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0)).
			Updates(map[string]interface{}{
				"is_deleted":  1,
				"is_default":  0,
				"update_time": now,
			}); err != nil {
			return err
		}

		if deletedDefault {
			return ensureMemberHasDefaultAddress(l.ctx, tx, in.MemberId)
		}
		return nil
	})
	if err != nil {
		logc.Errorf(l.ctx, "删除会员收货地址失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除会员收货地址失败")
	}

	return &umsclient.DeleteMemberAddressResp{}, nil
}
