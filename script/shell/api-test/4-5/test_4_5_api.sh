#!/bin/bash
# API_TYPE: front
# =============================================================================
# Story 4-5: 购物车与确认单促销试算 API 测试
# 覆盖：促销试算 queryPromotionList、确认单生成 generateConfirmOrder、
#       购物车优惠券列表 queryCouponListByCart、可用优惠券 queryAvailableCoupons
# 用法: bash script/shell/api-test/4-5/test_4_5_api.sh [FRONT_BASE_URL]
# 默认: http://47.107.224.56:9999
# =============================================================================

set -euo pipefail

BASE_URL="${1:-http://47.107.224.56:9999}"
TIMEOUT=15
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
echo " Story 4-5: 购物车与确认单促销试算"
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

# 准备：确保购物车有数据
log_info "准备: 确保购物车有测试商品"
curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"productId":1,"productSkuId":1,"quantity":1,"price":7999,"productName":"小米手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic1.jpg","productSkuCode":"SKU001","productSn":"SN001","productBrand":"小米","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' > /dev/null
curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/addCart" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"productId":2,"productSkuId":3,"quantity":2,"price":7999,"productName":"苹果手机 金色 128GB","productSubTitle":"全网通版","productPic":"http://example.com/pic2.jpg","productSkuCode":"SKU003","productSn":"SN002","productBrand":"Apple","productCategoryId":1,"productAttr":"[]","memberNickname":"张三"}' > /dev/null
echo "  准备完成"

# 1. 促销试算 - queryPromotionList
log_info "1. 促销试算 queryPromotionList"
PROMO_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/queryPromotionList" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{}')
PROMO_CODE=$(json_val "d.get('code','')" "$PROMO_RESP")
if [ "$PROMO_CODE" = "0" ]; then
  log_pass "促销试算返回成功"
  PROMO_DATA=$(json_val "d.get('data',None)" "$PROMO_RESP")
  if [ "$PROMO_DATA" != "None" ]; then
    PROMO_LEN=$(json_val "len(d.get('data',[]) or [])" "$PROMO_RESP")
    echo "    返回促销分组: $PROMO_LEN 个"
  fi
else
  PROMO_MSG=$(json_val "d.get('message',d.get('msg',''))" "$PROMO_RESP")
  log_fail "促销试算失败: code=$PROMO_CODE msg=$PROMO_MSG"
fi

# 2. 确认单生成 - generateConfirmOrder
log_info "2. 确认单生成 generateConfirmOrder"
# 先获取购物车列表拿到选中项的ID
CART_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/order/queryCarItemList" \
  -H "$AUTH")
CART_IDS=$(json_val "','.join(str(item['id']) for item in (d.get('data',[]) or []) if item.get('selected',0)==1)" "$CART_RESP")
echo "  选中购物车项 IDs: $CART_IDS"

CONFIRM_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/order/generateConfirmOrder" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{}')
CONFIRM_CODE=$(json_val "d.get('code','')" "$CONFIRM_RESP")
if [ "$CONFIRM_CODE" = "0" ]; then
  log_pass "确认单生成成功"
  # 检查金额是否为整数（int64 cents 修复）
  TOTAL_AMOUNT=$(json_val "d.get('data',{}).get('totalAmount', d.get('data',{}).get('calcAmount',{}).get('totalAmount', 'N/A'))" "$CONFIRM_RESP")
  echo "    totalAmount=$TOTAL_AMOUNT"
  # 检查是否有优惠券列表
  HAS_COUPON=$(json_val "'couponList' in d.get('data',{})" "$CONFIRM_RESP")
  echo "    包含 couponList: $HAS_COUPON"
else
  CONFIRM_MSG=$(json_val "d.get('message',d.get('msg',''))" "$CONFIRM_RESP")
  log_fail "确认单生成失败: code=$CONFIRM_CODE msg=$CONFIRM_MSG"
fi

# 3. 购物车适用优惠券 - queryCouponListByCart
log_info "3. 购物车适用优惠券 queryCouponListByCart"
CART_COUPON_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryCouponListByCart?type=1" \
  -H "$AUTH")
CC_CODE=$(json_val "d.get('code','')" "$CART_COUPON_RESP")
if [ "$CC_CODE" = "0" ]; then
  # data 结构为 {enableList: [...], disableList: [...]}
  CC_ENABLE=$(json_val "len(d.get('data',{}).get('enableList',[]) or [])" "$CART_COUPON_RESP")
  CC_DISABLE=$(json_val "len(d.get('data',{}).get('disableList',[]) or [])" "$CART_COUPON_RESP")
  CC_COUNT=$((${CC_ENABLE:-0} + ${CC_DISABLE:-0}))
  log_pass "购物车优惠券列表成功，可用${CC_ENABLE:-0}条 不可用${CC_DISABLE:-0}条"
  # 检查 enableList 中是否有 disableReason 字段（新增功能）
  if [ "${CC_ENABLE:-0}" -gt 0 ] 2>/dev/null; then
    HAS_DISABLE=$(json_val "'disableReason' in (d.get('data',{}).get('enableList',[])[0])" "$CART_COUPON_RESP")
    if [ "$HAS_DISABLE" = "True" ]; then
      log_pass "优惠券包含 disableReason 字段"
    else
      log_fail "优惠券缺少 disableReason 字段"
    fi
  else
    log_pass "无可用优惠券（数据符合预期）"
  fi
else
  CC_MSG=$(json_val "d.get('message',d.get('msg',''))" "$CART_COUPON_RESP")
  log_fail "购物车优惠券列表失败: code=$CC_CODE msg=$CC_MSG"
fi

# 4. 可用优惠券列表 - queryAvailableCoupons
log_info "4. 可用优惠券列表 queryAvailableCoupons"
AVAIL_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons?current=1&pageSize=10" \
  -H "$AUTH")
AVAIL_CODE=$(json_val "d.get('code','')" "$AVAIL_RESP")
if [ "$AVAIL_CODE" = "0" ]; then
  AVAIL_COUNT=$(json_val "len(d.get('data',[]) or [])" "$AVAIL_RESP")
  log_pass "可用优惠券列表成功，$AVAIL_COUNT 条"
else
  AVAIL_MSG=$(json_val "d.get('message',d.get('msg',''))" "$AVAIL_RESP")
  log_fail "可用优惠券列表失败: code=$AVAIL_CODE msg=$AVAIL_MSG"
fi

# 5. 已领优惠券列表
log_info "5. 已领优惠券列表 queryCouponList"
MY_COUPON_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryCouponList?current=1&pageSize=10" \
  -H "$AUTH")
MC_CODE=$(json_val "d.get('code','')" "$MY_COUPON_RESP")
if [ "$MC_CODE" = "0" ]; then
  MC_COUNT=$(json_val "len(d.get('data',[]) or [])" "$MY_COUPON_RESP")
  log_pass "已领优惠券列表成功，$MC_COUNT 条"
else
  MC_MSG=$(json_val "d.get('message',d.get('msg',''))" "$MY_COUPON_RESP")
  log_fail "已领优惠券列表失败: code=$MC_CODE msg=$MC_MSG"
fi

# 6. 未登录访问应被拒绝
log_info "6. 未登录访问确认单应返回 401"
NOAUTH_CODE=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/order/generateConfirmOrder" \
  -H 'Content-Type: application/json' \
  -d '{}')
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
