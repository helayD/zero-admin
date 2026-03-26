package cart

import (
	"context"
	"fmt"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCartItemQuantityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCartItemQuantityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCartItemQuantityLogic {
	return &UpdateCartItemQuantityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateCartItemQuantity 修改购物车中某个商品的数量（带商品可售状态校验）
// 编排顺序：①查询购物车项 → ②查询SKU库存 → ③校验 → ④写入购物车
// AC#1: 消费者修改数量 → 即时更新购物车结果并显示促销信息
func (l *UpdateCartItemQuantityLogic) UpdateCartItemQuantity(req *types.UpdateCartItemQuantityReq) (resp *types.CartItemResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 参数合法性校验
	if req.Id <= 0 {
		logc.Errorf(l.ctx, "修改购物车数量参数非法, id=%d", req.Id)
		return nil, errorx.NewDefaultError("参数错误")
	}
	if req.Quantity <= 0 {
		logc.Errorf(l.ctx, "修改购物车数量非法, quantity=%d", req.Quantity)
		return nil, errorx.NewDefaultError("数量必须大于0")
	}

	// scope: front-api JWT 不携带 platform_id/tenant_id/merchant_id，
	// ResolveEffectiveGovernanceScope 永远返回 platform:1。
	// 本 story 阶段不做 scope 过滤，商户归属在 Epic 5 统一处理（已知架构债务）。
	scope := common.ResolveEffectiveGovernanceScope(l.ctx)

	// ① 先查询当前购物车项，获取 productId 和 productSkuId
	listResp, err := l.svcCtx.CartItemService.QueryCartItemList(l.ctx, &omsclient.QueryCartItemListReq{
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询购物车列表失败, memberId=%d, err=%s", memberId, err.Error())
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	var cartItem *omsclient.CartItemData
	for _, item := range listResp.List {
		if item.Id == req.Id {
			cartItem = item
			break
		}
	}
	if cartItem == nil {
		logc.Errorf(l.ctx, "购物车项不存在, id=%d, memberId=%d", req.Id, memberId)
		return nil, errorx.NewDefaultError("购物车项不存在")
	}

	// ② 调用 PMS ProductSkuService.QueryProductSkuDetail 查询 SKU 库存
	skuResp, err := l.svcCtx.ProductSkuService.QueryProductSkuDetail(l.ctx, &pmsclient.QueryProductSkuDetailReq{
		Id:    cartItem.ProductSkuId,
		Scope: common.PMSGovernanceScope(scope),
	})
	if err != nil {
		s, _ := status.FromError(err)
		// SKU 不存在
		if strings.Contains(s.Message(), "商品SKU不存在") || strings.Contains(s.Message(), "不存在") {
			return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
		}
		// PMS 服务异常
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询商品SKU失败, productSkuId=%d, err=%s", cartItem.ProductSkuId, err.Error())
			return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
		}
		logc.Errorf(l.ctx, "查询商品SKU详情失败, productSkuId=%d, err=%s", cartItem.ProductSkuId, err.Error())
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	if skuResp == nil {
		return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
	}

	// ③ 调用 PMS ProductSpuService.QueryProductSpuDetail 查询 SPU 状态
	spuResp, err := l.svcCtx.ProductSpuService.QueryProductSpuDetail(l.ctx, &pmsclient.QueryProductSpuDetailReq{
		Id:    cartItem.ProductId,
		Scope: common.PMSGovernanceScope(scope),
	})
	if err != nil {
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询商品SPU失败, productId=%d, err=%s", cartItem.ProductId, err.Error())
			return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
		}
		s, _ := status.FromError(err)
		if strings.Contains(s.Message(), "商品SPU不存在") || strings.Contains(s.Message(), "商品不存在") {
			return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
		}
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	if spuResp.Data != nil {
		// 下架校验
		if spuResp.Data.PublishStatus != 1 {
			logc.Infof(l.ctx, "商品已下架, productId=%d, publishStatus=%d", cartItem.ProductId, spuResp.Data.PublishStatus)
			return nil, errorx.NewDefaultError(ErrCodeCartProductOffline)
		}
		// 审核状态校验
		if spuResp.Data.VerifyStatus != 1 {
			logc.Infof(l.ctx, "商品审核中或未通过, productId=%d, verifyStatus=%d", cartItem.ProductId, spuResp.Data.VerifyStatus)
			return nil, errorx.NewDefaultError(ErrCodeCartProductUnverified)
		}
	}

	// 库存校验：修改后数量 > 库存 → 拒绝
	if int32(skuResp.Stock) < req.Quantity {
		logc.Infof(l.ctx, "库存不足, productSkuId=%d, stock=%d, requested=%d",
			cartItem.ProductSkuId, skuResp.Stock, req.Quantity)
		return nil, errorx.NewDefaultError(fmt.Sprintf(ErrMsgStockInsufficientTpl, skuResp.Stock))
	}

	// ④ 透传给 oms-rpc UpdateCartItemQuantity
	_, err = l.svcCtx.CartItemService.UpdateCartItemQuantity(l.ctx, &omsclient.UpdateCartItemQuantityReq{
		Quantity: req.Quantity,
		Id:       req.Id,
		MemberId: memberId,
	})

	if err != nil {
		logc.Errorf(l.ctx, "修改购物车中某个商品的数量失败,参数memberId: %+v,异常：%s", memberId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.CartItemResp{
		Code:    0,
		Message: "修改购物车中某个商品的数量成功",
	}, nil
}
