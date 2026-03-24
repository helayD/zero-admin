package couponservicelogic

import (
	"context"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/gen/model"
	"github.com/feihua/zero-admin/rpc/sms/gen/query"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddCouponLogic 添加优惠券
/*
Author: LiuFeiHua
Date: 2025/06/11 10:44:50
*/
type AddCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCouponLogic {
	return &AddCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddCoupon 添加优惠券
func (l *AddCouponLogic) AddCoupon(in *smsclient.AddCouponReq) (*smsclient.AddCouponResp, error) {
	currentScope, err := logiccommon.ResolveWriteScope(l.ctx, l.svcCtx.DB, in.Scope, in.CreateBy)
	if err != nil {
		return nil, err
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", in.StartTime)
	if err != nil {
		return nil, fmt.Errorf("生效时间格式无效，请使用 yyyy-MM-dd HH:mm:ss 格式")
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", in.EndTime)
	if err != nil {
		return nil, fmt.Errorf("失效时间格式无效，请使用 yyyy-MM-dd HH:mm:ss 格式")
	}
	if !startTime.Before(endTime) {
		return nil, fmt.Errorf("生效时间必须早于失效时间")
	}
	if in.TotalCount <= 0 {
		return nil, fmt.Errorf("发放总量必须大于0")
	}
	if in.PerLimit < 1 {
		return nil, fmt.Errorf("每人限领数量至少为1")
	}
	if in.Amount <= 0 {
		return nil, fmt.Errorf("优惠金额必须大于0")
	}
	if in.MinAmount < 0 {
		return nil, fmt.Errorf("最低使用金额不能为负数")
	}
	item := &model.SmsCoupon{
		TypeID:      in.TypeId,             // 优惠券类型ID
		Name:        in.Name,               // 优惠券名称
		Code:        in.Code,               // 优惠券码
		Amount:      float64(in.Amount),    // 优惠金额/折扣率
		MinAmount:   float64(in.MinAmount), // 最低使用金额
		StartTime:   startTime,             // 生效时间
		EndTime:     endTime,               // 失效时间
		TotalCount:  in.TotalCount,         // 发放总量
		PerLimit:    in.PerLimit,           // 每人限领数量
		Status:      0,                     // 新建优惠券默认为草稿状态(0)
		IsEnabled:   in.IsEnabled,          // 是否启用
		Description: in.Description,        // 使用说明
		CreateBy:    in.CreateBy,           // 创建人ID

	}

	var data []*model.SmsCouponScope
	if len(in.Scopes) == 0 {
		data = append(data, &model.SmsCouponScope{
			CouponID:  item.ID, // 优惠券ID
			ScopeType: 0,       // 范围类型：0-全场，1-分类，2-商品
			ScopeID:   0,       // 范围ID（分类ID或商品ID）
		})
	} else {
		for _, x := range in.Scopes {
			data = append(data, &model.SmsCouponScope{
				CouponID:  item.ID,     // 优惠券ID
				ScopeType: x.ScopeType, // 范围类型：0-全场，1-分类，2-商品
				ScopeID:   x.ScopeId,   // 范围ID（分类ID或商品ID）
			})
		}
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		count, err := qtx.SmsCoupon.WithContext(l.ctx).Where(qtx.SmsCoupon.Name.Eq(in.Name)).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("优惠券：%s,已存在", in.Name)
		}

		if err := qtx.SmsCoupon.WithContext(l.ctx).Create(item); err != nil {
			return err
		}
		for _, row := range data {
			row.CouponID = item.ID
		}

		if err := logiccommon.ApplyCouponScope(l.ctx, tx, item.ID, currentScope); err != nil {
			return err
		}

		return qtx.SmsCouponScope.WithContext(l.ctx).CreateInBatches(data, len(data))
	})
	if err != nil {
		logc.Errorf(l.ctx, "添加优惠券失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}
	return &smsclient.AddCouponResp{}, nil
}
