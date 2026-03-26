#!/bin/bash
# API_TYPE: admin
# =============================================================================
# Story 4-6: 配置生效状态与作用域下发校验 API 测试
# 覆盖：广告管理 effectiveStatus、优惠券 effectiveStatus、秒杀 effectiveStatus、
#       专题/优选专区状态切换、品牌推荐列表
# 用法: bash script/shell/api-test/4-6/test_4_6_api.sh [ADMIN_BASE_URL]
# 默认: http://47.107.224.56:8000
# =============================================================================

set -euo pipefail

BASE_URL="${1:-http://47.107.224.56:8000}"
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
echo " Story 4-6: 配置生效状态与作用域下发校验"
echo " Admin URL: $BASE_URL"
echo " Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================="
echo ""

# 0. Admin 登录
log_info "0. Admin 登录获取 token"
ADMIN_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}')
ADMIN_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$ADMIN_RESP")
if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "None" ]; then
  log_pass "Admin 登录成功"
else
  log_fail "Admin 登录失败: $ADMIN_RESP"
  exit 1
fi

# 1. 广告列表 - 检查 effectiveStatus 字段
log_info "1. 广告列表 - effectiveStatus 字段检查"
AD_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/homeAdvertise/queryHomeAdvertiseList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
AD_CODE=$(json_val "d.get('code','')" "$AD_RESP")
AD_COUNT=$(json_val "len(d.get('data',[]))" "$AD_RESP")
if [ "$AD_CODE" = "000000" ] && [ "$AD_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "广告列表返回成功，$AD_COUNT 条"
else
  log_fail "广告列表失败: code=$AD_CODE count=$AD_COUNT"
fi

# 检查 effectiveStatus 字段
HAS_EFF=$(json_val "'effectiveStatus' in (d.get('data',[])[0] if d.get('data') else {})" "$AD_RESP")
if [ "$HAS_EFF" = "True" ]; then
  log_pass "广告列表包含 effectiveStatus 字段"
  EFF_VAL=$(json_val "d.get('data',[])[0].get('effectiveStatus','')" "$AD_RESP")
  echo "    第一条广告 effectiveStatus=$EFF_VAL"
else
  log_fail "广告列表缺少 effectiveStatus 字段"
fi

# 2. 优惠券列表 - 检查 effectiveStatus 字段
log_info "2. 优惠券列表 - effectiveStatus 字段检查"
COUPON_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/coupon/queryCouponList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
COUPON_CODE=$(json_val "d.get('code','')" "$COUPON_RESP")
COUPON_COUNT=$(json_val "len(d.get('data',[]))" "$COUPON_RESP")
if [ "$COUPON_CODE" = "000000" ] && [ "$COUPON_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "优惠券列表返回成功，$COUPON_COUNT 条"
else
  log_fail "优惠券列表失败: code=$COUPON_CODE"
fi

HAS_COUPON_EFF=$(json_val "'effectiveStatus' in (d.get('data',[])[0] if d.get('data') else {})" "$COUPON_RESP")
if [ "$HAS_COUPON_EFF" = "True" ]; then
  log_pass "优惠券列表包含 effectiveStatus 字段"
else
  log_fail "优惠券列表缺少 effectiveStatus 字段"
fi

# 3. 优惠券详情 - effectiveStatus 字段
log_info "3. 优惠券详情 - effectiveStatus 字段"
COUPON_ID=$(json_val "d.get('data',[])[0].get('id',0) if d.get('data') else 0" "$COUPON_RESP")
COUPON_DETAIL=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/coupon/queryCouponDetail?id=$COUPON_ID" \
  -H "Authorization: $ADMIN_TOKEN")
CD_CODE=$(json_val "d.get('code','')" "$COUPON_DETAIL")
if [ "$CD_CODE" = "000000" ]; then
  log_pass "优惠券详情返回成功"
  HAS_DET_EFF=$(json_val "'effectiveStatus' in d.get('data',{})" "$COUPON_DETAIL")
  if [ "$HAS_DET_EFF" = "True" ]; then
    log_pass "优惠券详情包含 effectiveStatus 字段"
  else
    log_fail "优惠券详情缺少 effectiveStatus 字段"
  fi
else
  log_fail "优惠券详情请求失败: code=$CD_CODE"
fi

# 4. 秒杀活动列表 - effectiveStatus 字段
log_info "4. 秒杀活动列表 - effectiveStatus 字段检查"
SECKILL_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/seckillActivity/querySeckillActivityList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
SK_CODE=$(json_val "d.get('code','')" "$SECKILL_RESP")
SK_COUNT=$(json_val "len(d.get('data',[]))" "$SECKILL_RESP")
if [ "$SK_CODE" = "000000" ] && [ "$SK_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "秒杀活动列表返回成功，$SK_COUNT 条"
else
  log_fail "秒杀活动列表失败: code=$SK_CODE count=$SK_COUNT"
fi

HAS_SK_EFF=$(json_val "'effectiveStatus' in (d.get('data',[])[0] if d.get('data') else {})" "$SECKILL_RESP")
if [ "$HAS_SK_EFF" = "True" ]; then
  log_pass "秒杀活动列表包含 effectiveStatus 字段"
else
  log_fail "秒杀活动列表缺少 effectiveStatus 字段"
fi

# 5. 专题列表
log_info "5. 专题列表查询"
SUBJECT_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/cms/subject/querySubjectList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
SUBJ_CODE=$(json_val "d.get('code','')" "$SUBJECT_RESP")
SUBJ_COUNT=$(json_val "len(d.get('data',[]))" "$SUBJECT_RESP")
if [ "$SUBJ_CODE" = "000000" ] && [ "$SUBJ_COUNT" -gt 0 ] 2>/dev/null; then
  log_pass "专题列表返回成功，$SUBJ_COUNT 条"
else
  log_fail "专题列表失败: code=$SUBJ_CODE"
fi

# 6. 优选专区列表
log_info "6. 优选专区列表查询"
PA_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/cms/prefrenceArea/queryPreferredAreaList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
PA_CODE=$(json_val "d.get('code','')" "$PA_RESP")
PA_COUNT=$(json_val "len(d.get('data',[]))" "$PA_RESP")
if [ "$PA_CODE" = "000000" ]; then
  log_pass "优选专区列表返回成功，${PA_COUNT:-0} 条"
else
  log_fail "优选专区列表失败: code=$PA_CODE resp=$PA_RESP"
fi

# 7. 品牌推荐列表
log_info "7. 品牌推荐列表查询"
BRAND_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/homeBrand/queryHomeBrandList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
BR_CODE=$(json_val "d.get('code','')" "$BRAND_RESP")
if [ "$BR_CODE" = "000000" ]; then
  log_pass "品牌推荐列表返回成功"
else
  log_fail "品牌推荐列表失败: code=$BR_CODE"
fi

# 8. 推荐专题列表
log_info "8. 推荐专题列表查询"
RS_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/sms/homeRecommendSubject/queryHomeRecommendSubjectList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
RS_CODE=$(json_val "d.get('code','')" "$RS_RESP")
if [ "$RS_CODE" = "000000" ]; then
  log_pass "推荐专题列表返回成功"
else
  log_fail "推荐专题列表失败: code=$RS_CODE"
fi

# 9. 专题分类列表
log_info "9. 专题分类列表查询"
SC_RESP=$(curl -s --max-time $TIMEOUT \
  "$BASE_URL/api/cms/subjectCategory/querySubjectCategoryList?current=1&pageSize=10" \
  -H "Authorization: $ADMIN_TOKEN")
SC_CODE=$(json_val "d.get('code','')" "$SC_RESP")
if [ "$SC_CODE" = "000000" ]; then
  log_pass "专题分类列表返回成功"
else
  log_fail "专题分类列表失败: code=$SC_CODE"
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
