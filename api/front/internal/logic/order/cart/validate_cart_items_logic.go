package cart

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type ValidateCartItemsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewValidateCartItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateCartItemsLogic {
	return &ValidateCartItemsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ValidateCartItems 批量结算前校验商品有效性（Task 3 & AC#2）
// 编排顺序：对每个已选购物车项并行查询 PMS SPU + SKU 状态
// 返回每项的校验结果，前端据此高亮无效商品
func (l *ValidateCartItemsLogic) ValidateCartItems(req *types.CartValidateReq) (resp *types.CartValidateResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// ids 为空时校验所有已选商品；非空时只校验指定 ids
	var ids []int64
	if len(req.Ids) > 0 {
		ids = req.Ids
	} else {
		// 拉取当前会员已选中的购物车项
		listResp, err := l.svcCtx.CartItemService.QueryCartItemList(l.ctx, &omsclient.QueryCartItemListReq{
			MemberId: memberId,
		})
		if err != nil {
			logc.Errorf(l.ctx, "查询购物车列表失败, memberId=%d, err=%s", memberId, err.Error())
			return nil, errorx.NewDefaultError("查询购物车列表失败")
		}
		for _, item := range listResp.List {
			if item.Selected == 1 {
				ids = append(ids, item.Id)
			}
		}
	}

	if len(ids) == 0 {
		return &types.CartValidateResp{
			Code:    0,
			Message: "操作成功",
			Data:    []types.CartValidateResult{},
		}, nil
	}

	// 查询购物车项详情（用于获取 productId/skuId）
	cartResp, err := l.svcCtx.CartItemService.QueryCartItemList(l.ctx, &omsclient.QueryCartItemListReq{
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询购物车详情失败, memberId=%d, err=%s", memberId, err.Error())
		return nil, errorx.NewDefaultError("查询购物车详情失败")
	}

	// 构建 cartItems map（id -> cartItem）
	cartItemMap := make(map[int64]*omsclient.CartItemData)
	for _, item := range cartResp.List {
		cartItemMap[item.Id] = item
	}

	// 对每个需校验的购物车项做并行校验
	scope := common.ResolveEffectiveGovernanceScope(l.ctx)
	pmsScope := common.PMSGovernanceScope(scope)

	// 按 id 顺序存储结果
	results := make([]types.CartValidateResult, len(ids))
	resultMu := sync.Mutex{}

		eg, ctx := errgroup.WithContext(l.ctx)
	for i, id := range ids {
		idx, cartItemId := i, id
		eg.Go(func() error {
			// MEDIUM-7: 每个 goroutine 携带子 traceId
			traceId := trace.TraceIDFromContext(ctx)

			cartItem, ok := cartItemMap[cartItemId]
			if !ok {
				result := types.CartValidateResult{
					Id:           cartItemId,
					Valid:        false,
					ErrorCode:    ErrCodeCartProductNotFound,
					ErrorMessage: "商品不存在",
					TraceId:      traceId,
				}
				resultMu.Lock()
				results[idx] = result
				resultMu.Unlock()
				return nil
			}

			result := l.validateSingleItem(ctx, cartItem, pmsScope)
			result.TraceId = traceId // MEDIUM-7: 补填 traceId
			resultMu.Lock()
			results[idx] = result
			resultMu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, errorx.NewDefaultError(ErrCodeCartSystemError)
	}

	// MEDIUM-7: 填充链路追踪字段
	requestId := trace.TraceIDFromContext(l.ctx)
	serverTime := time.Now().UnixMilli()

	return &types.CartValidateResp{
		Code:       0,
		Message:    "操作成功",
		RequestId:  requestId,
		ServerTime: serverTime,
		Data:       results,
	}, nil
}

// validateSingleItem 校验单个购物车项的商品状态
func (l *ValidateCartItemsLogic) validateSingleItem(ctx context.Context, item *omsclient.CartItemData, pmsScope *pmsclient.GovernanceScope) types.CartValidateResult {
	result := types.CartValidateResult{
		Id:    item.Id,
		Valid: true,
	}

	// ① 查询 SPU 状态
	spuResp, err := l.svcCtx.ProductSpuService.QueryProductSpuDetail(ctx, &pmsclient.QueryProductSpuDetailReq{
		Id:    item.ProductId,
		Scope: pmsScope,
	})
	if err != nil {
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询SPU失败, productId=%d, err=%s", item.ProductId, err.Error())
			result.Valid = false
			result.ErrorCode = ErrCodeCartSystemError
			result.ErrorMessage = "系统异常，请稍后重试"
			return result
		}
		s, _ := status.FromError(err)
		if strings.Contains(s.Message(), "商品SPU不存在") || strings.Contains(s.Message(), "商品不存在") {
			result.Valid = false
			result.ErrorCode = ErrCodeCartProductNotFound
			result.ErrorMessage = "商品不存在"
			return result
		}
		logc.Errorf(l.ctx, "查询SPU详情失败, productId=%d, err=%s", item.ProductId, err.Error())
		result.Valid = false
		result.ErrorCode = ErrCodeCartSystemError
		result.ErrorMessage = "系统异常，请稍后重试"
		return result
	}

	if spuResp.Data == nil {
		result.Valid = false
		result.ErrorCode = ErrCodeCartProductNotFound
		result.ErrorMessage = "商品不存在"
		return result
	}

	// 检查 publish_status
	if spuResp.Data.PublishStatus != 1 {
		result.Valid = false
		result.ErrorCode = ErrCodeCartProductOffline
		result.ErrorMessage = "商品已下架"
		return result
	}

	// 检查 verify_status
	if spuResp.Data.VerifyStatus != 1 {
		result.Valid = false
		result.ErrorCode = ErrCodeCartProductUnverified
		result.ErrorMessage = "商品审核中"
		return result
	}

	// ② 查询 SKU 库存
	skuResp, err := l.svcCtx.ProductSkuService.QueryProductSkuDetail(ctx, &pmsclient.QueryProductSkuDetailReq{
		Id:    item.ProductSkuId,
		Scope: pmsScope,
	})
	if err != nil {
		if isPMSError(err) {
			logc.Errorf(l.ctx, "PMS服务查询SKU失败, productSkuId=%d, err=%s", item.ProductSkuId, err.Error())
			result.Valid = false
			result.ErrorCode = ErrCodeCartSystemError
			result.ErrorMessage = "系统异常，请稍后重试"
			return result
		}
		s, _ := status.FromError(err)
		if strings.Contains(s.Message(), "商品SKU不存在") || strings.Contains(s.Message(), "不存在") {
			result.Valid = false
			result.ErrorCode = ErrCodeCartProductNotFound
			result.ErrorMessage = "商品不存在"
			return result
		}
		result.Valid = false
		result.ErrorCode = ErrCodeCartSystemError
		result.ErrorMessage = "系统异常，请稍后重试"
		return result
	}

	if skuResp == nil {
		result.Valid = false
		result.ErrorCode = ErrCodeCartProductNotFound
		result.ErrorMessage = "商品不存在"
		return result
	}

	// ③ 检查库存（与 update_cart_item_quantity_logic.go 保持一致，使用模板）
	if int32(skuResp.Stock) < item.Quantity {
		result.Valid = false
		result.ErrorCode = ErrCodeCartStockInsufficient
		result.ErrorMessage = fmt.Sprintf(ErrMsgStockInsufficientTpl, skuResp.Stock)
		return result
	}

	return result
}
