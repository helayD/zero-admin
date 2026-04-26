package memberaddressservicelogic

import (
	"context"
	"errors"
	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/gen/query"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// AddMemberAddressLogic 添加会员收货地址
/*
Author: LiuFeiHua
Date: 2025/05/21 10:37:06
*/
type AddMemberAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMemberAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMemberAddressLogic {
	return &AddMemberAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddMemberAddress 添加会员收货地址
func (l *AddMemberAddressLogic) AddMemberAddress(in *umsclient.AddMemberAddressReq) (*umsclient.AddMemberAddressResp, error) {
	err := query.Q.Transaction(func(tx *query.Query) error {
		q := tx.UmsMemberAddress
		activeCount, err := q.WithContext(l.ctx).
			Where(q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0)).
			Count()
		if err != nil {
			return err
		}

		isDefault := int32(0)
		if activeCount == 0 || in.IsDefault == 1 {
			isDefault = 1
			if err = clearMemberDefaultAddresses(l.ctx, tx, in.MemberId); err != nil {
				return err
			}
		}

		if err = q.WithContext(l.ctx).Create(&model.UmsMemberAddress{
			MemberID:      in.MemberId,      // 会员ID
			ReceiverName:  in.ReceiverName,  // 收货人姓名
			ReceiverPhone: in.ReceiverPhone, // 收货人电话
			Province:      in.Province,      // 省份
			City:          in.City,          // 城市
			District:      in.District,      // 区县
			DetailAddress: in.DetailAddress, // 详细地址
			PostalCode:    in.PostalCode,    // 邮政编码
			Tag:           in.Tag,           // 地址标签：家、公司等
			IsDefault:     isDefault,        // 是否默认地址
			IsDeleted:     0,                // 是否删除
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加会员收货地址失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加会员收货地址失败")
	}

	return &umsclient.AddMemberAddressResp{}, nil
}
