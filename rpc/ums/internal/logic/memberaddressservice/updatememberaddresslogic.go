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
	"gorm.io/gorm"
	"time"
)

// UpdateMemberAddressLogic 更新会员收货地址
/*
Author: LiuFeiHua
Date: 2025/05/21 10:37:06
*/
type UpdateMemberAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMemberAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMemberAddressLogic {
	return &UpdateMemberAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateMemberAddress 更新会员收货地址
func (l *UpdateMemberAddressLogic) UpdateMemberAddress(in *umsclient.UpdateMemberAddressReq) (*umsclient.UpdateMemberAddressResp, error) {
	err := query.Q.Transaction(func(tx *query.Query) error {

		q := tx.UmsMemberAddress

		item, err := q.WithContext(l.ctx).Where(q.ID.Eq(in.Id), q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0)).First()

		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			logc.Errorf(l.ctx, "会员收货地址不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
			return err
		case err != nil:
			logc.Errorf(l.ctx, "查询会员收货地址异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
			return err
		}

		isDefault := int32(0)
		if in.IsDefault == 1 {
			isDefault = 1
			if err = clearMemberDefaultAddresses(l.ctx, tx, in.MemberId); err != nil {
				return err
			}
		} else if item.IsDefault == 1 {
			otherDefaultCount, countErr := q.WithContext(l.ctx).
				Where(q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0), q.IsDefault.Eq(1), q.ID.Neq(in.Id)).
				Count()
			if countErr != nil {
				return countErr
			}
			if otherDefaultCount == 0 {
				isDefault = 1
			}
		}

		now := time.Now()
		address := &model.UmsMemberAddress{
			ID:            in.Id,            // 主键ID
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
			CreateTime:    item.CreateTime,  // 创建时间
			UpdateTime:    &now,             // 更新时间
			IsDeleted:     item.IsDeleted,   // 是否删除
		}

		if _, err = q.WithContext(l.ctx).Where(q.ID.Eq(in.Id), q.MemberID.Eq(in.MemberId), q.IsDeleted.Eq(0)).Updates(address); err != nil {
			return err
		}
		return ensureMemberHasDefaultAddress(l.ctx, tx, in.MemberId)

	})

	if err != nil {
		logc.Errorf(l.ctx, "更新会员收货地址失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新会员收货地址失败")
	}

	return &umsclient.UpdateMemberAddressResp{}, nil
}
