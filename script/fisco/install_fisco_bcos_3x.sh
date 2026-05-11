#!/bin/bash
# Story 10.11 / Task 1 - Phase 1
# Install FISCO BCOS 3.x air mode (single group, 4 nodes) on 47.107.224.56.
#
# Idempotency:
#   - 多次执行不重复安装：检测到 ${FISCO_HOME}/nodes 已存在则跳过 build_chain.sh，
#     仅做 start_all.sh + 验证。
#   - 如需完全重装，请先手动 rm -rf ${FISCO_HOME}/nodes 并 stop 节点。
#
# 用法（远程服务器）：
#   sudo bash /tmp/install_fisco_bcos_3x.sh
#
# 输出：
#   /opt/fisco/                                   FISCO 根目录
#   /opt/fisco/fisco-bcos                         节点二进制
#   /opt/fisco/build_chain.sh                     部署脚本
#   /opt/fisco/nodes/127.0.0.1/node{0,1,2,3}/     四节点目录
#   /opt/fisco/nodes/127.0.0.1/sdk/{ca,sdk}.crt   SDK 证书 (原始位置)
#   /opt/fisco/sdk/{ca.crt,sdk.crt,sdk.key}       SDK 证书 (标准化位置, 0400)
#   /opt/fisco/INSTALL.log                        本次安装日志

set -euo pipefail

# ---- 可调参数 ----
FISCO_VERSION="${FISCO_VERSION:-v3.16.4}"
FISCO_HOME="${FISCO_HOME:-/opt/fisco}"
NODE_HOST="${NODE_HOST:-127.0.0.1}"
NODE_COUNT="${NODE_COUNT:-4}"
P2P_PORT_BASE="${P2P_PORT_BASE:-30300}"
RPC_PORT_BASE="${RPC_PORT_BASE:-20200}"
DOWNLOAD_BASE="${DOWNLOAD_BASE:-https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download}"
BUILD_CHAIN_URL="${BUILD_CHAIN_URL:-https://raw.githubusercontent.com/FISCO-BCOS/FISCO-BCOS/${FISCO_VERSION}/tools/BcosAirBuilder/build_chain.sh}"

LOG_FILE="${FISCO_HOME}/INSTALL.log"

log() { printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "${LOG_FILE}"; }

# ---- 0. 前置检查 ----
if [ "$(id -u)" -ne 0 ]; then
    echo "错误：请以 root 运行（FISCO 需要在 /opt/fisco 安装并使用低端口）"
    exit 1
fi

mkdir -p "${FISCO_HOME}"
: > "${LOG_FILE}"
log "FISCO BCOS install starting: version=${FISCO_VERSION} home=${FISCO_HOME} nodes=${NODE_HOST}:${NODE_COUNT}"

# ---- 1. 系统依赖 ----
log "Step 1: install OS prerequisites (openssl/curl/wget/bc/python3/net-tools/jq)"
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq openssl curl wget bc python3 net-tools jq lsof >/dev/null

# ---- 2. 端口冲突检查 ----
log "Step 2: check port conflicts"
for ((i=0; i<NODE_COUNT; i++)); do
    p2p=$((P2P_PORT_BASE + i))
    rpc=$((RPC_PORT_BASE + i))
    for port in $p2p $rpc; do
        if lsof -i :"${port}" -sTCP:LISTEN -t >/dev/null 2>&1; then
            log "ERROR: port ${port} already in use"
            lsof -i :"${port}" -sTCP:LISTEN | tee -a "${LOG_FILE}"
            exit 2
        fi
    done
done
log "  all required ports free: P2P ${P2P_PORT_BASE}-$((P2P_PORT_BASE+NODE_COUNT-1)) RPC ${RPC_PORT_BASE}-$((RPC_PORT_BASE+NODE_COUNT-1))"

# ---- 3. 下载 fisco-bcos 二进制 ----
BINARY_TGZ="${FISCO_HOME}/fisco-bcos-linux-x86_64.tar.gz"
BINARY_PATH="${FISCO_HOME}/fisco-bcos"
if [ ! -x "${BINARY_PATH}" ]; then
    log "Step 3: download fisco-bcos ${FISCO_VERSION} binary"
    curl -fsSL --max-time 300 -o "${BINARY_TGZ}" \
        "${DOWNLOAD_BASE}/${FISCO_VERSION}/fisco-bcos-linux-x86_64.tar.gz"
    tar -xzf "${BINARY_TGZ}" -C "${FISCO_HOME}"
    chmod +x "${BINARY_PATH}"
    "${BINARY_PATH}" --version | tee -a "${LOG_FILE}"
else
    log "Step 3: fisco-bcos binary already present, skipping download"
fi

# ---- 4. 下载 build_chain.sh ----
BUILD_CHAIN="${FISCO_HOME}/build_chain.sh"
if [ ! -x "${BUILD_CHAIN}" ]; then
    log "Step 4: download build_chain.sh from ${BUILD_CHAIN_URL}"
    curl -fsSL --max-time 60 -o "${BUILD_CHAIN}" "${BUILD_CHAIN_URL}"
    chmod +x "${BUILD_CHAIN}"
else
    log "Step 4: build_chain.sh already present, skipping download"
fi

# ---- 5. 生成节点 ----
NODES_DIR="${FISCO_HOME}/nodes"
if [ ! -d "${NODES_DIR}/${NODE_HOST}" ]; then
    log "Step 5: generate ${NODE_COUNT} nodes via build_chain.sh"
    cd "${FISCO_HOME}"
    bash "${BUILD_CHAIN}" \
        -p "${P2P_PORT_BASE},${RPC_PORT_BASE}" \
        -l "${NODE_HOST}:${NODE_COUNT}" \
        -o "${NODES_DIR}" \
        -e "${BINARY_PATH}" 2>&1 | tee -a "${LOG_FILE}"
else
    log "Step 5: nodes dir ${NODES_DIR}/${NODE_HOST} already exists, skipping generation"
fi

# ---- 6. 启动节点 ----
START_ALL="${NODES_DIR}/${NODE_HOST}/start_all.sh"
if [ -x "${START_ALL}" ]; then
    log "Step 6: start all 4 nodes"
    bash "${START_ALL}" 2>&1 | tee -a "${LOG_FILE}"
    sleep 5
else
    log "ERROR: start_all.sh not found at ${START_ALL}"
    exit 3
fi

# ---- 7. 进程 / 端口 / 区块高度 验证 ----
log "Step 7: verify nodes are up"
sleep 3
PROCS=$(pgrep -fa fisco-bcos | grep -v 'install_fisco' || true)
if [ -z "${PROCS}" ]; then
    log "ERROR: no fisco-bcos process running after start_all.sh"
    exit 4
fi
echo "${PROCS}" | tee -a "${LOG_FILE}"

log "Step 7.1: confirm RPC ports listening"
for ((i=0; i<NODE_COUNT; i++)); do
    p=$((RPC_PORT_BASE + i))
    if ss -ltn 2>/dev/null | grep -q ":${p} "; then
        log "  RPC port ${p}: LISTEN"
    else
        log "  WARN: RPC port ${p} not listening yet (may still be starting)"
    fi
done

log "Step 7.2: tail last 5 lines of node0 log for sanity"
NODE0_LOG=$(ls -1t "${NODES_DIR}/${NODE_HOST}/node0/log/log_"*.log 2>/dev/null | head -1 || true)
if [ -n "${NODE0_LOG}" ]; then
    log "  log file: ${NODE0_LOG}"
    tail -n 5 "${NODE0_LOG}" | sed 's/^/    /' | tee -a "${LOG_FILE}"
fi

# ---- 8. 标准化 SDK 证书位置 ----
SDK_SRC="${NODES_DIR}/${NODE_HOST}/sdk"
SDK_DST="${FISCO_HOME}/sdk"
if [ -d "${SDK_SRC}" ]; then
    log "Step 8: standardize SDK certs to ${SDK_DST}"
    mkdir -p "${SDK_DST}"
    cp -f "${SDK_SRC}/ca.crt" "${SDK_DST}/ca.crt"
    cp -f "${SDK_SRC}/sdk.crt" "${SDK_DST}/sdk.crt"
    cp -f "${SDK_SRC}/sdk.key" "${SDK_DST}/sdk.key"
    chmod 0444 "${SDK_DST}/ca.crt" "${SDK_DST}/sdk.crt"
    chmod 0400 "${SDK_DST}/sdk.key"
    log "  SDK certs:"
    ls -la "${SDK_DST}" | sed 's/^/    /' | tee -a "${LOG_FILE}"
else
    log "WARN: SDK dir not found at ${SDK_SRC}, skipping cert standardization"
fi

# ---- 9. 总结 ----
log "Step 9: install complete"
log "Summary:"
log "  fisco binary    : ${BINARY_PATH}"
log "  nodes dir       : ${NODES_DIR}/${NODE_HOST}"
log "  start script    : ${START_ALL}"
log "  stop script     : ${NODES_DIR}/${NODE_HOST}/stop_all.sh"
log "  SDK certs       : ${SDK_DST}/{ca.crt,sdk.crt,sdk.key}"
log "  RPC endpoints   : ${NODE_HOST}:${RPC_PORT_BASE}-$((RPC_PORT_BASE+NODE_COUNT-1)) (TLS, mTLS via SDK certs)"
log "  group           : group0"
log "  chain           : chain0"
log ""
log "下一步（Task 2）：编写 contracts/CardToken.sol 并通过 Java console2 或 Go SDK 部署到 group0"
