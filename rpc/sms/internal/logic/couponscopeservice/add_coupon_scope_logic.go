package couponscopeservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddCouponScopeLogic 添加优惠券使用范围
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type AddCouponScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCouponScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCouponScopeLogic {
	return &AddCouponScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddCouponScope 添加优惠券使用范围
func (l *AddCouponScopeLogic) AddCouponScope(in *smsclient.AddCouponScopeReq) (*smsclient.AddCouponScopeResp, error) {
	if len(in.Data) == 0 {
		return nil, errors.New("使用范围数据不能为空")
	}

	// 校验 scope_type 合法性
	for _, x := range in.Data {
		if x.ScopeType < 0 || x.ScopeType > 2 {
			return nil, fmt.Errorf("使用范围类型无效：%d，有效值为 0(全场)/1(分类)/2(商品)", x.ScopeType)
		}
		// scope_type=1(分类) 或 scope_type=2(商品) 时，scope_id 必须大于0
		if (x.ScopeType == 1 || x.ScopeType == 2) && x.ScopeId <= 0 {
			return nil, fmt.Errorf("使用范围类型为%d时，范围ID必须大于0", x.ScopeType)
		}
	}

	// 确定 coupon_id（取第一条记录的 couponId 作为批量操作的依据）
	couponID := in.Data[0].CouponId
	if couponID <= 0 {
		return nil, errors.New("优惠券ID无效")
	}

	var data []*model.SmsCouponScope
	for _, x := range in.Data {
		data = append(data, &model.SmsCouponScope{
			CouponID:  x.CouponId,  // 优惠券ID
			ScopeType: x.ScopeType, // 范围类型：0-全场，1-分类，2-商品
			ScopeID:   x.ScopeId,   // 范围ID（分类ID或商品ID）
		})
	}

	// 事务内先删旧 scope 再批量写入新 scope
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if _, err := qtx.SmsCouponScope.WithContext(l.ctx).Where(qtx.SmsCouponScope.CouponID.Eq(couponID)).Delete(); err != nil {
			return err
		}
		return qtx.SmsCouponScope.WithContext(l.ctx).CreateInBatches(data, len(data))
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加优惠券使用范围失败,参数:%+v,异常:%s", data, err.Error())
		return nil, errors.New("添加优惠券使用范围失败")
	}

	return &smsclient.AddCouponScopeResp{}, nil
}
