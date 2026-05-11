#!/bin/bash
# Story 10.11 / Task 1.5 - install systemd unit files for FISCO BCOS air-mode 4 nodes.
# 节点二进制由 ${FISCO_HOME}/nodes/${NODE_HOST}/nodeN/start.sh 启动；
# 我们用一个统一的 fisco-bcos.service 包住 start_all.sh / stop_all.sh，避免给每个节点
# 单独写 unit。后续如要做 per-node systemd（更细粒度健康检查），可在本 Story 之外扩展。
#
# 用法：
#   sudo bash /tmp/setup_systemd.sh
#
# 行为：
#   - 写入 /etc/systemd/system/fisco-bcos.service
#   - systemctl daemon-reload
#   - systemctl enable fisco-bcos
#   - 不自动 start，避免与 install_fisco_bcos_3x.sh 已经启动的进程冲突；
#     你确认 install 脚本里启的进程已 stop 后，再 systemctl start fisco-bcos。

set -euo pipefail

FISCO_HOME="${FISCO_HOME:-/opt/fisco}"
NODE_HOST="${NODE_HOST:-127.0.0.1}"
SERVICE_FILE="/etc/systemd/system/fisco-bcos.service"

if [ "$(id -u)" -ne 0 ]; then
    echo "请以 root 运行"
    exit 1
fi

if [ ! -x "${FISCO_HOME}/nodes/${NODE_HOST}/start_all.sh" ]; then
    echo "FISCO 节点未安装：${FISCO_HOME}/nodes/${NODE_HOST}/start_all.sh 不存在"
    exit 2
fi

cat > "${SERVICE_FILE}" <<UNIT
[Unit]
Description=FISCO BCOS 3.x 4-node group0 (air mode)
After=network.target

[Service]
Type=forking
User=root
WorkingDirectory=${FISCO_HOME}/nodes/${NODE_HOST}
ExecStart=${FISCO_HOME}/nodes/${NODE_HOST}/start_all.sh
ExecStop=${FISCO_HOME}/nodes/${NODE_HOST}/stop_all.sh
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
UNIT

chmod 0644 "${SERVICE_FILE}"

systemctl daemon-reload
systemctl enable fisco-bcos

echo "systemd unit installed: ${SERVICE_FILE}"
echo "Enabled at boot. Use:"
echo "  systemctl status  fisco-bcos"
echo "  systemctl start   fisco-bcos    # 启动 (前提：手动启动的进程已 stop)"
echo "  systemctl stop    fisco-bcos"
echo "  systemctl restart fisco-bcos"
