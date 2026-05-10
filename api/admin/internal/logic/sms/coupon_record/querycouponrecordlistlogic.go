package coupon_record

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// QueryCouponRecordListLogic 查询优惠券领取记录列表
/*
Author: LiuFeiHua
Date: 2025/06/12 10:02:03
*/
type QueryCouponRecordListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCouponRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCouponRecordListLogic {
	return &QueryCouponRecordListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryCouponRecordList 查询优惠券领取记录列表
func (l *QueryCouponRecordListLogic) QueryCouponRecordList(req *types.QueryCouponRecordListReq) (resp *types.QueryCouponRecordListResp, err error) {
	result, err := l.svcCtx.CouponRecordService.QueryCouponRecordList(l.ctx, &smsclient.QueryCouponRecordListReq{
		PageNum:  req.Current,
		PageSize: req.PageSize,
		CouponId: req.CouponId, // 优惠券ID
		MemberId: req.MemberId, // 用户ID
		GetType:  req.GetType,  // 获取类型：0->后台赠送；1->主动获取
		OrderId:  req.OrderId,  // 使用订单ID
		Status:   req.Status,   // 状态：0-未使用，1-已使用，2-已过期，3-已失效
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询字优惠券领取记录列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// 当前作用域用于跨服务 RPC 的治理过滤；失败时降级为空 scope，保证主链路不中断
	queryScope, scopeErr := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{})
	if scopeErr != nil {
		logc.Errorf(l.ctx, "解析治理作用域失败,降级为空scope继续拼装,参数：%+v,异常：%s", req, scopeErr.Error())
	}

	// 1) 按 CouponId 拉一次 sms_coupon 详情，拿优惠券码（前端列表每次都按 couponId 过滤，整页共享同一 code）
	couponCode := ""
	if req.CouponId > 0 {
		couponDetail, detailErr := l.svcCtx.CouponService.QueryCouponDetail(l.ctx, &smsclient.QueryCouponDetailReq{
			Id:    req.CouponId,
			Scope: admincommon.SMSGovernanceScope(queryScope),
		})
		if detailErr != nil {
			logc.Errorf(l.ctx, "查询优惠券码失败,couponId=%d,异常：%s", req.CouponId, detailErr.Error())
		} else if couponDetail != nil {
			couponCode = couponDetail.Code
		}
	}

	// 2) 收集 memberId，一次性调 ums 批量查询脱敏昵称（C 端监管口径，此处走 ums 的脱敏字段）
	memberNickMap := make(map[int64]string)
	if len(result.List) > 0 {
		seen := make(map[int64]struct{}, len(result.List))
		memberIds := make([]int64, 0, len(result.List))
		for _, detail := range result.List {
			if detail.MemberId <= 0 {
				continue
			}
			if _, ok := seen[detail.MemberId]; ok {
				continue
			}
			seen[detail.MemberId] = struct{}{}
			memberIds = append(memberIds, detail.MemberId)
		}
		if len(memberIds) > 0 {
			briefResp, briefErr := l.svcCtx.MemberInfoService.QueryMemberBriefByIds(l.ctx, &umsclient.QueryMemberBriefByIdsReq{
				MemberIds: memberIds,
			})
			if briefErr != nil {
				logc.Errorf(l.ctx, "批量查询会员简要信息失败,memberIds=%v,异常：%s", memberIds, briefErr.Error())
			} else if briefResp != nil {
				for _, brief := range briefResp.List {
					if brief == nil {
						continue
					}
					memberNickMap[brief.MemberId] = brief.NicknameMasked
				}
			}
		}
	}

	var list []*types.QueryCouponRecordListData

	for _, detail := range result.List {
		list = append(list, &types.QueryCouponRecordListData{
			Id:             detail.Id,                      // 主键ID
			CouponId:       detail.CouponId,                // 优惠券ID
			CouponCode:     couponCode,                     // 优惠券码（按 CouponId 过滤场景下整页共享）
			MemberId:       detail.MemberId,                // 用户ID
			MemberNickName: memberNickMap[detail.MemberId], // 会员脱敏昵称
			GetTime:        detail.GetTime,                 // 领取时间
			GetType:        detail.GetType,                 // 获取类型：0->后台赠送；1->主动获取
			UseTime:        detail.UseTime,                 // 使用时间
			OrderId:        detail.OrderId,                 // 使用订单ID
			OrderAmount:    detail.OrderAmount,             // 订单金额
			DiscountAmount: detail.DiscountAmount,          // 优惠金额
			Status:         detail.Status,                  // 状态：0-未使用，1-已使用，2-已过期，3-已失效
			InvalidTime:    detail.InvalidTime,             // 失效时间
			InvalidReason:  detail.InvalidReason,           // 失效原因
			CreateTime:     detail.CreateTime,              // 创建时间
		})
	}

	return &types.QueryCouponRecordListResp{
		Code:     "000000",
		Message:  "查询优惠券领取记录列表成功",
		Data:     list,
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    result.Total,
		Success:  true,
	}, nil
}
