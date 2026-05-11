#!/bin/bash
# Story 10.11 - 节点健康检查（在 47.107.224.56 上跑）。
# 仅做只读查询，不会改变节点状态。

set -uo pipefail

FISCO_HOME="${FISCO_HOME:-/opt/fisco}"
NODE_HOST="${NODE_HOST:-127.0.0.1}"
NODE_COUNT="${NODE_COUNT:-4}"
RPC_PORT_BASE="${RPC_PORT_BASE:-20200}"

echo "=== fisco-bcos processes ==="
pgrep -fa fisco-bcos | grep -v 'check_nodes' || echo "(no fisco-bcos process)"

echo
echo "=== RPC port LISTEN status ==="
for ((i=0; i<NODE_COUNT; i++)); do
    p=$((RPC_PORT_BASE + i))
    if ss -ltn 2>/dev/null | grep -q ":${p} "; then
        echo "  RPC ${p}: LISTEN"
    else
        echo "  RPC ${p}: NOT LISTEN"
    fi
done

echo
echo "=== Latest log line per node (last 1 line) ==="
for ((i=0; i<NODE_COUNT; i++)); do
    NODE_DIR="${FISCO_HOME}/nodes/${NODE_HOST}/node${i}"
    LOG_FILE=$(ls -1t "${NODE_DIR}/log/log_"*.log 2>/dev/null | head -1 || true)
    if [ -n "${LOG_FILE}" ]; then
        echo "  node${i}: $(tail -n 1 "${LOG_FILE}" | sed 's/[\t]//g' | cut -c1-160)"
    else
        echo "  node${i}: no log file under ${NODE_DIR}/log/"
    fi
done

echo
echo "=== Block height query (mTLS via SDK certs) ==="
SDK_DIR="${FISCO_HOME}/sdk"
if [ -f "${SDK_DIR}/sdk.key" ] && [ -f "${SDK_DIR}/sdk.crt" ] && [ -f "${SDK_DIR}/ca.crt" ]; then
    # FISCO BCOS 3.x RPC 支持 JSON-RPC over WebSocket SSL；HTTP JSON-RPC 走 https + mTLS。
    PAYLOAD='{"jsonrpc":"2.0","method":"getBlockNumber","params":["group0",""],"id":1}'
    for ((i=0; i<NODE_COUNT; i++)); do
        p=$((RPC_PORT_BASE + i))
        RESP=$(curl -k -s --max-time 5 \
            --cacert "${SDK_DIR}/ca.crt" \
            --cert "${SDK_DIR}/sdk.crt" \
            --key "${SDK_DIR}/sdk.key" \
            -X POST "https://${NODE_HOST}:${p}" \
            -H "Content-Type: application/json" \
            -d "${PAYLOAD}" 2>&1 || true)
        echo "  node${i} (https://${NODE_HOST}:${p}): ${RESP:0:200}"
    done
else
    echo "  SDK certs not found in ${SDK_DIR}, skipping JSON-RPC check"
fi
