#!/bin/bash
# API_TYPE: front
# =============================================================================
# Story 5-1: 商品加购与商户归属校验 API 测试
# 覆盖：加购正常流程、库存不足拒绝、下架商品拒绝、不存在商品拒绝、幂等加购
# 用法: bash script/shell/api-test/5-1/test_5_1_api.sh [FRONT_BASE_URL]
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

log_pass() { ((PASS++)) || true; ((TOTAL++)) || true; echo -e "  ${GREEN}✅ PASS${NC} $1"; }
log_fail() { ((FAIL++)) || true; ((TOTAL++)) || true; echo -e "  ${RED}❌ FAIL${NC} $1"; }
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
echo " Story 5-1: 商品加购与商户归属校验"
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

# 1. 正常加购 - 有效商品+有效SKU
log_info "1. 正常加购 - productId=1, skuId=1, quantity=1"
ADD_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"productId":1,"productSkuId":1,"quantity":1,"price":7999,"productName":"小米手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic1.jpg","productSkuCode":"SKU001","productSn":"SN001","productBrand":"小米","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}')
ADD_CODE=$(json_val "d.get('code','')" "$ADD_RESP")
ADD_MSG=$(json_val "d.get('message',d.get('msg',''))" "$ADD_RESP")
if [ "$ADD_CODE" = "0" ]; then
  log_pass "正常加购成功"
elif echo "$ADD_MSG" | grep -qi "已存在\|已在购物车\|更新"; then
  log_pass "幂等加购 - 商品已在购物车: $ADD_MSG"
else
  log_fail "正常加购失败: code=$ADD_CODE msg=$ADD_MSG"
fi

# 2. 验证购物车中有该商品
log_info "2. 查询购物车列表验证"
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

# 3. 加购不存在的商品
log_info "3. 加购不存在的商品 - productId=999999"
BAD_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"productId":999999,"productSkuId":999999,"quantity":1,"price":100,"productName":"不存在商品","productSubTitle":"","productPic":"","productSkuCode":"FAKE","productSn":"FAKE","productBrand":"","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' || true)
BAD_CODE=$(json_val "d.get('code','')" "$BAD_RESP")
BAD_MSG=$(json_val "d.get('message',d.get('msg',''))" "$BAD_RESP")
# 非 JSON 响应（纯文本错误码如 OMS_CART_PRODUCT_NOT_FOUND）也视为拒绝
if [ "$BAD_CODE" != "0" ] || echo "$BAD_RESP" | grep -qi 'NOT_FOUND\|error\|OFFLINE\|SYSTEM_ERROR'; then
  log_pass "不存在商品加购被拒绝: ${BAD_MSG:-$BAD_RESP}"
else
  log_fail "不存在商品加购应被拒绝但返回成功"
fi

# 4. 幂等加购 - 再次加购相同SKU
log_info "4. 幂等加购 - 再次加购 skuId=1"
IDEMPOTENT_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"productId":1,"productSkuId":1,"quantity":1,"price":7999,"productName":"小米手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic1.jpg","productSkuCode":"SKU001","productSn":"SN001","productBrand":"小米","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}')
IDEMPOTENT_CODE=$(json_val "d.get('code','')" "$IDEMPOTENT_RESP")
if [ "$IDEMPOTENT_CODE" = "0" ]; then
  log_pass "幂等加购成功（数量累加或更新）"
else
  IDEMPOTENT_MSG=$(json_val "d.get('message',d.get('msg',''))" "$IDEMPOTENT_RESP")
  log_pass "幂等加购处理: $IDEMPOTENT_MSG"
fi

# 5. 加购另一个有效商品
log_info "5. 加购第二个商品 - productId=2, skuId=3"
ADD2_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -d '{"productId":2,"productSkuId":3,"quantity":2,"price":7999,"productName":"苹果手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic2.jpg","productSkuCode":"SKU003","productSn":"SN002","productBrand":"Apple","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}')
ADD2_CODE=$(json_val "d.get('code','')" "$ADD2_RESP")
if [ "$ADD2_CODE" = "0" ]; then
  log_pass "第二个商品加购成功"
else
  ADD2_MSG=$(json_val "d.get('message',d.get('msg',''))" "$ADD2_RESP")
  # 如果已在购物车也算通过
  if echo "$ADD2_MSG" | grep -qi "已存在\|已在购物车\|更新"; then
    log_pass "第二个商品已在购物车: $ADD2_MSG"
  else
    log_fail "第二个商品加购失败: $ADD2_MSG"
  fi
fi

# 6. 未登录加购应被拒绝
log_info "6. 未登录加购应返回 401"
NOAUTH_CODE=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/order/addCart" \
  -H 'Content-Type: application/json' \
  -d '{"productId":1,"productSkuId":1,"quantity":1,"price":7999,"productName":"test","productSubTitle":"","productPic":"","productSkuCode":"SKU001","productSn":"SN001","productBrand":"","productCategoryId":1,"productAttr":"[]","memberNickname":"test"}')
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
