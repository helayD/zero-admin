#!/bin/bash
# Story 10.11 / Task 2.3 - 在 47.107.224.56 安装 console2 v3.7.0 并部署 CardToken 合约。
#
# 前置：
#   1. /tmp/console.tar.gz 已上传（144MB）
#   2. /tmp/CardToken.sol 已上传
#   3. /opt/fisco/sdk/{ca.crt,sdk.crt,sdk.key} 已就绪 (Task 1 产出)
#   4. 4 个节点已运行 (127.0.0.1:20200-20203)
#
# 执行后产出：
#   /opt/fisco/console/                              console2 安装目录
#   /opt/fisco/console/conf/config.toml              已配置好节点 + 证书路径
#   /opt/fisco/contracts/CardToken.sol               合约源码副本
#   /opt/fisco/contracts/CardToken.abi               合约 ABI（供 Go SDK 加载）
#   /opt/fisco/contracts/CONTRACT_ADDR               部署后合约地址
#   /opt/fisco/contracts/DEPLOY_TX                   部署交易 hash

set -euo pipefail

FISCO_HOME="${FISCO_HOME:-/opt/fisco}"
NODE_HOST="${NODE_HOST:-127.0.0.1}"
RPC_PORT_BASE="${RPC_PORT_BASE:-20200}"
CONSOLE_DIR="${FISCO_HOME}/console"
CONTRACTS_DIR="${FISCO_HOME}/contracts"

log() { printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"; }

if [ "$(id -u)" -ne 0 ]; then
    echo "请以 root 运行"; exit 1
fi

# ---- 1. 安装 Java 17 ----
if ! command -v java >/dev/null 2>&1; then
    log "Step 1: install OpenJDK 17 JRE headless"
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y -qq openjdk-17-jre-headless >/dev/null
fi
java -version 2>&1 | head -2

# ---- 2. 解压 console2 ----
if [ ! -d "${CONSOLE_DIR}" ]; then
    log "Step 2: extract console.tar.gz to ${CONSOLE_DIR}"
    [ -f /tmp/console.tar.gz ] || { echo "ERROR: /tmp/console.tar.gz missing"; exit 2; }
    mkdir -p "${FISCO_HOME}"
    tar -xzf /tmp/console.tar.gz -C "${FISCO_HOME}"
    # tarball 解出来叫 console/
    [ -d "${CONSOLE_DIR}" ] || { echo "ERROR: extracted dir not at ${CONSOLE_DIR}"; ls -la "${FISCO_HOME}"; exit 3; }
fi
ls "${CONSOLE_DIR}/console.sh" >/dev/null
log "  console.sh: ${CONSOLE_DIR}/console.sh"

# ---- 3. 配置 conf/config.toml ----
log "Step 3: configure conf/config.toml"
mkdir -p "${CONSOLE_DIR}/conf"

# 复制 SDK 证书
cp -f "${FISCO_HOME}/sdk/ca.crt" "${CONSOLE_DIR}/conf/ca.crt"
cp -f "${FISCO_HOME}/sdk/sdk.crt" "${CONSOLE_DIR}/conf/sdk.crt"
cp -f "${FISCO_HOME}/sdk/sdk.key" "${CONSOLE_DIR}/conf/sdk.key"

# 写 config.toml（air mode + 4 节点 RPC）
cat > "${CONSOLE_DIR}/conf/config.toml" <<TOML
[cryptoMaterial]
useSMCrypto = "false"
caCert = "conf/ca.crt"
sslCert = "conf/sdk.crt"
sslKey = "conf/sdk.key"

[network]
peers = ["${NODE_HOST}:${RPC_PORT_BASE}", "${NODE_HOST}:$((RPC_PORT_BASE+1))", "${NODE_HOST}:$((RPC_PORT_BASE+2))", "${NODE_HOST}:$((RPC_PORT_BASE+3))"]
defaultGroup = "group0"

[account]
keyStoreDir = "account"
accountFileFormat = "pem"

[log]
level = "info"
TOML

# ---- 4. 放入 CardToken.sol ----
log "Step 4: place CardToken.sol into console contracts dir"
[ -f /tmp/CardToken.sol ] || { echo "ERROR: /tmp/CardToken.sol missing"; exit 4; }
mkdir -p "${CONSOLE_DIR}/contracts/solidity"
cp -f /tmp/CardToken.sol "${CONSOLE_DIR}/contracts/solidity/CardToken.sol"

# ---- 5. 备份合约副本到 /opt/fisco/contracts ----
mkdir -p "${CONTRACTS_DIR}"
cp -f /tmp/CardToken.sol "${CONTRACTS_DIR}/CardToken.sol"
[ -f /tmp/CardToken.abi ] && cp -f /tmp/CardToken.abi "${CONTRACTS_DIR}/CardToken.abi"
[ -f /tmp/CardToken.bin ] && cp -f /tmp/CardToken.bin "${CONTRACTS_DIR}/CardToken.bin"

# ---- 6. 测试连通 ----
log "Step 6: console connectivity check (getBlockNumber)"
cd "${CONSOLE_DIR}"
echo "getBlockNumber" | bash console.sh 2>&1 | tee /tmp/console_check.log | grep -E "Block|Error|exception|info" | head -10 || true
if grep -q "Welcome to FISCO" /tmp/console_check.log; then
    log "  console banner ok"
fi

# ---- 7. 部署 CardToken ----
log "Step 7: deploy CardToken to group0"
DEPLOY_OUT="${CONTRACTS_DIR}/deploy.out"
echo "deploy CardToken" | bash console.sh > "${DEPLOY_OUT}" 2>&1 || true
cat "${DEPLOY_OUT}"

CONTRACT_ADDR=$(grep -oE "contract address: 0x[0-9a-fA-F]{40}" "${DEPLOY_OUT}" | head -1 | awk '{print $3}')
DEPLOY_TX=$(grep -oE "transaction hash: 0x[0-9a-fA-F]{64}" "${DEPLOY_OUT}" | head -1 | awk '{print $3}')

if [ -z "${CONTRACT_ADDR}" ]; then
    log "ERROR: CONTRACT_ADDR not found in deploy output, check ${DEPLOY_OUT}"
    exit 5
fi

echo "${CONTRACT_ADDR}" > "${CONTRACTS_DIR}/CONTRACT_ADDR"
echo "${DEPLOY_TX}" > "${CONTRACTS_DIR}/DEPLOY_TX"
chmod 0444 "${CONTRACTS_DIR}/CONTRACT_ADDR" "${CONTRACTS_DIR}/DEPLOY_TX" 2>/dev/null || true

log "Deploy complete:"
log "  CONTRACT_ADDR : ${CONTRACT_ADDR}"
log "  DEPLOY_TX     : ${DEPLOY_TX}"
log "  ABI file      : ${CONTRACTS_DIR}/CardToken.abi"
log ""
log "下一步（Task 2.5 / Task 3）：grant MINTER_ROLE 给 sms-rpc 运行时账号地址"
