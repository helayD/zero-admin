#!/usr/bin/env bash
# Story 10.11 / Task 9.4：FISCO BCOS 3.x 紧急回滚脚本。
#
# 适用场景：
#   - 生产 FISCO 节点宕机 / 证书过期 / 合约 bug，需要立刻让三服务停止再调真实链；
#   - 仍然要保证业务主链路（订单、卡片、任务表）继续工作；
#   - 新创建的发放任务自动停在 pending_dispatch + mint_status=mint_compensating，
#     等 FISCO 恢复后再 把 Fisco.Enabled 切回 true，三服务自循环扫描自动追发。
#
# 设计原则：
#   - 只关闭 Fisco.Enabled，不删配置（节点地址 / 合约地址 / 证书保留），方便恢复。
#   - 不操作数据库（已成功的 chain_tx_id / token_id 不动）。
#   - 严格 idempotent：重复执行无副作用。
#
# 用法：
#   bash scripts/rollback-fisco-to-disabled.sh [--remote]
#     默认仅修改本地 etc/*.yaml；加 --remote 走 SSH 修改远程 47.107.224.56:/root/zero-admin/target 并重启服务。
#
# 远程模式依赖：
#   - sshpass 已安装
#   - 环境变量 SSHPASS 已设置（生产环境密码）
#
# 验证：
#   ssh 47.107.224.56 "grep -A1 'Fisco:' /root/zero-admin/target/sms-rpc/sms-rpc.yaml | head"
#   应该看到 Enabled: false。

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

REMOTE_MODE=false
for arg in "$@"; do
    case "$arg" in
        --remote)
            REMOTE_MODE=true
            ;;
        *)
            echo "未知参数: $arg" >&2
            exit 2
            ;;
    esac
done

CONFIG_FILES=(
    "rpc/sms/etc/sms.yaml"
    "consumer/etc/consumer-api.yaml"
    "job/etc/job-api.yaml"
)

# 把 yaml 中 Fisco: 节下的 Enabled: true 改为 Enabled: false。
# 使用 awk 而不是 sed -i 处理跨平台差异（macOS 默认 BSD sed）。
disable_local() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        echo "⚠️  跳过不存在的文件: $file"
        return 0
    fi
    awk '
        BEGIN { in_fisco=0; modified=0 }
        /^Fisco:/ { in_fisco=1; print; next }
        in_fisco && /^[A-Za-z]/ { in_fisco=0 }
        in_fisco && /^[[:space:]]+Enabled:[[:space:]]*true/ {
            sub(/Enabled:[[:space:]]*true/, "Enabled: false")
            modified=1
        }
        { print }
        END {
            if (modified) {
                print "修改: 已将 Fisco.Enabled 设置为 false" > "/dev/stderr"
            } else {
                print "无需修改: Fisco.Enabled 已是 false 或未找到该字段" > "/dev/stderr"
            }
        }
    ' "$file" > "$file.tmp"
    mv "$file.tmp" "$file"
    echo "✅ 处理完成: $file"
}

if [[ "$REMOTE_MODE" == "true" ]]; then
    if [[ -z "${SSHPASS:-}" ]]; then
        echo "❌ 远程模式需要设置 SSHPASS 环境变量" >&2
        exit 3
    fi
    SSH_CMD='sshpass -e ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56'
    REMOTE_BASE="/root/zero-admin/target"
    REMOTE_FILES=(
        "$REMOTE_BASE/sms-rpc/sms-rpc.yaml"
        "$REMOTE_BASE/consumer/consumer-api.yaml"
        "$REMOTE_BASE/job/job-api.yaml"
    )
    echo "→ 远程修改 yaml..."
    for f in "${REMOTE_FILES[@]}"; do
        # shellcheck disable=SC2029
        $SSH_CMD "sed -i 's/Enabled: true/Enabled: false/' $f && echo $f done"
    done
    echo "→ 重启 sms-rpc / consumer / job..."
    # shellcheck disable=SC2029
    $SSH_CMD "pkill -9 -f 'sms-rpc|consumer|job' || true; sleep 2; cd $REMOTE_BASE && \
        nohup ./sms-rpc/sms-rpc -f ./sms-rpc/sms-rpc.yaml > ./logs/sms-rpc.log 2>&1 & \
        nohup ./consumer/consumer -f ./consumer/consumer-api.yaml > ./logs/consumer.log 2>&1 & \
        nohup ./job/job -f ./job/job-api.yaml > ./logs/job.log 2>&1 &"
    echo "✅ 远程回滚完成。请在 1 分钟后用日志验证："
    echo "   ssh 47.107.224.56 'tail -50 /root/zero-admin/target/logs/sms-rpc.log | grep buildChainClient'"
    echo "   预期：Fisco.Enabled=false"
else
    for file in "${CONFIG_FILES[@]}"; do
        disable_local "$file"
    done
    echo ""
    echo "✅ 本地回滚完成。下一步："
    echo "   1) 重新构建：make build"
    echo "   2) 重启三服务"
    echo "   3) 在日志中确认 buildChainClient: chainType=\"fisco_bcos_3x\" Enabled=false"
fi
