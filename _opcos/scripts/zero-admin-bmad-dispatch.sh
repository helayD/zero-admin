#!/bin/bash
set -euo pipefail

# zero-admin BMAD dispatcher (v1 draft)
# 作用：
# 1. 读取 BMAD / git / autopilot 状态
# 2. 计算当前最优下一步
# 3. 只派发一个明确的 BMAD 节点任务给 codex-autopilot
#
# 说明：
# - 当前版本是“可运行骨架 + 明确决策逻辑”的草案
# - 默认只做最小安全派发，不做激进自动推进
# - 若检测到阻塞（登录页/权限页/测试红线/工件缺失），只记录，不继续派发

PROJECT_ROOT="/Users/helay/Documents/GitHub/zero-admin"
AUTOPILOT_ROOT="$HOME/.autopilot"
SESSION_NAME="autopilot"
WINDOW_NAME="zero-admin"
STATE_DIR="$PROJECT_ROOT/_opcos/automation-state"
LOG_FILE="$STATE_DIR/dispatcher.log"
LOCK_FILE="$STATE_DIR/dispatcher.lock"
ACTIVE_ACTION_FILE="$STATE_DIR/active-action.json"
SPRINT_STATUS="$PROJECT_ROOT/_opcos/implementation-artifacts/sprint-status.yaml"
WATCHDOG_LOG="$AUTOPILOT_ROOT/logs/watchdog.log"
AUTOPILOT_STATE="$AUTOPILOT_ROOT/state/zero-admin.json"
TMUX_BIN="/opt/homebrew/bin/tmux"
TMUX_SEND="$AUTOPILOT_ROOT/scripts/tmux-send.sh"
TASK_QUEUE="$AUTOPILOT_ROOT/scripts/task-queue.sh"

mkdir -p "$STATE_DIR"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*" | tee -a "$LOG_FILE"
}

with_lock() {
  if ! mkdir "$LOCK_FILE" 2>/dev/null; then
    log "dispatcher already running, skip"
    exit 0
  fi
  trap 'rm -rf "$LOCK_FILE"' EXIT
}

require_file() {
  local f="$1"
  [ -f "$f" ] || { log "missing required file: $f"; exit 0; }
}

get_first_ready_story() {
  python3 - <<'PY'
import yaml, os
path = os.path.expanduser('/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/sprint-status.yaml')
with open(path, 'r', encoding='utf-8') as f:
    data = yaml.safe_load(f)
for k, v in (data.get('development_status') or {}).items():
    if isinstance(k, str) and '-' in k and not k.startswith('epic-') and v == 'ready-for-dev':
        print(k)
        break
PY
}

get_first_in_progress_story() {
  python3 - <<'PY'
import yaml, os
path = os.path.expanduser('/Users/helay/Documents/GitHub/zero-admin/_opcos/implementation-artifacts/sprint-status.yaml')
with open(path, 'r', encoding='utf-8') as f:
    data = yaml.safe_load(f)
for k, v in (data.get('development_status') or {}).items():
    if isinstance(k, str) and '-' in k and not k.startswith('epic-') and v == 'in-progress':
        print(k)
        break
PY
}

story_path_for_key() {
  local key="$1"
  find "$PROJECT_ROOT/_opcos/implementation-artifacts" -maxdepth 1 -type f -name "$key*.md" | head -n1
}

has_uncommitted_changes() {
  cd "$PROJECT_ROOT"
  [ -n "$(git status --short)" ]
}

current_active_action() {
  [ -f "$ACTIVE_ACTION_FILE" ] || return 0
  python3 - <<'PY'
import json, os
path=os.path.expanduser('/Users/helay/Documents/GitHub/zero-admin/_opcos/automation-state/active-action.json')
with open(path,'r',encoding='utf-8') as f:
    data=json.load(f)
print(data.get('action',''))
PY
}

set_active_action() {
  local action="$1"
  python3 - <<PY
import json, os, time
path = os.path.expanduser('$ACTIVE_ACTION_FILE')
data = {
  'action': '$action',
  'updated_at': int(time.time())
}
with open(path, 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
PY
}

clear_active_action() {
  rm -f "$ACTIVE_ACTION_FILE"
}

watchdog_indicates_block() {
  [ -f "$WATCHDOG_LOG" ] || return 1
  tail -n 80 "$WATCHDOG_LOG" | grep -E "Sign in with ChatGPT|permission detected|shell recovery" >/dev/null 2>&1
}

choose_action() {
  local ready_story in_progress_story
  ready_story="$(get_first_ready_story || true)"
  in_progress_story="$(get_first_in_progress_story || true)"

  if watchdog_indicates_block; then
    echo "blocked:watchdog"
    return 0
  fi

  if [ -n "$in_progress_story" ] && has_uncommitted_changes; then
    echo "analyze-current-state:$in_progress_story"
    return 0
  fi

  if [ -n "$ready_story" ]; then
    local story_path
    story_path="$(story_path_for_key "$ready_story")"
    if [ -z "$story_path" ]; then
      echo "create-story"
      return 0
    fi
    if has_uncommitted_changes; then
      echo "analyze-current-state:$ready_story"
      return 0
    fi
    echo "dev-story:$ready_story"
    return 0
  fi

  echo "create-story"
}

build_prompt() {
  local action="$1"
  case "$action" in
    create-story)
      cat <<'EOF'
按 zero-admin 的 BMAD 规则执行：create-story。
要求：
1. 只推进当前最优下一条 backlog story
2. 生成 story 文件并更新 sprint-status 到 ready-for-dev
3. 不跨到 dev-story / review
4. 完成后给出生成结果与 story key
EOF
      ;;
    dev-story:*)
      local story_key="${action#dev-story:}"
      cat <<EOF
按 zero-admin 的 BMAD 规则执行：dev-story ${story_key}。
要求：
1. 仅围绕 ${story_key} 推进
2. 不跨到下一条 story
3. 未完成不伪装完成
4. 若遇阻塞，明确给出最小阻塞点
EOF
      ;;
    analyze-current-state:*)
      local story_key="${action#analyze-current-state:}"
      cat <<EOF
按 zero-admin 的 BMAD 规则执行：analyze-current-state ${story_key}。
要求：
1. 读取当前 story、sprint-status、git status、相关改动
2. 判断已完成项、未完成项、是否可继续开发或应先收口
3. 输出当前最优下一步
4. 不直接跨 story 扩张
EOF
      ;;
    blocked:watchdog)
      ;;
  esac
}

dispatch_action() {
  local action="$1"
  local prompt
  prompt="$(build_prompt "$action")"

  if [ -z "$prompt" ]; then
    log "no prompt built for action=$action"
    return 0
  fi

  if [ -x "$TASK_QUEUE" ]; then
    "$TASK_QUEUE" enqueue "$WINDOW_NAME" "$prompt" >/dev/null 2>&1 || true
    log "queued action via task-queue: $action"
  elif [ -x "$TMUX_SEND" ]; then
    "$TMUX_SEND" "$WINDOW_NAME" "$prompt" >/dev/null 2>&1 || true
    log "sent action via tmux-send: $action"
  else
    "$TMUX_BIN" has-session -t "$SESSION_NAME" >/dev/null 2>&1 || {
      log "tmux session $SESSION_NAME not found"
      return 0
    }
    "$TMUX_BIN" send-keys -t "$SESSION_NAME:$WINDOW_NAME" -l -- "$prompt"
    "$TMUX_BIN" send-keys -t "$SESSION_NAME:$WINDOW_NAME" Enter
    log "sent action via raw tmux: $action"
  fi

  set_active_action "$action"
}

main() {
  with_lock
  require_file "$SPRINT_STATUS"

  local action active
  action="$(choose_action)"
  active="$(current_active_action || true)"

  log "chosen action=$action active=$active"

  if [[ "$action" == blocked:* ]]; then
    log "blocked, no dispatch: $action"
    exit 0
  fi

  if [ -n "$active" ] && [ "$active" = "$action" ]; then
    log "same active action already recorded, skip duplicate dispatch"
    exit 0
  fi

  dispatch_action "$action"
}

main "$@"
