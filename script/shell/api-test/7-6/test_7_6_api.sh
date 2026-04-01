#!/bin/bash
# API_TYPE: admin
# =============================================================================
# Story 7.6: 授权链路重试回放与升级 API 测试
# 覆盖：链路监控列表、干预动作查询、重试、回放、暂停、升级
# 用法: bash script/shell/api-test/7-6/test_7_6_api.sh [ADMIN_BASE_URL]
# 默认: http://47.107.224.56:8888
# =============================================================================

set -euo pipefail

BASE_URL="${1:-http://47.107.224.56:8888}"
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
echo " Story 7.6: 授权链路重试回放与升级"
echo " Admin URL: $BASE_URL"
echo " Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "============================================="
echo ""

# 0. Admin 登录
log_info "0. Admin 登录"
LOGIN_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/sys/user/login" \
  -H 'Content-Type: application/json' \
  -d '{"account":"admin","password":"123456"}')
ADMIN_TOKEN=$(json_val "d.get('data',{}).get('token','')" "$LOGIN_RESP")
ADMIN_CODE=$(json_val "d.get('code','')" "$LOGIN_RESP")
if [ -n "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "None" ] && [ "$ADMIN_CODE" = "000000" ]; then
  log_pass "Admin 登录成功"
else
  log_fail "Admin 登录失败: $LOGIN_RESP"
  exit 1
fi
AUTH="Authorization: Bearer $ADMIN_TOKEN"

# 1. 查询链路监控列表 - queryChainMonitorList
log_info "1. 查询链路监控列表 queryChainMonitorList"
CHAIN_RESP=$(curl -s --max-time $TIMEOUT "$BASE_URL/api/oms/order/queryChainMonitorList?current=1&pageSize=10" \
  -H "$AUTH")
CHAIN_CODE=$(json_val "d.get('code','')" "$CHAIN_RESP")
if [ "$CHAIN_CODE" = "000000" ] || [ "$CHAIN_CODE" = "0" ]; then
  log_pass "链路监控列表查询成功"
  CHAIN_TOTAL=$(json_val "d.get('data',{}).get('total', '0')" "$CHAIN_RESP")
  echo "    总记录数: $CHAIN_TOTAL"
else
  CHAIN_MSG=$(json_val "d.get('message',d.get('msg',''))" "$CHAIN_RESP")
  log_fail "链路监控列表查询失败: code=$CHAIN_CODE msg=$CHAIN_MSG"
fi

# 2. 获取一个有效的 orderId 用于后续测试
ORDER_ID=$(json_val "d.get('data',{}).get('list',[{}])[0].get('orderId', 0)" "$CHAIN_RESP")
log_info "2. 准备测试: 使用 orderId=$ORDER_ID"

if [ "$ORDER_ID" = "0" ] || [ -z "$ORDER_ID" ]; then
  log_info "无链路数据可测试，跳过干预操作测试"
  log_info "仅验证 API 路由存在性"

  # 验证干预 API 路由存在（通过检查 400/401/403 而非 404）
  log_info "3. 验证干预 API 路由存在"

  # 测试 retryChain 路由
  RETRY_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/retryChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"orderId":999999,"remark":"test"}')
  RETRY_CODE=$(json_val "d.get('code','')" "$RETRY_RESP")
  if [ "$RETRY_CODE" != "" ] && [ "$RETRY_CODE" != "404" ]; then
    log_pass "retryChain 路由存在 (code=$RETRY_CODE，非 404)"
  else
    log_fail "retryChain 路由返回 404，路由未注册"
  fi

  # 测试 replayChain 路由
  REPLAY_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/replayChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"orderId":999999,"replayReason":"test"}')
  REPLAY_CODE=$(json_val "d.get('code','')" "$REPLAY_RESP")
  if [ "$REPLAY_CODE" != "" ] && [ "$REPLAY_CODE" != "404" ]; then
    log_pass "replayChain 路由存在 (code=$REPLAY_CODE，非 404)"
  else
    log_fail "replayChain 路由返回 404，路由未注册"
  fi

  # 测试 pauseChain 路由
  PAUSE_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/pauseChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"orderId":999999,"pauseReason":"test"}')
  PAUSE_CODE=$(json_val "d.get('code','')" "$PAUSE_RESP")
  if [ "$PAUSE_CODE" != "" ] && [ "$PAUSE_CODE" != "404" ]; then
    log_pass "pauseChain 路由存在 (code=$PAUSE_CODE，非 404)"
  else
    log_fail "pauseChain 路由返回 404，路由未注册"
  fi

  # 测试 escalateChain 路由
  ESC_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/escalateChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"orderId":999999,"escalateReason":"test"}')
  ESC_CODE=$(json_val "d.get('code','')" "$ESC_RESP")
  if [ "$ESC_CODE" != "" ] && [ "$ESC_CODE" != "404" ]; then
    log_pass "escalateChain 路由存在 (code=$ESC_CODE，非 404)"
  else
    log_fail "escalateChain 路由返回 404，路由未注册"
  fi

  # 测试 queryChainActions 路由
  ACTIONS_RESP=$(curl -s --max-time $TIMEOUT "$BASE_URL/api/oms/order/queryChainActions?orderId=999999" \
    -H "$AUTH")
  ACTIONS_CODE=$(json_val "d.get('code','')" "$ACTIONS_RESP")
  if [ "$ACTIONS_CODE" != "" ] && [ "$ACTIONS_CODE" != "404" ]; then
    log_pass "queryChainActions 路由存在 (code=$ACTIONS_CODE，非 404)"
  else
    log_fail "queryChainActions 路由返回 404，路由未注册"
  fi

else
  echo "    存在有效订单，开始测试干预操作"

  # 3. 查询可用干预动作 - queryChainActions
  log_info "3. 查询可用干预动作 queryChainActions"
  ACTIONS_RESP=$(curl -s --max-time $TIMEOUT "$BASE_URL/api/oms/order/queryChainActions?orderId=$ORDER_ID" \
    -H "$AUTH")
  ACTIONS_CODE=$(json_val "d.get('code','')" "$ACTIONS_RESP")
  if [ "$ACTIONS_CODE" = "000000" ] || [ "$ACTIONS_CODE" = "0" ]; then
    log_pass "查询可用干预动作成功"
    ACTIONS_LIST=$(json_val "d.get('data',{}).get('availableActions',[])" "$ACTIONS_RESP")
    echo "    可用动作: $ACTIONS_LIST"
  else
    ACTIONS_MSG=$(json_val "d.get('message',d.get('msg',''))" "$ACTIONS_RESP")
    log_fail "查询可用干预动作失败: code=$ACTIONS_CODE msg=$ACTIONS_MSG"
  fi

  # 4. 暂停链路 - pauseChain (测试用，暂停已暂停的链路应返回错误)
  log_info "4. 暂停链路 pauseChain (测试错误场景)"
  PAUSE_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/pauseChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"orderId\":$ORDER_ID,\"pauseReason\":\"QA测试暂停\"}")
  PAUSE_CODE=$(json_val "d.get('code','')" "$PAUSE_RESP")
  # 可能返回 400(已暂停) 或 0(成功)
  if [ "$PAUSE_CODE" = "000000" ] || [ "$PAUSE_CODE" = "0" ] || [ "$PAUSE_CODE" = "400" ]; then
    log_pass "暂停链路操作响应正确 (code=$PAUSE_CODE)"
  else
    PAUSE_MSG=$(json_val "d.get('message',d.get('msg',''))" "$PAUSE_RESP")
    log_fail "暂停链路操作失败: code=$PAUSE_CODE msg=$PAUSE_MSG"
  fi

  # 5. 升级链路 - escalateChain (测试错误场景)
  log_info "5. 升级链路 escalateChain (测试错误场景)"
  ESC_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/escalateChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"orderId\":$ORDER_ID,\"escalateReason\":\"QA测试升级\"}")
  ESC_CODE=$(json_val "d.get('code','')" "$ESC_RESP")
  # 可能返回 400(不满足条件) 或 0(成功)
  if [ "$ESC_CODE" = "000000" ] || [ "$ESC_CODE" = "0" ] || [ "$ESC_CODE" = "400" ]; then
    log_pass "升级链路操作响应正确 (code=$ESC_CODE)"
  else
    ESC_MSG=$(json_val "d.get('message',d.get('msg',''))" "$ESC_RESP")
    log_fail "升级链路操作失败: code=$ESC_CODE msg=$ESC_MSG"
  fi

  # 6. 重试链路 - retryChain (测试错误场景)
  log_info "6. 重试链路 retryChain (测试错误场景)"
  RETRY_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/retryChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"orderId\":$ORDER_ID,\"remark\":\"QA测试重试\"}")
  RETRY_CODE=$(json_val "d.get('code','')" "$RETRY_RESP")
  # 可能返回 400(不满足条件) 或 0(成功)
  if [ "$RETRY_CODE" = "000000" ] || [ "$RETRY_CODE" = "0" ] || [ "$RETRY_CODE" = "400" ]; then
    log_pass "重试链路操作响应正确 (code=$RETRY_CODE)"
  else
    RETRY_MSG=$(json_val "d.get('message',d.get('msg',''))" "$RETRY_RESP")
    log_fail "重试链路操作失败: code=$RETRY_CODE msg=$RETRY_MSG"
  fi

  # 7. 回放链路 - replayChain (测试错误场景)
  log_info "7. 回放链路 replayChain (测试错误场景)"
  REPLAY_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/replayChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d "{\"orderId\":$ORDER_ID,\"replayReason\":\"QA测试回放\"}")
  REPLAY_CODE=$(json_val "d.get('code','')" "$REPLAY_RESP")
  # 可能返回 400(不满足条件) 或 0(成功)
  if [ "$REPLAY_CODE" = "000000" ] || [ "$REPLAY_CODE" = "0" ] || [ "$REPLAY_CODE" = "400" ]; then
    log_pass "回放链路操作响应正确 (code=$REPLAY_CODE)"
  else
    REPLAY_MSG=$(json_val "d.get('message',d.get('msg',''))" "$REPLAY_RESP")
    log_fail "回放链路操作失败: code=$REPLAY_CODE msg=$REPLAY_MSG"
  fi

  # 8. 跨主体权限校验 - 用错误的主体范围
  log_info "8. 跨主体权限校验 (应返回 403)"
  WRONG_SCOPE_RESP=$(curl -s --max-time $TIMEOUT -X POST "$BASE_URL/api/oms/order/pauseChain" \
    -H "$AUTH" -H 'Content-Type: application/json' \
    -d '{"orderId":1,"pauseReason":"跨主体测试","platformId":999,"tenantId":999,"merchantId":999}')
  WRONG_CODE=$(json_val "d.get('code','')" "$WRONG_SCOPE_RESP")
  if [ "$WRONG_CODE" = "403" ]; then
    log_pass "跨主体权限校验正确拒绝 (code=403)"
  else
    WRONG_MSG=$(json_val "d.get('message',d.get('msg',''))" "$WRONG_SCOPE_RESP")
    log_fail "跨主体权限校验异常: code=$WRONG_CODE msg=$WRONG_MSG"
  fi

  # 9. 干预动作后的链路状态验证
  log_info "9. 验证干预后链路状态更新"
  CHAIN_AFTER_RESP=$(curl -s --max-time $TIMEOUT "$BASE_URL/api/oms/order/queryChainMonitorList?current=1&pageSize=10" \
    -H "$AUTH")
  CHAIN_AFTER_CODE=$(json_val "d.get('code','')" "$CHAIN_AFTER_RESP")
  if [ "$CHAIN_AFTER_CODE" = "000000" ] || [ "$CHAIN_AFTER_CODE" = "0" ]; then
    log_pass "干预后链路列表查询正常"
  else
    log_fail "干预后链路列表查询失败"
  fi
fi

# 10. 未授权访问测试
log_info "10. 未授权访问干预 API (应返回 401)"
NOAUTH_RESP=$(curl -s --max-time $TIMEOUT -o /dev/null -w "%{http_code}" \
  -X POST "$BASE_URL/api/oms/order/pauseChain" \
  -H 'Content-Type: application/json' \
  -d '{"orderId":1,"pauseReason":"test"}')
if [ "$NOAUTH_RESP" = "401" ]; then
  log_pass "未授权访问正确返回 401"
else
  log_fail "未授权访问应返回 401，实际: $NOAUTH_RESP"
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