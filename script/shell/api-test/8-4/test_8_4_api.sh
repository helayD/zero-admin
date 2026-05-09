#!/bin/bash
# API_TYPE: admin
# =============================================================================
# Story 8-4: 经营漏斗与优惠券核销看板 API 测试
# 覆盖：admin 经营漏斗接口、front 首页广告点击埋点、活动筛选与字段透传
# 用法: bash script/shell/api-test/8-4/test_8_4_api.sh [ADMIN_BASE_URL] [FRONT_BASE_URL]
# 默认: http://47.107.224.56:8000 http://47.107.224.56:9999
# =============================================================================

set -euo pipefail

ADMIN_BASE_URL="${1:-http://47.107.224.56:8000}"
FRONT_BASE_URL="${2:-http://47.107.224.56:9999}"
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
import sys, json
try:
    d = json.load(sys.stdin)
    print($1)
except Exception:
    print('')
" 2>/dev/null <<< "$2" || echo ''
}

NOW_START=$(TZ=Asia/Shanghai date -v-1H '+%Y-%m-%d %H:00:00')
NOW_END=$(TZ=Asia/Shanghai date -v+1H '+%Y-%m-%d %H:00:00')
NOW_START_ESC="${NOW_START// /%20}"
NOW_END_ESC="${NOW_END// /%20}"
AD_START=$(TZ=Asia/Shanghai date -v-1d '+%Y-%m-%d %H:%M:%S')
AD_END=$(TZ=Asia/Shanghai date -v+1d '+%Y-%m-%d %H:%M:%S')

TMP_NAME="8-4-acceptance-$(date +%s)"
TMP_AD_ID=""

cleanup() {
  if [ -n "${TMP_AD_ID:-}" ] && [ -n "${ADMIN_TOKEN:-}" ]; then
    curl -s --max-time "$TIMEOUT" \
      "$ADMIN_BASE_URL/api/sms/homeAdvertise/deleteHomeAdvertise?ids=$TMP_AD_ID" \
      -H "Authorization: $ADMIN_TOKEN" >/dev/null || true
  fi
}
trap cleanup EXIT

echo "============================================="
echo " Story 8-4: 经营漏斗与优惠券核销看板"
echo " Admin URL: $ADMIN_BASE_URL"
echo " Front URL: $FRONT_BASE_URL"
echo " Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================="
echo ""

log_info "0. Admin 登录"
ADMIN_RESP=$(curl -s --max-time "$TIMEOUT" -X POST "$ADMIN_BASE_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}')
ADMIN_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$ADMIN_RESP")
if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "None" ]; then
  log_pass "Admin 登录成功"
else
  log_fail "Admin 登录失败: $ADMIN_RESP"
  exit 1
fi

log_info "1. Front 会员登录"
FRONT_LOGIN=$(curl -s --max-time "$TIMEOUT" -X POST "$FRONT_BASE_URL/api/member/login" \
  -H 'Content-Type: application/json' \
  -d '{"mobile":"13800138001","password":"123456"}')
FRONT_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$FRONT_LOGIN")
if [ -n "$FRONT_TOKEN" ] && [ "$FRONT_TOKEN" != "None" ]; then
  log_pass "Front 会员登录成功"
else
  log_fail "Front 登录失败: $FRONT_LOGIN"
  exit 1
fi

log_info "2. 创建临时首页广告"
ADD_AD_RESP=$(curl -s --max-time "$TIMEOUT" -X POST "$ADMIN_BASE_URL/api/sms/homeAdvertise/addHomeAdvertise" \
  -H "Authorization: $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$TMP_NAME\",\"type\":1,\"pic\":\"http://example.com/8-4.png\",\"startTime\":\"$AD_START\",\"endTime\":\"$AD_END\",\"status\":1,\"url\":\"https://example.com/8-4\",\"remark\":\"8-4 acceptance\",\"sort\":1}")
ADD_AD_CODE=$(json_val "d.get('code','')" "$ADD_AD_RESP")
if [ "$ADD_AD_CODE" = "000000" ]; then
  log_pass "临时首页广告创建成功"
else
  log_fail "创建临时首页广告失败: $ADD_AD_RESP"
  exit 1
fi

AD_QUERY_RESP=$(curl -s --max-time "$TIMEOUT" \
  "$ADMIN_BASE_URL/api/sms/homeAdvertise/queryHomeAdvertiseList?current=1&pageSize=20&name=$TMP_NAME" \
  -H "Authorization: $ADMIN_TOKEN")
TMP_AD_ID=$(json_val "(d.get('data',[]) or [{}])[0].get('id',0)" "$AD_QUERY_RESP")
if [ -n "$TMP_AD_ID" ] && [ "$TMP_AD_ID" != "0" ] && [ "$TMP_AD_ID" != "None" ]; then
  log_pass "临时首页广告查询成功，id=$TMP_AD_ID"
else
  log_fail "查询临时首页广告失败: $AD_QUERY_RESP"
  exit 1
fi

log_info "3. 首页接口返回广告活动字段"
HOME_RESP=$(curl -s --max-time "$TIMEOUT" "$FRONT_BASE_URL/api/home/index")
MATCHED_TYPE=$(json_val "next((item.get('activityType','') for item in (d.get('data',{}).get('advertiseList',[]) or []) if item.get('id')==$TMP_AD_ID), '')" "$HOME_RESP")
MATCHED_ACTIVITY_ID=$(json_val "next((item.get('activityId',0) for item in (d.get('data',{}).get('advertiseList',[]) or []) if item.get('id')==$TMP_AD_ID), 0)" "$HOME_RESP")
if [ "$MATCHED_TYPE" = "home_advertise" ] && [ "$MATCHED_ACTIVITY_ID" = "$TMP_AD_ID" ]; then
  log_pass "首页广告返回 activityType/activityId"
else
  log_fail "首页广告缺少活动字段或值错误: type=$MATCHED_TYPE activityId=$MATCHED_ACTIVITY_ID"
fi

log_info "4. 记录首页广告点击"
CLICK_RESP=$(curl -s --max-time "$TIMEOUT" -X POST "$FRONT_BASE_URL/api/home/recordHomeAdvertiseClick" \
  -H "Authorization: Bearer $FRONT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"advertiseId\":$TMP_AD_ID,\"traceId\":\"8-4-$TMP_AD_ID-$(date +%s)\"}")
CLICK_CODE=$(json_val "d.get('code','')" "$CLICK_RESP")
if [ "$CLICK_CODE" = "0" ]; then
  log_pass "首页广告点击埋点成功"
else
  log_fail "首页广告点击埋点失败: $CLICK_RESP"
fi

log_info "5. 查询经营漏斗看板"
DASHBOARD_RESP=$(curl -s --max-time "$TIMEOUT" \
  "$ADMIN_BASE_URL/api/sms/operateDashboard/queryOperateFunnelDashboard?startTime=$NOW_START_ESC&endTime=$NOW_END_ESC&bucket=hour&activityType=home_advertise&activityId=$TMP_AD_ID" \
  -H "Authorization: $ADMIN_TOKEN")
DASHBOARD_CODE=$(json_val "d.get('code','')" "$DASHBOARD_RESP")
DASHBOARD_CLICK=$(json_val "d.get('data',{}).get('overview',{}).get('click',0)" "$DASHBOARD_RESP")
DASHBOARD_BUCKET=$(json_val "d.get('data',{}).get('bucket','')" "$DASHBOARD_RESP")
if [ "$DASHBOARD_CODE" = "000000" ]; then
  log_pass "经营漏斗看板接口返回成功"
else
  log_fail "经营漏斗看板接口失败: $DASHBOARD_RESP"
fi
if [ "$DASHBOARD_BUCKET" = "hour" ]; then
  log_pass "经营漏斗看板返回实际 bucket=hour"
else
  log_fail "经营漏斗看板 bucket 不符合预期: $DASHBOARD_BUCKET"
fi
if [ "$DASHBOARD_CLICK" -ge 1 ] 2>/dev/null; then
  log_pass "点击埋点已体现在经营漏斗看板中，click=$DASHBOARD_CLICK"
else
  log_fail "经营漏斗看板未观察到点击数据: $DASHBOARD_RESP"
fi

HAS_SERIES=$(json_val "len(d.get('data',{}).get('series',[]) or [])" "$DASHBOARD_RESP")
HAS_ACTIVITY_OPTIONS=$(json_val "'activityOptions' in d.get('data',{})" "$DASHBOARD_RESP")
if [ "$HAS_SERIES" -gt 0 ] 2>/dev/null; then
  log_pass "经营漏斗看板返回时间序列"
else
  log_fail "经营漏斗看板缺少时间序列"
fi
if [ "$HAS_ACTIVITY_OPTIONS" = "True" ]; then
  log_pass "经营漏斗看板返回活动选项"
else
  log_fail "经营漏斗看板缺少活动选项"
fi

log_info "6. 优惠券活动筛选接口可用性检查"
OPTIONS_RESP=$(curl -s --max-time "$TIMEOUT" \
  "$ADMIN_BASE_URL/api/sms/operateDashboard/queryOperateFunnelDashboard?startTime=$NOW_START_ESC&endTime=$NOW_END_ESC&bucket=hour" \
  -H "Authorization: $ADMIN_TOKEN")
COUPON_ACTIVITY_ID=$(json_val "next((item.get('activityId',0) for item in (d.get('data',{}).get('activityOptions',[]) or []) if item.get('activityType')=='coupon' and item.get('activityId',0) > 0), 0)" "$OPTIONS_RESP")
if [ "$COUPON_ACTIVITY_ID" -gt 0 ] 2>/dev/null; then
  COUPON_RESP=$(curl -s --max-time "$TIMEOUT" \
    "$ADMIN_BASE_URL/api/sms/operateDashboard/queryOperateFunnelDashboard?startTime=$NOW_START_ESC&endTime=$NOW_END_ESC&bucket=hour&activityType=coupon&activityId=$COUPON_ACTIVITY_ID" \
    -H "Authorization: $ADMIN_TOKEN")
  COUPON_CODE=$(json_val "d.get('code','')" "$COUPON_RESP")
  if [ "$COUPON_CODE" = "000000" ]; then
    log_pass "优惠券活动筛选请求成功，activityId=$COUPON_ACTIVITY_ID"
  else
    log_fail "优惠券活动筛选请求失败: $COUPON_RESP"
  fi
else
  log_pass "当前环境没有可用优惠券活动，跳过 coupon 筛选检查"
fi

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
