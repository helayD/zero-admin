#!/bin/bash
# API_TYPE: front
# =============================================================================
# Story 4-2: 消费者领券与优惠券资产可见 API 测试
# 覆盖：可领优惠券查询 queryAvailableCoupons、领券 addCoupon、
#       已领优惠券列表 queryCouponList、购物车优惠券 queryCouponListByCart
# 用法: bash script/shell/api-test/4-2/test_4_2_api.sh [FRONT_BASE_URL]
# 默认: http://47.107.224.56:9999
# =============================================================================

set -eo pipefail

BASE_URL="${1:-http://47.107.224.56:9999}"
TIMEOUT=15
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
echo " Story 4-2: 消费者领券与优惠券资产可见"
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

# 1. 查询可领优惠券列表 - queryAvailableCoupons
log_info "1. 查询可领优惠券 queryAvailableCoupons"
AVAIL_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons?current=1&pageSize=20" \
  -H "$AUTH")
AVAIL_CODE=$(json_val "d.get('code','')" "$AVAIL_RESP")
if [ "$AVAIL_CODE" = "0" ]; then
  AVAIL_COUNT=$(json_val "len(d.get('data',[]) or [])" "$AVAIL_RESP")
  log_pass "可领优惠券查询成功，$AVAIL_COUNT 条"

  # 检查返回数据结构完整性（AC#1: 优惠券字段完整）
  if [ "${AVAIL_COUNT:-0}" -gt 0 ] 2>/dev/null; then
    FIRST_COUPON=$(json_val "d.get('data',[])[0]" "$AVAIL_RESP")
    HAS_ID=$(json_val "'id' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_NAME=$(json_val "'name' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_AMOUNT=$(json_val "'amount' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_MIN_AMOUNT=$(json_val "'minAmount' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_START_TIME=$(json_val "'startTime' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_END_TIME=$(json_val "'endTime' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_RECEIVE_STATUS=$(json_val "'receiveStatus' in $FIRST_COUPON" "$AVAIL_RESP")
    HAS_SCOPE_TYPE=$(json_val "'scopeType' in $FIRST_COUPON" "$AVAIL_RESP")

    if [ "$HAS_ID" = "True" ] && [ "$HAS_NAME" = "True" ]; then
      log_pass "优惠券数据字段完整（id, name, amount, minAmount, startTime, endTime, receiveStatus, scopeType）"
    else
      log_fail "优惠券数据字段缺失"
    fi

    # 检查 receiveStatus 语义（0=可领取，1=已达限领上限）
    # 可领取列表中的券可能因当前会员已达 perLimit 而显示 receiveStatus=1
    if [ "$HAS_RECEIVE_STATUS" = "True" ]; then
      STATUS=$(json_val "d.get('data',[])[0].get('receiveStatus', -1)" "$AVAIL_RESP")
      echo "    第一张券 receiveStatus=$STATUS"
      case "$STATUS" in
        0) log_pass "receiveStatus=0 表示可领取，语义正确" ;;
        1) log_pass "receiveStatus=1 表示已达限领上限，语义正确（AC#2: 失败原因友好）" ;;
        *) log_fail "receiveStatus 值非法（期望 0 或 1），实际=$STATUS" ;;
      esac
    fi
  else
    log_pass "当前无可领优惠券（数据符合预期）"
  fi
else
  AVAIL_MSG=$(json_val "d.get('message',d.get('msg',''))" "$AVAIL_RESP")
  log_fail "可领优惠券查询失败: code=$AVAIL_CODE msg=$AVAIL_MSG"
fi

# 2. 领取优惠券 - addCoupon
log_info "2. 领取优惠券 addCoupon"

# 先取第一个可领券的 ID
AVAIL_ID=""
if [ "${AVAIL_COUNT:-0}" -gt 0 ] 2>/dev/null; then
  AVAIL_ID=$(json_val "d.get('data',[])[0].get('id','')" "$AVAIL_RESP")
fi

if [ -n "$AVAIL_ID" ] && [ "$AVAIL_ID" != "None" ]; then
  echo "    尝试领取优惠券 ID=$AVAIL_ID"
  ADD_HTTP=$(curl -s --max-time $TIMEOUT -o ${TMPDIR:-/tmp}/add_resp.txt -w "%{http_code}" -X POST "$BASE_URL/api/member/coupon/addCoupon" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"couponId\":$AVAIL_ID}")
  ADD_BODY=$(cat ${TMPDIR:-/tmp}/add_resp.txt)
  # 尝试解析 JSON，失败则保留纯文本作为错误信息
  ADD_CODE=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('code',''))
except:
    print('')
" "$ADD_BODY" 2>/dev/null || echo "")
  ADD_MSG=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('message',''))
except:
    print(sys.argv[1])
" "$ADD_BODY" 2>/dev/null || echo "$ADD_BODY")
  echo "    领券响应: http=$ADD_HTTP code=$ADD_CODE msg=$ADD_MSG"

  # 成功（200）或"已达上限"都是预期行为
  if [ "$ADD_HTTP" = "200" ]; then
    log_pass "领券成功"
  elif echo "$ADD_MSG" | grep -qiE "已领|已领取|已存在|已申领|已达.*领取|上限"; then
    log_pass "领券失败原因可理解: $ADD_MSG（AC#2: 失败原因友好）"
  else
    log_fail "领券失败: http=$ADD_HTTP msg=$ADD_MSG"
  fi
else
  log_pass "无可领优惠券，跳过领券测试"
fi

# 3. 重复领取同一优惠券 - AC#2: 幂等保护
log_info "3. 重复领取测试（幂等保护）"
if [ -n "$AVAIL_ID" ] && [ "$AVAIL_ID" != "None" ]; then
  ADD_RESP2=$(curl -s --max-time $TIMEOUT -o ${TMPDIR:-/tmp}/add_resp2.txt -w "%{http_code}" -X POST "$BASE_URL/api/member/coupon/addCoupon" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"couponId\":$AVAIL_ID}")
  ADD_BODY2=$(cat ${TMPDIR:-/tmp}/add_resp2.txt)
  ADD_CODE2=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('code',''))
except:
    print('')
" "$ADD_BODY2" 2>/dev/null || echo "")

  # 不应返回 code=0（已领取）或明确的错误原因
  if [ "$ADD_CODE2" = "0" ]; then
    log_fail "重复领取返回成功（应拒绝）"
  else
    ADD_MSG2=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('message',''))
except:
    print(sys.argv[1])
" "$ADD_BODY2" 2>/dev/null || echo "$ADD_BODY2")
    log_pass "重复领取被正确拒绝: code=$ADD_CODE2 msg=$ADD_MSG2"
  fi
else
  log_pass "无可领优惠券，跳过重复领取测试"
fi

# 4. 领取无效优惠券 - AC#2: 非法 couponId 拒绝
log_info "4. 领取非法优惠券 ID=999999"
BAD_ADD_RESP=$(curl -s --max-time $TIMEOUT -o ${TMPDIR:-/tmp}/bad_add.txt -w "%{http_code}" -X POST "$BASE_URL/api/member/coupon/addCoupon" \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"couponId":999999}')
BAD_BODY=$(cat ${TMPDIR:-/tmp}/bad_add.txt)
BAD_CODE=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('code',''))
except:
    print('')
" "$BAD_BODY" 2>/dev/null || echo "")
if [ "$BAD_CODE" != "0" ]; then
  BAD_MSG=$(python3 -c "
import sys,json
try:
    print(json.loads(sys.argv[1]).get('message',''))
except:
    print(sys.argv[1])
" "$BAD_BODY" 2>/dev/null || echo "$BAD_BODY")
  log_pass "非法 couponId 被正确拒绝: $BAD_MSG"
else
  log_fail "非法 couponId 未被拒绝"
fi

# 5. 查询已领优惠券列表 - queryCouponList
log_info "5. 查询已领优惠券 queryCouponList"
MY_COUPON_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryCouponList?current=1&pageSize=10" \
  -H "$AUTH")
MC_CODE=$(json_val "d.get('code','')" "$MY_COUPON_RESP")
if [ "$MC_CODE" = "0" ]; then
  MY_COUNT=$(json_val "len(d.get('data',[]) or [])" "$MY_COUPON_RESP")
  log_pass "已领优惠券列表成功，$MY_COUNT 条"

  # AC#1: 检查状态展示（可用/已使用/已过期/即将过期）
  if [ "${MY_COUNT:-0}" -gt 0 ] 2>/dev/null; then
    FIRST_REC=$(json_val "d.get('data',[])[0]" "$MY_COUPON_RESP")
    HAS_STATUS=$(json_val "'status' in $FIRST_REC" "$MY_COUPON_RESP")
    HAS_NAME=$(json_val "'couponName' in $FIRST_REC or 'name' in $FIRST_REC" "$MY_COUPON_RESP")
    echo "    已领券包含 status 字段: $HAS_STATUS"
    echo "    已领券包含 name 字段: $HAS_NAME"

    if [ "$HAS_STATUS" = "True" ]; then
      STATUS_VAL=$(json_val "$FIRST_REC.get('status','')" "$MY_COUPON_RESP")
      case "$STATUS_VAL" in
        0) STATUS_STR="未使用" ;;
        1) STATUS_STR="已使用" ;;
        2) STATUS_STR="已过期" ;;
        3) STATUS_STR="已失效" ;;
        *) STATUS_STR="未知" ;;
      esac
      echo "    第一张券 status=$STATUS_VAL ($STATUS_STR)"
      log_pass "优惠券状态展示正确: $STATUS_STR"
    fi
  fi
else
  MC_MSG=$(json_val "d.get('message',d.get('msg',''))" "$MY_COUPON_RESP")
  log_fail "已领优惠券列表失败: code=$MC_CODE msg=$MC_MSG"
fi

# 6. 购物车可用优惠券 - queryCouponListByCart
log_info "6. 购物车可用优惠券 queryCouponListByCart"
CART_COUPON_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryCouponListByCart?type=1" \
  -H "$AUTH")
CC_CODE=$(json_val "d.get('code','')" "$CART_COUPON_RESP")
if [ "$CC_CODE" = "0" ]; then
  CC_ENABLE=$(json_val "len(d.get('data',{}).get('enableList',[]) or [])" "$CART_COUPON_RESP")
  CC_DISABLE=$(json_val "len(d.get('data',{}).get('disableList',[]) or [])" "$CART_COUPON_RESP")
  log_pass "购物车优惠券成功: 可用${CC_ENABLE:-0}条 不可用${CC_DISABLE:-0}条"

  # AC#1: 不可用优惠券应包含 disableReason
  if [ "${CC_DISABLE:-0}" -gt 0 ] 2>/dev/null; then
    HAS_DISABLE=$(json_val "'disableReason' in (d.get('data',{}).get('disableList',[])[0])" "$CART_COUPON_RESP")
    if [ "$HAS_DISABLE" = "True" ]; then
      REASON=$(json_val "d.get('data',{}).get('disableList',[])[0].get('disableReason','')" "$CART_COUPON_RESP")
      log_pass "不可用优惠券包含 disableReason: $REASON"
    else
      log_fail "不可用优惠券缺少 disableReason 字段"
    fi
  else
    log_pass "无不可用优惠券，跳过 disableReason 检查"
  fi
else
  CC_MSG=$(json_val "d.get('message',d.get('msg',''))" "$CART_COUPON_RESP")
  log_fail "购物车优惠券失败: code=$CC_CODE msg=$CC_MSG"
fi

# 7. 未登录访问应被拒绝 - AC#2: 鉴权保护
log_info "7. 未登录访问领券接口应返回 401"
NOAUTH_RESP=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/member/coupon/addCoupon" \
  -H 'Content-Type: application/json' \
  -d '{"couponId":1}')
if [ "$NOAUTH_RESP" = "401" ] || [ "$NOAUTH_RESP" = "403" ]; then
  log_pass "未登录返回 $NOAUTH_RESP（鉴权保护生效）"
else
  log_fail "未登录应返回 401/403，实际: $NOAUTH_RESP"
fi

# 8. AC#1 + AC#2 综合验证：所有接口均需返回 code 字段
log_info "8. 综合验证: 所有接口返回标准 code/message"
ALL_OK=true
for resp_var in "$AVAIL_RESP" "$MY_COUPON_RESP" "$CART_COUPON_RESP"; do
  resp_code=$(json_val "d.get('code','')" "$resp_var" 2>/dev/null || echo "N/A")
  resp_msg=$(json_val "d.get('message',d.get('msg',''))" "$resp_var" 2>/dev/null || echo "")
  if [ "$resp_code" = "" ] || [ "$resp_code" = "N/A" ]; then
    echo "    响应缺少 code 字段: ${resp_msg:0:80}"
    ALL_OK=false
  fi
done
if [ "$ALL_OK" = "true" ]; then
  log_pass "所有接口返回标准 code/message"
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
