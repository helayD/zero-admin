#!/usr/bin/env bash
# query_logs.sh — Zero-Admin 远程日志查看脚本
# 用法: query_logs.sh <service|all> <log-type> [lines]
#
# service:  admin | front | sys | ums | pms | oms | sms | cms | search | consumer | job | all
# log-type: access | error | slow | severe | restart
# lines:    默认 100

set -euo pipefail

SSH_TARGET="${ZERO_ADMIN_SSH_TARGET:-root@47.107.224.56}"
SSH_PASS="${ZERO_ADMIN_SSH_PASS:-Qianmai1#}"
LOGS_ROOT="/root/zero-admin/target/logs"

# 数据库连接信息
MYSQL_HOST="${ZERO_ADMIN_MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${ZERO_ADMIN_MYSQL_PORT:-3306}"
MYSQL_USER="${ZERO_ADMIN_MYSQL_USER:-root}"
MYSQL_PASS="${ZERO_ADMIN_MYSQL_PASS:-12341qweqfsd2356}"
MYSQL_DB="${ZERO_ADMIN_MYSQL_DB:-gozero}"
REDIS_HOST="${ZERO_ADMIN_REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${ZERO_ADMIN_REDIS_PORT:-16379}"
REDIS_PASS="${ZERO_ADMIN_REDIS_PASS:-123456}"

SERVICE="${1:-}"
LOG_TYPE="${2:-error}"
LINES="${3:-100}"

SSH_CMD="sshpass -p '${SSH_PASS}' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no ${SSH_TARGET}"

# 服务名到日志子目录映射
declare -A SVC_DIR_MAP=(
  [admin-api]="admin"
  [admin]="admin"
  [front-api]="front"
  [front]="front"
  [sys-rpc]="sys"
  [sys]="sys"
  [ums-rpc]="ums"
  [ums]="ums"
  [pms-rpc]="pms"
  [pms]="pms"
  [oms-rpc]="oms"
  [oms]="oms"
  [sms-rpc]="sms"
  [sms]="sms"
  [cms-rpc]="cms"
  [cms]="cms"
  [search-rpc]="search"
  [search]="search"
  [consumer]="consumer"
  [job]="job"
)

ALL_SERVICES=(admin front sys ums pms oms sms cms search consumer job)

RESTART_LOG_MAP=(
  [admin]=""
  [front]="front-api-restart.log"
  [sys]=""
  [ums]="ums-rpc.log"
  [pms]=""
  [oms]=""
  [sms]="sms-rpc-restart.log"
  [cms]=""
  [search]=""
  [consumer]="consumer-restart.log"
  [job]="job-restart.log"
)

usage() {
  echo "用法: $0 <service|all|db|redis> <log-type|sql|cmd> [lines]"
  echo ""
  echo "  service:   admin | front | sys | ums | pms | oms | sms | cms | search | consumer | job | all"
  echo "  log-type:  access | error | slow | severe | restart | list"
  echo "  lines:     默认 100"
  echo ""
  echo "数据库模式:"
  echo "  $0 db \"<SQL>\"            # 远程执行 MySQL SQL（数据库: gozero）"
  echo "  $0 redis \"<CMD>\"         # 远程执行 Redis 命令"
  echo ""
  echo "示例:"
  echo "  $0 sms error 50          # 查 sms-rpc 最近 50 条错误"
  echo "  $0 all error 20          # 所有服务最近 20 条错误（巡检）"
  echo "  $0 job restart 100       # job 服务重启日志"
  echo "  $0 admin list            # 列出 admin-api 所有日志文件"
  echo "  $0 db \"SELECT id, mint_status FROM sms_card_instance ORDER BY id DESC LIMIT 10\""
  echo "  $0 redis \"SCAN 0 MATCH '*token*' COUNT 20\""
}

if [[ -z "$SERVICE" ]]; then
  usage
  exit 1
fi

query_service_log() {
  local svc_dir="$1"
  local log_type="$2"
  local lines="$3"

  local log_file=""
  case "$log_type" in
    access)  log_file="${LOGS_ROOT}/${svc_dir}/access.log" ;;
    error)   log_file="${LOGS_ROOT}/${svc_dir}/error.log" ;;
    slow)    log_file="${LOGS_ROOT}/${svc_dir}/slow.log" ;;
    severe)  log_file="${LOGS_ROOT}/${svc_dir}/severe.log" ;;
    list)
      echo "=== ${svc_dir} 日志文件列表 ==="
      eval "${SSH_CMD} 'ls -lh ${LOGS_ROOT}/${svc_dir}/'"
      return
      ;;
    restart)
      local restart_file="${RESTART_LOG_MAP[$svc_dir]:-}"
      if [[ -z "$restart_file" ]]; then
        echo "[WARN] ${svc_dir} 没有独立的重启日志文件，显示 error.log 最后 ${lines} 行"
        log_file="${LOGS_ROOT}/${svc_dir}/error.log"
      else
        log_file="${LOGS_ROOT}/${restart_file}"
      fi
      ;;
    *)
      echo "[ERROR] 未知日志类型: ${log_type}，支持: access | error | slow | severe | restart | list"
      exit 1
      ;;
  esac

  echo "=== ${svc_dir} / ${log_type}.log (最近 ${lines} 行) ==="
  eval "${SSH_CMD} 'tail -${lines} ${log_file} 2>/dev/null || echo \"[文件不存在或为空: ${log_file}]\"'"
  echo ""
}

# db 子命令：远程执行 MySQL SQL
if [[ "$SERVICE" == "db" ]]; then
  SQL="${LOG_TYPE:-}"
  if [[ -z "$SQL" ]]; then
    echo "[ERROR] 请提供 SQL，例如: $0 db \"SELECT 1\""
    exit 1
  fi
  echo "=== MySQL @ ${MYSQL_HOST}:${MYSQL_PORT}/${MYSQL_DB} ==="
  sshpass -p "${SSH_PASS}" ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no "${SSH_TARGET}" \
    "mysql -h${MYSQL_HOST} -P${MYSQL_PORT} -u${MYSQL_USER} -p'${MYSQL_PASS}' ${MYSQL_DB} -e \"${SQL}\""
  exit 0
fi

# redis 子命令：远程执行 Redis 命令
if [[ "$SERVICE" == "redis" ]]; then
  CMD="${LOG_TYPE:-}"
  if [[ -z "$CMD" ]]; then
    echo "[ERROR] 请提供 Redis 命令，例如: $0 redis \"DBSIZE\""
    exit 1
  fi
  echo "=== Redis @ ${REDIS_HOST}:${REDIS_PORT} ==="
  sshpass -p "${SSH_PASS}" ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no "${SSH_TARGET}" \
    "redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} -a '${REDIS_PASS}' ${CMD}"
  exit 0
fi

# 解析服务名
if [[ "$SERVICE" == "all" ]]; then
  if [[ "$LOG_TYPE" == "list" ]]; then
    echo "=== 所有日志文件总览 ==="
    eval "${SSH_CMD} 'ls -lh ${LOGS_ROOT}/'"
    exit 0
  fi
  for svc in "${ALL_SERVICES[@]}"; do
    query_service_log "$svc" "$LOG_TYPE" "$LINES"
  done
else
  SVC_DIR="${SVC_DIR_MAP[$SERVICE]:-}"
  if [[ -z "$SVC_DIR" ]]; then
    echo "[ERROR] 未知服务: ${SERVICE}"
    echo "支持的服务: ${!SVC_DIR_MAP[*]}"
    exit 1
  fi
  query_service_log "$SVC_DIR" "$LOG_TYPE" "$LINES"
fi
