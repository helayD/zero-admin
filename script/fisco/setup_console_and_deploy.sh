#!/bin/bash
# Story 10.11 / Task 2.3 - 在 47.107.224.56 安装 console2 v3.7.0 并部署 CardToken 合约。
#
# 前置：
#   1. /tmp/console.tar.gz 已上传（144MB）
#   2. /tmp/CardToken.sol 已上传
#   3. /opt/fisco/sdk/{ca.crt,sdk.crt,sdk.key} 已就绪 (Task 1 产出)
#   4. 4 个节点已运行 (127.0.0.1:20200-20203)
#
# 可选环境变量：
#   SMS_MINTER_ADDR  sms-rpc 运行时账号地址（0x... 40 hex），
#                    设置后脚本会在部署完自动调用 grantMinter；
#                    未设置则跳过，运维需稍后手动 grantMinter。
#   FORCE_REDEPLOY   设 1 时即使 CONTRACT_ADDR 已存在也强制重部署，
#                    旧地址会备份到 CONTRACT_ADDR.bak.<unix-ts>。
#
# 执行后产出：
#   /opt/fisco/console/                              console2 安装目录
#   /opt/fisco/console/conf/config.toml              已配置好节点 + 证书路径
#   /opt/fisco/contracts/CardToken.sol               合约源码副本
#   /opt/fisco/contracts/CardToken.abi               合约 ABI（供 Go SDK 加载，从 console 输出回抓）
#   /opt/fisco/contracts/CardToken.bin               合约字节码副本
#   /opt/fisco/contracts/CONTRACT_ADDR               部署后合约地址
#   /opt/fisco/contracts/DEPLOY_TX                   部署交易 hash
#   /opt/fisco/contracts/MINTER_GRANTED              grantMinter 已执行的 minter 地址列表

set -euo pipefail

FISCO_HOME="${FISCO_HOME:-/opt/fisco}"
NODE_HOST="${NODE_HOST:-127.0.0.1}"
RPC_PORT_BASE="${RPC_PORT_BASE:-20200}"
CONSOLE_DIR="${FISCO_HOME}/console"
CONTRACTS_DIR="${FISCO_HOME}/contracts"
SMS_MINTER_ADDR="${SMS_MINTER_ADDR:-}"
FORCE_REDEPLOY="${FORCE_REDEPLOY:-0}"
CONSOLE_TIMEOUT="${CONSOLE_TIMEOUT:-90}"

log() { printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"; }

# run_console: 通过 stdin 多行喂入命令并显式 quit 退出，避免 console2 JLine REPL
# 在 EOF 时丢命令或 hang 住。所有命令通过 timeout 守卫。
# 用法: run_console <out_file> "<console2 命令>"
#       run_console 只接受一条命令；多条命令请连续多次调用。
run_console() {
    local out_file="$1"
    local cmd="$2"
    cd "${CONSOLE_DIR}"
    {
        printf '%s\n' "${cmd}"
        printf 'quit\n'
    } | timeout "${CONSOLE_TIMEOUT}" bash console.sh > "${out_file}" 2>&1
}

if [ "$(id -u)" -ne 0 ]; then
    echo "请以 root 运行"; exit 1
fi

# ---- 0. 幂等守卫：已部署且未指定 FORCE_REDEPLOY 则直接退出 ----
if [ -f "${CONTRACTS_DIR}/CONTRACT_ADDR" ] && [ "${FORCE_REDEPLOY}" != "1" ]; then
    EXISTING_ADDR=$(cat "${CONTRACTS_DIR}/CONTRACT_ADDR" 2>/dev/null || true)
    log "已部署: ${EXISTING_ADDR}，跳过 deploy"
    log "如需强制重新部署，请：FORCE_REDEPLOY=1 bash $0"
    log "或手动: mv ${CONTRACTS_DIR}/CONTRACT_ADDR ${CONTRACTS_DIR}/CONTRACT_ADDR.bak.\$(date +%s)"
    # 即使跳过部署，仍允许补 grantMinter（如果 SMS_MINTER_ADDR 给了且尚未授权）
    if [ -n "${SMS_MINTER_ADDR}" ] && ! grep -qFx "${SMS_MINTER_ADDR}" "${CONTRACTS_DIR}/MINTER_GRANTED" 2>/dev/null; then
        log "检测到 SMS_MINTER_ADDR=${SMS_MINTER_ADDR} 尚未授权，仅执行 grantMinter"
        GRANT_OUT="${CONTRACTS_DIR}/grant_minter.$(date +%s).out"
        run_console "${GRANT_OUT}" "call CardToken ${EXISTING_ADDR} grantMinter ${SMS_MINTER_ADDR}" || true
        cat "${GRANT_OUT}"
        if grep -q '"status": "0x0"\|status: 0x0\|transaction status: 0' "${GRANT_OUT}"; then
            echo "${SMS_MINTER_ADDR}" >> "${CONTRACTS_DIR}/MINTER_GRANTED"
            log "  grantMinter 成功: ${SMS_MINTER_ADDR}"
        else
            log "  WARN: grantMinter 输出疑似失败，请人工核对 ${GRANT_OUT}"
        fi
    fi
    exit 0
fi

# 强制重部署模式：备份旧地址
if [ "${FORCE_REDEPLOY}" = "1" ] && [ -f "${CONTRACTS_DIR}/CONTRACT_ADDR" ]; then
    BACKUP_TS=$(date +%s)
    mv "${CONTRACTS_DIR}/CONTRACT_ADDR" "${CONTRACTS_DIR}/CONTRACT_ADDR.bak.${BACKUP_TS}"
    [ -f "${CONTRACTS_DIR}/DEPLOY_TX" ] && mv "${CONTRACTS_DIR}/DEPLOY_TX" "${CONTRACTS_DIR}/DEPLOY_TX.bak.${BACKUP_TS}"
    [ -f "${CONTRACTS_DIR}/MINTER_GRANTED" ] && mv "${CONTRACTS_DIR}/MINTER_GRANTED" "${CONTRACTS_DIR}/MINTER_GRANTED.bak.${BACKUP_TS}"
    log "FORCE_REDEPLOY=1: 旧地址备份到 ${CONTRACTS_DIR}/CONTRACT_ADDR.bak.${BACKUP_TS}"
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
CONNECT_OUT="/tmp/console_check.log"
run_console "${CONNECT_OUT}" "getBlockNumber" || true
grep -E "Block|Error|exception|info|Welcome" "${CONNECT_OUT}" | head -10 || true
if grep -q "Welcome to FISCO" "${CONNECT_OUT}"; then
    log "  console banner ok"
else
    log "  WARN: console2 banner 缺失，请确认 console.sh 是否能正常启动"
fi

# ---- 7. 部署 CardToken ----
log "Step 7: deploy CardToken to group0"
DEPLOY_OUT="${CONTRACTS_DIR}/deploy.out"
run_console "${DEPLOY_OUT}" "deploy CardToken" || true
cat "${DEPLOY_OUT}"

CONTRACT_ADDR=$(grep -oE "contract address: 0x[0-9a-fA-F]{40}" "${DEPLOY_OUT}" | head -1 | awk '{print $3}')
DEPLOY_TX=$(grep -oE "transaction hash: 0x[0-9a-fA-F]{64}" "${DEPLOY_OUT}" | head -1 | awk '{print $3}')

if [ -z "${CONTRACT_ADDR}" ]; then
    log "ERROR: CONTRACT_ADDR not found in deploy output, check ${DEPLOY_OUT}"
    exit 5
fi

echo "${CONTRACT_ADDR}" > "${CONTRACTS_DIR}/CONTRACT_ADDR"
echo "${DEPLOY_TX}"     > "${CONTRACTS_DIR}/DEPLOY_TX"
chmod 0444 "${CONTRACTS_DIR}/CONTRACT_ADDR" "${CONTRACTS_DIR}/DEPLOY_TX" 2>/dev/null || true

# 从 console 输出目录回抓最新 ABI / bin（部署后是权威源），覆盖 Step 5 早先从 /tmp 拷贝的版本
CONSOLE_BUILD_DIR="${CONSOLE_DIR}/contracts/solidity"
for candidate in \
    "${CONSOLE_BUILD_DIR}/CardToken/CardToken.abi" \
    "${CONSOLE_BUILD_DIR}/sm/CardToken/CardToken.abi" \
    "${CONSOLE_BUILD_DIR}/CardToken.abi"; do
    if [ -f "${candidate}" ]; then
        cp -f "${candidate}" "${CONTRACTS_DIR}/CardToken.abi"
        log "  ABI 已从 console 输出回抓: ${candidate}"
        break
    fi
done
for candidate in \
    "${CONSOLE_BUILD_DIR}/CardToken/CardToken.bin" \
    "${CONSOLE_BUILD_DIR}/sm/CardToken/CardToken.bin" \
    "${CONSOLE_BUILD_DIR}/CardToken.bin"; do
    if [ -f "${candidate}" ]; then
        cp -f "${candidate}" "${CONTRACTS_DIR}/CardToken.bin"
        break
    fi
done

# ---- 8. 部署后链上自检：调 name() 应返回 "Zero-Admin CardToken" ----
log "Step 8: post-deploy on-chain self-check (call name())"
NAME_OUT="${CONTRACTS_DIR}/name_check.out"
run_console "${NAME_OUT}" "call CardToken ${CONTRACT_ADDR} name" || true
if grep -qF 'Zero-Admin CardToken' "${NAME_OUT}"; then
    log "  name() 返回正确，合约可调用"
else
    log "  WARN: name() 自检未通过，请查看 ${NAME_OUT}"
    cat "${NAME_OUT}" | head -20
fi

# ---- 9. grant MINTER_ROLE（如果指定了 SMS_MINTER_ADDR）----
if [ -n "${SMS_MINTER_ADDR}" ]; then
    log "Step 9: grantMinter ${SMS_MINTER_ADDR}"
    if ! [[ "${SMS_MINTER_ADDR}" =~ ^0x[0-9a-fA-F]{40}$ ]]; then
        log "ERROR: SMS_MINTER_ADDR 格式非法 (期望 0x + 40 hex)：${SMS_MINTER_ADDR}"
        exit 6
    fi
    GRANT_OUT="${CONTRACTS_DIR}/grant_minter.out"
    run_console "${GRANT_OUT}" "call CardToken ${CONTRACT_ADDR} grantMinter ${SMS_MINTER_ADDR}" || true
    cat "${GRANT_OUT}"
    # console2 输出格式：transaction status: 0x0 表示成功；output 行也会包含 result success。
    if grep -qE 'transaction status: 0x0|status: 0x0|"status": "0x0"' "${GRANT_OUT}"; then
        echo "${SMS_MINTER_ADDR}" >> "${CONTRACTS_DIR}/MINTER_GRANTED"
        log "  grantMinter 成功: ${SMS_MINTER_ADDR}"
    else
        log "  WARN: grantMinter 输出未匹配成功标志，请人工核对 ${GRANT_OUT}"
    fi
else
    log "Step 9: 跳过 grantMinter（SMS_MINTER_ADDR 未设置）"
    log "        请稍后用以下命令补授权："
    log "        SMS_MINTER_ADDR=0x<addr> bash $0"
fi

log ""
log "Deploy complete:"
log "  CONTRACT_ADDR : ${CONTRACT_ADDR}"
log "  DEPLOY_TX     : ${DEPLOY_TX}"
log "  ABI file      : ${CONTRACTS_DIR}/CardToken.abi"
if [ -n "${SMS_MINTER_ADDR}" ]; then
    log "  MINTER        : ${SMS_MINTER_ADDR}"
fi
log ""
log "下一步：把 CONTRACT_ADDR 写入三个 yaml 的 Fisco.ContractAddr 字段，重启服务"
log "  rpc/sms/etc/sms.yaml"
log "  consumer/etc/consumer-api.yaml"
log "  job/etc/job-api.yaml"
log "快捷命令（在主机上执行）："
log "  CONTRACT=\$(cat ${CONTRACTS_DIR}/CONTRACT_ADDR)"
log "  sed -i \"s|ContractAddr: \\\"\\\"|ContractAddr: \\\"\${CONTRACT}\\\"|\" /root/zero-admin/target/*/etc/*.yaml"
