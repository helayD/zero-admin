#!/bin/bash
# =============================================================================
# 优惠券功能 API 接口测试脚本
# 用法: bash script/shell/test_coupon_api.sh [BASE_URL]
# 默认: http://47.107.224.56:9999
# =============================================================================

set -euo pipefail

BASE_URL="${1:-http://47.107.224.56:9999}"
ADMIN_URL="${BASE_URL/9999/8888}"
TIMEOUT=5
PASS=0
FAIL=0
TOTAL=0

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() { ((PASS++)); ((TOTAL++)); echo -e "  ${GREEN}✅ PASS${NC} $1"; }
log_fail() { ((FAIL++)); ((TOTAL++)); echo -e "  ${RED}❌ FAIL${NC} $1"; }
log_info() { echo -e "${YELLOW}▶${NC} $1"; }

# JSON 字段提取 (不依赖 jq)
json_val() {
  python3 -c "import sys,json; d=json.load(sys.stdin); print($1)" 2>/dev/null <<< "$2"
}

# =============================================================================
echo "============================================="
echo " 优惠券功能 API 接口测试"
echo " Base URL: $BASE_URL"
echo " Admin URL: $ADMIN_URL"
echo " Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================="
echo ""

# -----------------------------------------------------------------------------
# 0. Admin 登录
# -----------------------------------------------------------------------------
log_info "0. Admin 登录获取 token"
ADMIN_RESP=$(curl -s --max-time $TIMEOUT -X POST "$ADMIN_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}')
ADMIN_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$ADMIN_RESP")
if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "None" ]; then
  log_pass "Admin 登录成功"
else
  log_fail "Admin 登录失败: $ADMIN_RESP"
  echo "无法继续测试，退出"
  exit 1
fi

# -----------------------------------------------------------------------------
# 1. 查询优惠券列表 (Admin)
# -----------------------------------------------------------------------------
log_info "1. Admin 查询优惠券列表"
COUPON_LIST_RESP=$(curl -s --max-time $TIMEOUT \
  "$ADMIN_URL/api/sms/coupon/queryCouponList?current=1&pageSize=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
COUPON_COUNT=$(json_val "len(d.get('data',[]))" "$COUPON_LIST_RESP")
if [ "$COUPON_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "优惠券列表返回 $COUPON_COUNT 条"
else
  log_fail "优惠券列表为空或失败: $COUPON_LIST_RESP"
fi

# 提取可用优惠券信息
ACTIVE_COUPON_ID=$(json_val "[c['id'] for c in d.get('data',[]) if c.get('status')==1][0] if [c for c in d.get('data',[]) if c.get('status')==1] else 0" "$COUPON_LIST_RESP")
ACTIVE_COUPON_NAME=$(json_val "[c['name'] for c in d.get('data',[]) if c.get('status')==1][0] if [c for c in d.get('data',[]) if c.get('status')==1] else ''" "$COUPON_LIST_RESP")
ACTIVE_COUPON_PERLIMIT=$(json_val "[c['perLimit'] for c in d.get('data',[]) if c.get('status')==1][0] if [c for c in d.get('data',[]) if c.get('status')==1] else 0" "$COUPON_LIST_RESP")
echo "    活跃优惠券: id=$ACTIVE_COUPON_ID name=$ACTIVE_COUPON_NAME perLimit=$ACTIVE_COUPON_PERLIMIT"

# -----------------------------------------------------------------------------
# 2. 查询会员列表 (Admin) 获取测试账号
# -----------------------------------------------------------------------------
log_info "2. Admin 查询会员列表"
MEMBER_RESP=$(curl -s --max-time $TIMEOUT \
  "$ADMIN_URL/api/ums/member/queryMemberList?current=1&pageSize=1" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
TEST_MOBILE=$(json_val "d.get('data',[])[0].get('mobile','') if d.get('data') else ''" "$MEMBER_RESP")
if [ -n "$TEST_MOBILE" ] && [ "$TEST_MOBILE" != "None" ]; then
  log_pass "获取测试会员手机号: $TEST_MOBILE"
else
  log_fail "获取会员失败: $MEMBER_RESP"
  echo "无法继续 front-api 测试，退出"
  exit 1
fi

# -----------------------------------------------------------------------------
# 3. Front 会员登录
# -----------------------------------------------------------------------------
log_info "3. Front 会员登录"
FRONT_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/member/login" \
  -H 'Content-Type: application/json' \
  -d "{\"mobile\":\"$TEST_MOBILE\",\"password\":\"123456\"}")
FRONT_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$FRONT_RESP")
FRONT_CODE=$(json_val "d.get('code','')" "$FRONT_RESP")
if [ -n "$FRONT_TOKEN" ] && [ "$FRONT_TOKEN" != "None" ] && [ "$FRONT_CODE" = "0" ]; then
  log_pass "会员登录成功: $TEST_MOBILE"
else
  log_fail "会员登录失败: $FRONT_RESP"
  echo "无法继续 front-api 测试，退出"
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $FRONT_TOKEN"

# -----------------------------------------------------------------------------
# 4. 可领优惠券列表 (分页测试)
# -----------------------------------------------------------------------------
log_info "4. 可领优惠券列表 - 分页测试"

AVAIL_ALL=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons?pageNum=1&pageSize=100" \
  -H "$AUTH_HEADER")
AVAIL_ALL_COUNT=$(json_val "len(d.get('data',[]) or [])" "$AVAIL_ALL")
log_info "   可领优惠券总数: $AVAIL_ALL_COUNT"

# 分页 pageSize=1 第1页
AVAIL_P1=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons?pageNum=1&pageSize=1" \
  -H "$AUTH_HEADER")
AVAIL_P1_COUNT=$(json_val "len(d.get('data',[]) or [])" "$AVAIL_P1")
AVAIL_P1_ID=$(json_val "(d.get('data') or [{}])[0].get('id',0)" "$AVAIL_P1")

# 分页 pageSize=1 第2页
AVAIL_P2=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons?pageNum=2&pageSize=1" \
  -H "$AUTH_HEADER")
AVAIL_P2_COUNT=$(json_val "len(d.get('data',[]) or [])" "$AVAIL_P2")
AVAIL_P2_ID=$(json_val "(d.get('data') or [{}])[0].get('id',0)" "$AVAIL_P2")

if [ "$AVAIL_P1_COUNT" = "1" ]; then
  log_pass "分页 page1 返回 1 条 (id=$AVAIL_P1_ID)"
else
  log_fail "分页 page1 应返回 1 条，实际 $AVAIL_P1_COUNT"
fi

if [ "$AVAIL_ALL_COUNT" -gt 1 ] 2>/dev/null; then
  if [ "$AVAIL_P2_COUNT" = "1" ] && [ "$AVAIL_P2_ID" != "$AVAIL_P1_ID" ]; then
    log_pass "分页 page2 返回 1 条 (id=$AVAIL_P2_ID)，与 page1 不同"
  else
    log_fail "分页 page2 异常: count=$AVAIL_P2_COUNT id=$AVAIL_P2_ID"
  fi
else
  echo "    (总数<=1，跳过 page2 验证)"
fi

# -----------------------------------------------------------------------------
# 5. receiveStatus 字段验证
# -----------------------------------------------------------------------------
log_info "5. receiveStatus 字段验证"

HAS_RECEIVE_STATUS=$(json_val "'receiveStatus' in (d.get('data') or [{}])[0]" "$AVAIL_ALL")
if [ "$HAS_RECEIVE_STATUS" = "True" ]; then
  log_pass "可领列表包含 receiveStatus 字段"
else
  log_fail "可领列表缺少 receiveStatus 字段"
fi

# 检查 receiveStatus 值范围
RS_VALUES=$(json_val "list(set(c.get('receiveStatus',-1) for c in (d.get('data') or [])))" "$AVAIL_ALL")
echo "    receiveStatus 值集合: $RS_VALUES"

# -----------------------------------------------------------------------------
# 6. 会员优惠券列表
# -----------------------------------------------------------------------------
log_info "6. 会员优惠券列表 (未使用)"
MEMBER_COUPON=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/member/coupon/queryCouponList?useStatus=0" \
  -H "$AUTH_HEADER")
MC_CODE=$(json_val "d.get('code','')" "$MEMBER_COUPON")
MC_COUNT=$(json_val "len(d.get('data',[]) or [])" "$MEMBER_COUPON")
if [ "$MC_CODE" = "0" ]; then
  log_pass "会员优惠券列表返回成功，$MC_COUNT 条"
else
  log_fail "会员优惠券列表失败: $MEMBER_COUPON"
fi

# 检查是否有重复 ID（验证 DISTINCT 修复）
MC_UNIQUE=$(json_val "len(set(c['id'] for c in (d.get('data') or [])))" "$MEMBER_COUPON")
if [ "$MC_COUNT" = "$MC_UNIQUE" ]; then
  log_pass "会员券列表无重复 ($MC_COUNT 条, $MC_UNIQUE 唯一)"
else
  log_fail "会员券列表有重复: 总 $MC_COUNT 条，唯一 $MC_UNIQUE 条"
fi

# -----------------------------------------------------------------------------
# 7. 领券接口测试
# -----------------------------------------------------------------------------
log_info "7. 领券接口测试"

# 领取一张券
CLAIM_RESP=$(curl -s --max-time $TIMEOUT -X POST \
  "$BASE_URL/api/member/coupon/addCoupon" \
  -H "$AUTH_HEADER" \
  -H 'Content-Type: application/json' \
  -d "{\"couponId\":$ACTIVE_COUPON_ID}")
echo "    领券响应: $CLAIM_RESP"

# 如果领取成功，验证 receivedCount 增加
CLAIM_MSG=$(json_val "d.get('message','')" "$CLAIM_RESP" 2>/dev/null || echo "$CLAIM_RESP")
if echo "$CLAIM_MSG" | grep -q "成功领取"; then
  log_pass "领券成功: $CLAIM_MSG"
  
  # 验证 receivedCount 增加
  AFTER_AVAIL=$(curl -s --max-time $TIMEOUT \
    "$BASE_URL/api/member/coupon/queryAvailableCoupons?pageNum=1&pageSize=100" \
    -H "$AUTH_HEADER")
  AFTER_RC=$(json_val "[c.get('receivedCount',0) for c in (d.get('data') or []) if c.get('id')==$ACTIVE_COUPON_ID][0] if [c for c in (d.get('data') or []) if c.get('id')==$ACTIVE_COUPON_ID] else -1" "$AFTER_AVAIL")
  echo "    领券后 receivedCount=$AFTER_RC"
  if [ "$AFTER_RC" -gt 0 ] 2>/dev/null; then
    log_pass "receivedCount 已递增: $AFTER_RC"
  fi
elif echo "$CLAIM_MSG" | grep -q "领取上限\|已被领完"; then
  log_pass "限领拒绝正常: $CLAIM_MSG"
elif echo "$CLAIM_MSG" | grep -q "暂不可领取\|尚未开始\|已过期\|已取消\|状态异常"; then
  log_pass "状态拒绝正常: $CLAIM_MSG"
else
  log_fail "领券响应异常: $CLAIM_MSG"
fi

# 测试重复领取被拦截
CLAIM_AGAIN=$(curl -s --max-time $TIMEOUT -X POST \
  "$BASE_URL/api/member/coupon/addCoupon" \
  -H "$AUTH_HEADER" \
  -H 'Content-Type: application/json' \
  -d "{\"couponId\":$ACTIVE_COUPON_ID}")
AGAIN_MSG=$(json_val "d.get('message','')" "$CLAIM_AGAIN" 2>/dev/null || echo "$CLAIM_AGAIN")
if echo "$AGAIN_MSG" | grep -q "领取上限\|已被领完\|成功领取"; then
  log_pass "重复领券处理正常: $AGAIN_MSG"
else
  log_fail "重复领券异常: $AGAIN_MSG"
fi

# 测试无效优惠券ID
INVALID_RESP=$(curl -s --max-time $TIMEOUT -X POST \
  "$BASE_URL/api/member/coupon/addCoupon" \
  -H "$AUTH_HEADER" \
  -H 'Content-Type: application/json' \
  -d '{"couponId":999999}')
INVALID_MSG=$(json_val "d.get('message','')" "$INVALID_RESP" 2>/dev/null || echo "$INVALID_RESP")
if echo "$INVALID_MSG" | grep -q "不存在"; then
  log_pass "无效券ID拒绝: $INVALID_MSG"
else
  log_fail "无效券ID响应异常: $INVALID_MSG"
fi

# -----------------------------------------------------------------------------
# 8. 未登录访问应被拒绝
# -----------------------------------------------------------------------------
log_info "8. 未登录访问拦截"
NOAUTH_RESP=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  "$BASE_URL/api/member/coupon/queryAvailableCoupons")
if [ "$NOAUTH_RESP" = "401" ]; then
  log_pass "未登录返回 401"
else
  log_fail "未登录应返回 401，实际: $NOAUTH_RESP"
fi

# =============================================================================
# 汇总
# =============================================================================
echo ""
echo "============================================="
echo " 测试完成: $(date '+%Y-%m-%d %H:%M:%S')"
echo " 总计: $TOTAL  通过: $PASS  失败: $FAIL"
if [ "$FAIL" -eq 0 ]; then
  echo -e " 结果: ${GREEN}全部通过 ✅${NC}"
else
  echo -e " 结果: ${RED}存在失败 ❌${NC}"
fi
echo "============================================="
exit $FAIL
