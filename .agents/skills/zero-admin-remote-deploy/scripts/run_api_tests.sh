#!/usr/bin/env bash
# =============================================================================
# Story API 测试运行器
# 自动发现并执行 script/shell/api-test/ 下所有 story 测试脚本
# 要求每个脚本 100% 通过，否则整体判定失败
#
# 用法:
#   bash run_api_tests.sh [options]
#
# Options:
#   --admin-url <url>   Admin API base URL (default: http://47.107.224.56:8000)
#   --front-url <url>   Front API base URL (default: http://47.107.224.56:9999)
#   --test-dir <path>   API test directory  (default: <repo>/script/shell/api-test)
#   --stories <csv>     Only run specified stories, e.g. "4-5,4-6,5-1"
#   --fail-fast         Stop on first story failure
#   -h, --help          Show this help
#
# Exit code:
#   0 = all stories 100% passed
#   1 = at least one story had failures
# =============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

ADMIN_URL="${ZERO_ADMIN_SMOKE_BASE_URL:-http://47.107.224.56:8000}"
FRONT_URL="${ZERO_ADMIN_FRONT_SMOKE_BASE_URL:-http://47.107.224.56:9999}"
TEST_DIR=""
STORIES_CSV=""
FAIL_FAST=0

usage() {
  sed -n '2,/^# =====/{ /^# =====/d; s/^# \?//; p; }' "$0"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --admin-url) ADMIN_URL="$2"; shift 2 ;;
    --front-url) FRONT_URL="$2"; shift 2 ;;
    --test-dir)  TEST_DIR="$2";  shift 2 ;;
    --stories)   STORIES_CSV="$2"; shift 2 ;;
    --fail-fast) FAIL_FAST=1;    shift ;;
    -h|--help)   usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 1 ;;
  esac
done

[[ -z "$TEST_DIR" ]] && TEST_DIR="$REPO_ROOT/script/shell/api-test"

if [[ ! -d "$TEST_DIR" ]]; then
  echo "API test directory not found: $TEST_DIR" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Story 与 API base URL 的映射规则：
# 优先从测试脚本头部读取 "# API_TYPE: admin" 或 "# API_TYPE: front" 标记
# 回退规则：根据 story 目录名前缀猜测（仅 admin-api 场景使用 4-6 等标记）
# 每个 story 测试脚本接受一个 base URL 参数
# ---------------------------------------------------------------------------
resolve_base_url_from_script() {
  local script_path="$1"
  local api_type
  api_type=$(grep -m1 '^# API_TYPE:' "$script_path" 2>/dev/null | sed 's/^# API_TYPE:[[:space:]]*//' | tr -d '[:space:]')
  case "$api_type" in
    admin) echo "$ADMIN_URL" ;;
    front) echo "$FRONT_URL" ;;
    *)     echo "$FRONT_URL" ;;  # default to front
  esac
}

# Collect story directories
STORY_DIRS=()
if [[ -n "$STORIES_CSV" ]]; then
  IFS=',' read -r -a requested <<< "$STORIES_CSV"
  for s in "${requested[@]}"; do
    s="$(echo "$s" | xargs)"  # trim
    dir="$TEST_DIR/$s"
    if [[ ! -d "$dir" ]]; then
      echo "Story directory not found: $dir" >&2
      exit 1
    fi
    STORY_DIRS+=("$dir")
  done
else
  # Auto-discover: sort by story id for deterministic order
  while IFS= read -r d; do
    STORY_DIRS+=("$d")
  done < <(find "$TEST_DIR" -mindepth 1 -maxdepth 1 -type d | sort)
fi

if [[ ${#STORY_DIRS[@]} -eq 0 ]]; then
  echo "No story test directories found in $TEST_DIR" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Run
# ---------------------------------------------------------------------------
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'
BOLD='\033[1m'

TOTAL_STORIES=0
PASSED_STORIES=0
FAILED_STORIES=0
FAILED_STORY_NAMES=()

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║        Story API 测试套件                    ║${NC}"
echo -e "${BOLD}╠══════════════════════════════════════════════╣${NC}"
echo -e "${BOLD}║${NC} Admin URL : $ADMIN_URL"
echo -e "${BOLD}║${NC} Front URL : $FRONT_URL"
echo -e "${BOLD}║${NC} Stories   : ${#STORY_DIRS[@]}"
echo -e "${BOLD}║${NC} Time      : $(date '+%Y-%m-%d %H:%M:%S')"
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"
echo ""

for story_dir in "${STORY_DIRS[@]}"; do
  story_name="$(basename "$story_dir")"
  # Find test script(s) in this story directory
  test_scripts=()
  while IFS= read -r f; do
    test_scripts+=("$f")
  done < <(find "$story_dir" -maxdepth 1 -name 'test_*.sh' -type f | sort)

  if [[ ${#test_scripts[@]} -eq 0 ]]; then
    echo -e "${YELLOW}⚠  Story $story_name: no test scripts found, skipping${NC}"
    continue
  fi

  for script in "${test_scripts[@]}"; do
    ((TOTAL_STORIES++)) || true
    script_name="$(basename "$script")"
    base_url="$(resolve_base_url_from_script "$script")"
    echo -e "${BOLD}━━━ Story $story_name ($script_name) → $base_url ━━━${NC}"

    # 清除 SMS 冷却 key，避免多 story 共用同一手机号时触发 60s 频控
    SSHPASS='Qianmai1#' sshpass -e ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no \
      -o ConnectTimeout=5 root@47.107.224.56 \
      'redis-cli -p 16379 -a 123456 --no-auth-warning KEYS "ums:sms:cooldown:*" | xargs -r redis-cli -p 16379 -a 123456 --no-auth-warning DEL' \
      2>/dev/null || true

    set +e
    output=$(bash "$script" "$base_url" 2>&1)
    exit_code=$?
    set -e

    echo "$output"

    if [[ $exit_code -eq 0 ]]; then
      ((PASSED_STORIES++)) || true
      echo -e "  ${GREEN}▸ Story $story_name: PASSED${NC}"
    else
      ((FAILED_STORIES++)) || true
      FAILED_STORY_NAMES+=("$story_name")
      echo -e "  ${RED}▸ Story $story_name: FAILED (exit=$exit_code)${NC}"

      if [[ $FAIL_FAST -eq 1 ]]; then
        echo -e "\n${RED}--fail-fast: stopping after first failure${NC}"
        break 2
      fi
    fi
    echo ""
  done
done

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║        测试套件总结                          ║${NC}"
echo -e "${BOLD}╠══════════════════════════════════════════════╣${NC}"
echo -e "${BOLD}║${NC} 总计 Stories : $TOTAL_STORIES"
echo -e "${BOLD}║${NC} 通过         : ${GREEN}$PASSED_STORIES${NC}"
echo -e "${BOLD}║${NC} 失败         : ${RED}$FAILED_STORIES${NC}"
if [[ ${#FAILED_STORY_NAMES[@]} -gt 0 ]]; then
  echo -e "${BOLD}║${NC} 失败清单     : ${RED}$(IFS=', '; echo "${FAILED_STORY_NAMES[*]}")${NC}"
fi
echo -e "${BOLD}╚══════════════════════════════════════════════╝${NC}"

if [[ $FAILED_STORIES -gt 0 ]]; then
  echo -e "\n${RED}API 测试未达到 100% 通过率，部署验证失败${NC}"
  exit 1
else
  echo -e "\n${GREEN}所有 Story API 测试 100% 通过 ✓${NC}"
  exit 0
fi
