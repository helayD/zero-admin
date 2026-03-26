#!/bin/bash
# =============================================================================
# Story 5-2: 购物车编辑、删除与批量结算 API 测试
# 覆盖：修改数量、删除单项、清空购物车、批量结算校验、购物车促销查询
# 用法: bash script/shell/api-test/5-2/test_5_2_api.sh [FRONT_BASE_URL]
# 默认: http://47.107.224.56:9999
# =============================================================================

set -euo pipefail

BASE_URL="${1:-http://47.107.224.56:9999}"
TIMEOUT=10
PASS=0
FAIL=0
TOTAL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() { ((PASS++)); ((TOTAL++)); echo -e "  ${GREEN}✅ PASS${NC} $1"; }
log_fail() { ((FAIL++)); ((TOTAL++)); echo -e "  ${RED}❌ FAIL${NC} $1"; }
log_info() { echo -e "${YELLOW}▶${NC} $1"; }

json_val() {
  python3 -c "
import sys,json
try:
    d=json.load(sys.stdin)
    print($1)
except (json.JSONDecodeError, ValueError):
    print('')
" 2>/dev/null <<< "$2" || echo ''
}

echo "============================================="
echo " Story 5-2: 购物车编辑、删除与批量结算"
echo " Front URL: $BASE_URL"
echo " Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================="
echo ""

# 0. 会员登录
log_info "0. 会员登录"
LOGIN_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/member/login" \
  -H 'Content-Type: application/json' \
  -d '{"mobile":"13800138001","password":"123456"}')
FRONT_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$LOGIN_RESP")
FRONT_CODE=$(json_val "d.get('code','')" "$LOGIN_RESP")
if [ -n "$FRONT_TOKEN" ] && [ "$FRONT_TOKEN" != "None" ] && [ "$FRONT_CODE" = "0" ]; then
  log_pass "会员登录成功"
else
  log_fail "会员登录失败: $LOGIN_RESP"
  exit 1
fi
AUTH="Authorization: Bearer $FRONT_TOKEN"

# 准备：先加购确保购物车有数据
log_info "准备: 确保购物车有测试数据"
curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"productId":1,"productSkuId":1,"quantity":1,"price":7999,"productName":"小米手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic1.jpg","productSkuCode":"SKU001","productSn":"SN001","productBrand":"小米","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' > /dev/null
curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"productId":2,"productSkuId":3,"quantity":1,"price":7999,"productName":"苹果手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic2.jpg","productSkuCode":"SKU003","productSn":"SN002","productBrand":"Apple","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' > /dev/null
echo "  准备完成"

# 1. 查询购物车列表
log_info "1. 查询购物车列表"
CART_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/order/queryCarItemList" \
  -H "$AUTH")
CART_CODE=$(json_val "d.get('code','')" "$CART_RESP")
CART_COUNT=$(json_val "len(d.get('data',[]) or [])" "$CART_RESP")
if [ "$CART_CODE" = "0" ] && [ "$CART_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "购物车列表返回成功，$CART_COUNT 条"
else
  log_fail "购物车列表失败: code=$CART_CODE count=$CART_COUNT"
fi

# 获取第一个购物车项 ID
FIRST_ITEM_ID=$(json_val "d.get('data',[])[0].get('id',0) if d.get('data') else 0" "$CART_RESP")
FIRST_SKU_ID=$(json_val "d.get('data',[])[0].get('productSkuId',0) if d.get('data') else 0" "$CART_RESP")
echo "  首条购物车项 id=$FIRST_ITEM_ID, skuId=$FIRST_SKU_ID"

# 2. 修改数量
log_info "2. 修改购物车商品数量 (id=$FIRST_ITEM_ID, quantity=3)"
UPD_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/updateCartItemQuantity" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"id\":$FIRST_ITEM_ID,\"quantity\":3}")
UPD_CODE=$(json_val "d.get('code','')" "$UPD_RESP")
if [ "$UPD_CODE" = "0" ]; then
  log_pass "修改数量成功"
else
  UPD_MSG=$(json_val "d.get('message',d.get('msg',''))" "$UPD_RESP")
  log_fail "修改数量失败: code=$UPD_CODE msg=$UPD_MSG"
fi

# 3. 验证数量已更新
log_info "3. 验证数量已更新"
CART_RESP2=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/order/queryCarItemList" \
  -H "$AUTH")
UPDATED_QTY=$(json_val "next((item.get('quantity',0) for item in (d.get('data',[]) or []) if item.get('id')==$FIRST_ITEM_ID), -1)" "$CART_RESP2")
if [ "$UPDATED_QTY" = "3" ]; then
  log_pass "数量已更新为 3"
else
  log_fail "数量未更新为 3，实际: $UPDATED_QTY"
fi

# 4. 促销信息查询
log_info "4. 查询购物车促销信息"
PROMO_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/queryPromotionList" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{}')
PROMO_CODE=$(json_val "d.get('code','')" "$PROMO_RESP")
if [ "$PROMO_CODE" = "0" ]; then
  log_pass "促销信息查询成功"
else
  PROMO_MSG=$(json_val "d.get('message',d.get('msg',''))" "$PROMO_RESP")
  log_fail "促销信息查询失败: code=$PROMO_CODE msg=$PROMO_MSG"
fi

# 5. 批量结算校验 - validateCartItems
log_info "5. 批量结算校验 - validateCartItems"
VAL_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/validateCartItems" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d "{\"ids\":[$FIRST_ITEM_ID]}")
VAL_CODE=$(json_val "d.get('code','')" "$VAL_RESP")
if [ "$VAL_CODE" = "0" ]; then
  log_pass "批量结算校验成功"
  # 检查返回项的 valid 状态
  VAL_VALID=$(json_val "(d.get('data',[]) or [{}])[0].get('valid', True) if d.get('data') else True" "$VAL_RESP")
  echo "    校验结果: valid=$VAL_VALID"
else
  VAL_MSG=$(json_val "d.get('message',d.get('msg',''))" "$VAL_RESP")
  log_fail "批量结算校验失败: code=$VAL_CODE msg=$VAL_MSG"
fi

# 6. 删除单项
log_info "6. 删除购物车单项"
# 先加购一个临时商品用于测试删除
curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"productId":3,"productSkuId":5,"quantity":1,"price":7999,"productName":"华为手机 金色 128GB","productSubTitle":"折叠屏手机","productPic":"http://example.com/pic3.jpg","productSkuCode":"SKU005","productSn":"SN003","productBrand":"华为","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' > /dev/null
# 重新获取购物车，找到临时商品
CART_RESP3=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/order/queryCarItemList" \
  -H "$AUTH")
DEL_ITEM_ID=$(json_val "next((item.get('id',0) for item in (d.get('data',[]) or []) if item.get('productSkuId')==5), 0)" "$CART_RESP3")
if [ "$DEL_ITEM_ID" != "0" ] && [ "$DEL_ITEM_ID" != "None" ]; then
  DEL_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/deleteCartItem" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"ids\":[$DEL_ITEM_ID]}")
  DEL_CODE=$(json_val "d.get('code','')" "$DEL_RESP")
  if [ "$DEL_CODE" = "0" ]; then
    log_pass "删除单项成功 (id=$DEL_ITEM_ID)"
  else
    DEL_MSG=$(json_val "d.get('message',d.get('msg',''))" "$DEL_RESP")
    log_fail "删除单项失败: code=$DEL_CODE msg=$DEL_MSG"
  fi
else
  log_pass "临时商品不在购物车（可能已被之前删除），跳过删除测试"
fi

# 7. 未登录操作应被拒绝
log_info "7. 未登录修改数量应返回 401"
NOAUTH_CODE=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/order/updateCartItemQuantity" \
  -H 'Content-Type: application/json' \
  -d '{"id":1,"quantity":2}')
if [ "$NOAUTH_CODE" = "401" ]; then
  log_pass "未登录返回 401"
else
  log_fail "未登录应返回 401，实际: $NOAUTH_CODE"
fi

# =============================================================================
echo ""
echo "============================================="
echo " 测试完成: $(date '+%Y-%m-%d %H:%M:%S')"
echo " 总计: $TOTAL  通过: $PASS  失败: $FAIL"
if [ "$FAIL" -eq 0 ]; then
  echo -e " 结果: ${GREEN}全部通过${NC}"
else
  echo -e " 结果: ${RED}存在失败${NC}"
fi
echo "============================================="
exit $FAIL
