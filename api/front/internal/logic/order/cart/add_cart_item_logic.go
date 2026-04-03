package cart

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeCartProductOffline    = "OMS_CART_PRODUCT_OFFLINE"
	ErrCodeCartProductUnverified = "OMS_CART_PRODUCT_UNVERIFIED"
	ErrCodeCartStockInsufficient = "OMS_CART_STOCK_INSUFFICIENT"
	ErrCodeCartProductNotFound   = "OMS_CART_PRODUCT_NOT_FOUND"
	ErrCodeCartSystemError       = "OMS_CART_SYSTEM_ERROR"
	ErrMsgStockInsufficientTpl   = "库存不足，当前仅剩 %d 件"
)

type AddCartItemLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCartItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCartItemLogic {
	return &AddCartItemLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddCartItem 添加商品进购物车（带商品可售状态校验）
// 编排顺序：①查询SPU状态 → ②查询SKU库存 → ③写入购物车
// AC#1: 商品详情页展示有效规格和库存，消费者选择规格与数量并点击加入购物车 → 成功
// AC#2: 商品库存不足、规格失效或商品已下架 → 拒绝加购并说明具体原因
func (l *AddCartItemLogic) AddCartItem(req *types.CartItemReq) (resp *types.CartItemResp, err error) {
	// ① 获取会员ID
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 参数合法性校验：防止无效脏数据写入购物车
	if req.ProductId <= 0 || req.ProductSkuId <= 0 || req.Quantity <= 0 {
		logc.Errorf(l.ctx, "加购参数非法, productId=%d, skuId=%d, quantity=%d", req.ProductId, req.ProductSkuId, req.Quantity)
		return nil, errorx.NewDefaultError("参数错误")
	}

	// ② 调用 PMS ProductSpuService.QueryProductSpuDetail 查询商品当前状态
	// scope 处理：front-api JWT 不携带 platform_id/tenant_id/merchant_id，
	// ResolveEffectiveGovernanceScope 永远返回 platform:1。
	// 本 story 阶段不做 scope 校验，商户归属在商品快照中体现（MVP 范围）。
	scope := common.ResolveEffectiveGovernanceScope(l.ctx)
	spuResp, err := l.svcCtx.ProductSpuService.QueryProductSpuDetail(l.ctx, &pmsclient.QueryProductSpuDetailReq{
		Id:    req.ProductId,
		Scope: common.PMSGovernanceScope(scope),
	})
	if err != nil {
		s, _ := status.FromError(err)
		// PMS 服务不可用时，降级为系统异常
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询商品SPU失败, productId=%d, err=%s", req.ProductId, err.Error())
			return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
		}
		// 商品不存在（gRPC NOT_FOUND）
		if strings.Contains(s.Message(), "商品SPU不存在") || strings.Contains(s.Message(), "商品不存在") {
			return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
		}
		logc.Errorf(l.ctx, "查询商品SPU详情失败, productId=%d, err=%s", req.ProductId, err.Error())
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	// 校验 SPU 可售状态
	if spuResp.Data == nil {
		return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
	}

	productData := spuResp.Data
	// publish_status ≠ 1 → 商品已下架
	if productData.PublishStatus != 1 {
		logc.Infof(l.ctx, "商品已下架, productId=%d, publishStatus=%d", req.ProductId, productData.PublishStatus)
		return nil, errorx.NewDefaultError(ErrCodeCartProductOffline)
	}
	// verify_status ≠ 1 → 商品审核中或审核未通过
	if productData.VerifyStatus != 1 {
		logc.Infof(l.ctx, "商品审核中或未通过, productId=%d, verifyStatus=%d", req.ProductId, productData.VerifyStatus)
		return nil, errorx.NewDefaultError(ErrCodeCartProductUnverified)
	}

	// ③ 调用 PMS ProductSkuService.QueryProductSkuDetail 查询 SKU 库存
	skuResp, err := l.svcCtx.ProductSkuService.QueryProductSkuDetail(l.ctx, &pmsclient.QueryProductSkuDetailReq{
		Id:    req.ProductSkuId,
		Scope: common.PMSGovernanceScope(scope),
	})
	if err != nil {
		s, _ := status.FromError(err)
		// SKU 不存在时同样返回商品不存在
		if strings.Contains(s.Message(), "商品SKU不存在") || strings.Contains(s.Message(), "不存在") {
			return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
		}
		// PMS 服务异常
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询商品SKU失败, productSkuId=%d, err=%s", req.ProductSkuId, err.Error())
			return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
		}
		logc.Errorf(l.ctx, "查询商品SKU详情失败, productSkuId=%d, err=%s", req.ProductSkuId, err.Error())
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	if skuResp == nil {
		return nil, errorx.NewDefaultError(ErrCodeCartProductNotFound)
	}

	// 校验 SKU 库存是否充足（MVP 阶段使用 stock >= requestedQuantity，不做 real_stock 校验）
	// real_stock（stock - lock_stock）需要 PMS proto 补 lock_stock 字段，Task 5 中评估
	if int32(skuResp.Stock) < req.Quantity {
		logc.Infof(l.ctx, "库存不足, productSkuId=%d, stock=%d, requested=%d",
			req.ProductSkuId, skuResp.Stock, req.Quantity)
		return nil, errorx.NewDefaultError(fmt.Sprintf(ErrMsgStockInsufficientTpl, skuResp.Stock))
	}

	// ④ 调用 oms-rpc AddCartItem 写入购物车
	_, err = l.svcCtx.CartItemService.AddCartItem(l.ctx, &omsclient.AddCartItemReq{
		MemberId:          memberId,
		ProductId:         req.ProductId,
		ProductSkuId:      req.ProductSkuId,
		Quantity:          req.Quantity,
		Price:             req.Price,
		Selected:          req.Selected,
		ProductName:       req.ProductName,
		ProductSubTitle:   req.ProductSubTitle,
		ProductPic:        req.ProductPic,
		ProductSkuCode:    req.ProductSkuCode,
		ProductSn:         req.ProductSn,
		ProductBrand:      req.ProductBrand,
		ProductCategoryId: req.ProductCategoryId,
		ProductAttr:       req.ProductAttr,
		MemberNickname:    req.MemberNickname,
		Source:            req.Source,
		ActivityType:      req.ActivityType,
		ActivityId:        req.ActivityId,
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加购物车失败, productId=%d, skuId=%d, err=%s", req.ProductId, req.ProductSkuId, err.Error())
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	productScope, scopeErr := pkgscope.NormalizeGovernanceScope(
		productData.ScopeType,
		productData.PlatformId,
		productData.TenantId,
		productData.MerchantId,
	)
	if scopeErr != nil {
		logc.Errorf(l.ctx, "解析加购埋点作用域失败, productId=%d, scope=%s/%d/%d/%d, err=%s",
			req.ProductId, productData.ScopeType, productData.PlatformId, productData.TenantId, productData.MerchantId, scopeErr.Error())
	} else {
		if _, recordErr := l.svcCtx.OperateDashboardService.RecordOperateFunnelEvent(l.ctx, &smsclient.RecordOperateFunnelEventReq{
			Scope:        common.SMSGovernanceScope(productScope),
			EventType:    operatefunnel.EventAddCart,
			Channel:      operatefunnel.CartChannelFromSource(req.Source),
			ActivityType: req.ActivityType,
			ActivityId:   req.ActivityId,
			MemberId:     memberId,
			ProductId:    req.ProductId,
			TraceId:      fmt.Sprintf("cart:add:%d:%d:%d:%d", memberId, req.ProductId, req.ProductSkuId, time.Now().UnixNano()),
			ExtraJson:    fmt.Sprintf(`{"skuId":%d,"quantity":%d}`, req.ProductSkuId, req.Quantity),
		}); recordErr != nil {
			logc.Errorf(l.ctx, "记录加购漏斗事件失败, productId=%d, skuId=%d, err=%s", req.ProductId, req.ProductSkuId, recordErr.Error())
		}
	}

	return &types.CartItemResp{
		Code:    0,
		Message: "添加购物车成功",
	}, nil
}

// isPMSError 判断是否为 PMS 服务不可用错误
// gRPC 状态码映射：Unavailable/Internal/DeadlineExceeded 通常表示服务或网络问题
func isPMSError(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	if !ok {
		return true
	}
	code := s.Code()
	return code == 14 || // Unavailable
		code == 13 || // Internal
		code == 4 // DeadlineExceeded
}
