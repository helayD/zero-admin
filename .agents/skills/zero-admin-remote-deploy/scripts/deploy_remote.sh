#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"

SSH_TARGET="${ZERO_ADMIN_SSH_TARGET:-root@47.107.224.56}"
REMOTE_ROOT="${ZERO_ADMIN_REMOTE_ROOT:-/root/zero-admin}"
SMOKE_BASE_URL="${ZERO_ADMIN_SMOKE_BASE_URL:-http://47.107.224.56:8000}"
FRONT_SMOKE_BASE_URL="${ZERO_ADMIN_FRONT_SMOKE_BASE_URL:-http://47.107.224.56:9999}"
ADMIN_ACCOUNT="${ZERO_ADMIN_ADMIN_ACCOUNT:-admin}"
ADMIN_PASSWORD="${ZERO_ADMIN_ADMIN_PASSWORD:-123456}"

MYSQL_HOST="${ZERO_ADMIN_MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${ZERO_ADMIN_MYSQL_PORT:-3306}"
MYSQL_DB="${ZERO_ADMIN_MYSQL_DB:-gozero}"
MYSQL_USER="${ZERO_ADMIN_MYSQL_USER:-root}"
MYSQL_PASSWORD="${ZERO_ADMIN_MYSQL_PASSWORD:-12341qweqfsd2356}"

SYNC_MODE="${ZERO_ADMIN_SYNC_MODE:-auto}"
GIT_REMOTE="${ZERO_ADMIN_GIT_REMOTE:-origin}"
GIT_REMOTE_URL="${ZERO_ADMIN_GIT_REMOTE_URL:-}"
GIT_PUSH_REMOTE="${ZERO_ADMIN_GIT_PUSH_REMOTE:-$GIT_REMOTE}"
GIT_REF="${ZERO_ADMIN_GIT_REF:-}"
SNAPSHOT_BRANCH_PREFIX="${ZERO_ADMIN_SNAPSHOT_BRANCH_PREFIX:-codex/deploy-snapshot}"
REMOTE_GO_BIN="${ZERO_ADMIN_REMOTE_GO_BIN:-}"
REMOTE_NPM_BIN="${ZERO_ADMIN_REMOTE_NPM_BIN:-}"
REMOTE_BIN_AUTO="__AUTO__"

DEFAULT_SERVICES="all"
SERVICES_CSV="${ZERO_ADMIN_DEPLOY_SERVICES:-$DEFAULT_SERVICES}"
SKIP_SERVICES_CSV=""
RUN_SMOKE=1
RUN_SYNC=1
MIGRATION_PATH=""

ALL_SERVICES=(
  admin-api
  front-api
  sys-rpc
  ums-rpc
  pms-rpc
  oms-rpc
  sms-rpc
  cms-rpc
  search-rpc
  consumer
  job
  web-admin
)

SELECTED_SERVICES=()

usage() {
  cat <<'EOF'
Usage: deploy_remote.sh [options]

Options:
  --host <ssh-target>          Override SSH target. Default: root@47.107.224.56
  --remote-root <path>         Override remote project root. Default: /root/zero-admin
  --services <csv|all>         Deploy the given services. Default: all
  --skip-services <csv>        Remove services from the deploy set after --services is applied
  --migration <repo-sql>       Apply a repo SQL migration after backing up the remote DB
  --sync-mode <auto|git|rsync|none>
                               Source sync mode. Default: auto
  --git-remote <name>          Local git remote used to resolve the fetch URL. Default: origin
  --git-push-remote <target>   Local git push target used for snapshot deploys. Default: same as --git-remote
  --git-remote-url <url>       Explicit git URL to fetch on the remote host
  --git-ref <ref>              Git ref to deploy. Default: upstream ref, or auto snapshot ref in dirty auto mode
  --skip-web                   Backward-compatible shortcut for --skip-services web-admin
  --skip-admin-api             Backward-compatible shortcut for --skip-services admin-api
  --skip-sys-rpc               Backward-compatible shortcut for --skip-services sys-rpc
  --skip-smoke                 Skip post-deploy smoke checks
  --skip-sync                  Skip repo source sync to the remote host
  -h, --help                   Show this help
EOF
}

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

join_by() {
  local delimiter="$1"
  shift || true
  local result=""
  local item
  for item in "$@"; do
    if [[ -z "$result" ]]; then
      result="$item"
    else
      result="${result}${delimiter}${item}"
    fi
  done
  printf '%s' "$result"
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Missing required command: $cmd" >&2
    exit 1
  fi
}

service_exists() {
  local service="$1"
  local item
  for item in "${ALL_SERVICES[@]}"; do
    if [[ "$item" == "$service" ]]; then
      return 0
    fi
  done
  return 1
}

append_service_unique() {
  local service="$1"
  local item
  for item in "${SELECTED_SERVICES[@]}"; do
    if [[ "$item" == "$service" ]]; then
      return 0
    fi
  done
  SELECTED_SERVICES+=("$service")
}

remove_service() {
  local service="$1"
  local filtered=()
  local item
  for item in "${SELECTED_SERVICES[@]}"; do
    if [[ "$item" != "$service" ]]; then
      filtered+=("$item")
    fi
  done
  SELECTED_SERVICES=("${filtered[@]}")
}

set_services_from_csv() {
  local csv="$1"
  local item
  local trimmed
  SELECTED_SERVICES=()

  if [[ "$csv" == "all" ]]; then
    SELECTED_SERVICES=("${ALL_SERVICES[@]}")
    return 0
  fi

  IFS=',' read -r -a items <<< "$csv"
  for item in "${items[@]}"; do
    trimmed="$(trim "$item")"
    [[ -z "$trimmed" ]] && continue
    if ! service_exists "$trimmed"; then
      echo "Unknown service: $trimmed" >&2
      exit 1
    fi
    append_service_unique "$trimmed"
  done
}

apply_skip_services() {
  local csv="$1"
  local item
  local trimmed

  [[ -z "$csv" ]] && return 0
  IFS=',' read -r -a items <<< "$csv"
  for item in "${items[@]}"; do
    trimmed="$(trim "$item")"
    [[ -z "$trimmed" ]] && continue
    if ! service_exists "$trimmed"; then
      echo "Unknown service in --skip-services: $trimmed" >&2
      exit 1
    fi
    remove_service "$trimmed"
  done
}

service_is_binary() {
  local service="$1"
  [[ "$service" != "web-admin" ]]
}

service_entrypoint() {
  case "$1" in
    admin-api) echo "./api/admin/admin.go" ;;
    front-api) echo "./api/front/front.go" ;;
    sys-rpc) echo "./rpc/sys/sys.go" ;;
    ums-rpc) echo "./rpc/ums/ums.go" ;;
    pms-rpc) echo "./rpc/pms/pms.go" ;;
    oms-rpc) echo "./rpc/oms/oms.go" ;;
    sms-rpc) echo "./rpc/sms/sms.go" ;;
    cms-rpc) echo "./rpc/cms/cms.go" ;;
    search-rpc) echo "./rpc/search/search.go" ;;
    consumer) echo "./consumer/consumer.go" ;;
    job) echo "./job/job.go" ;;
    *)
      echo "No entrypoint for service: $1" >&2
      exit 1
      ;;
  esac
}

service_config_source() {
  case "$1" in
    admin-api) echo "api/admin/etc/admin-api.yaml" ;;
    front-api) echo "api/front/etc/front-api.yaml" ;;
    sys-rpc) echo "rpc/sys/etc/sys.yaml" ;;
    ums-rpc) echo "rpc/ums/etc/ums.yaml" ;;
    pms-rpc) echo "rpc/pms/etc/pms.yaml" ;;
    oms-rpc) echo "rpc/oms/etc/oms.yaml" ;;
    sms-rpc) echo "rpc/sms/etc/sms.yaml" ;;
    cms-rpc) echo "rpc/cms/etc/cms.yaml" ;;
    search-rpc) echo "rpc/search/etc/search.yaml" ;;
    consumer) echo "consumer/etc/consumer-api.yaml" ;;
    job) echo "job/etc/job-api.yaml" ;;
    *)
      echo "No config source for service: $1" >&2
      exit 1
      ;;
  esac
}

service_config_name() {
  case "$1" in
    consumer) echo "consumer-api.yaml" ;;
    job) echo "job-api.yaml" ;;
    *) echo "$1.yaml" ;;
  esac
}

runtime_config_paths() {
  cat <<'EOF'
api/admin/etc/admin-api.yaml
api/front/etc/front-api.yaml
rpc/sys/etc/sys.yaml
rpc/ums/etc/ums.yaml
rpc/pms/etc/pms.yaml
rpc/oms/etc/oms.yaml
rpc/sms/etc/sms.yaml
rpc/cms/etc/cms.yaml
rpc/search/etc/search.yaml
consumer/etc/consumer-api.yaml
job/etc/job-api.yaml
EOF
}

service_selected() {
  local service="$1"
  local item
  for item in "${SELECTED_SERVICES[@]}"; do
    if [[ "$item" == "$service" ]]; then
      return 0
    fi
  done
  return 1
}

local_git_dirty() {
  [[ -n "$(git -C "$REPO_ROOT" status --porcelain 2>/dev/null || true)" ]]
}

resolve_effective_sync_mode() {
  local mode="$SYNC_MODE"

  if [[ "$RUN_SYNC" -eq 0 || "$mode" == "none" ]]; then
    echo "none"
    return 0
  fi

  if [[ "$mode" == "auto" ]]; then
    if local_git_dirty; then
      echo "snapshot"
    else
      echo "git"
    fi
    return 0
  fi

  if [[ "$mode" != "git" && "$mode" != "rsync" ]]; then
    echo "Unsupported sync mode: $mode" >&2
    exit 1
  fi

  echo "$mode"
}

print_local_git_status() {
  git -C "$REPO_ROOT" status --short --branch >&2 || true
}

require_clean_git_tree() {
  if local_git_dirty; then
    echo "Git deploy requires a clean local worktree. Commit/stash/clean your changes or use --sync-mode rsync explicitly." >&2
    print_local_git_status
    exit 1
  fi
}

ensure_git_branch_synced_with_upstream() {
  local branch
  local upstream
  local counts
  local behind
  local ahead

  branch="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD)"
  if [[ "$branch" == "HEAD" ]]; then
    echo "Git deploy from detached HEAD is not allowed by default. Checkout a branch with an upstream, or pass --git-ref explicitly." >&2
    exit 1
  fi

  upstream="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)"
  if [[ -z "$upstream" ]]; then
    echo "Current branch '$branch' has no upstream tracking branch. Push it first or pass --git-ref explicitly." >&2
    exit 1
  fi

  counts="$(git -C "$REPO_ROOT" rev-list --left-right --count "$upstream...HEAD")"
  behind="${counts%% *}"
  ahead="${counts##* }"
  if [[ "$ahead" != "0" || "$behind" != "0" ]]; then
    echo "Git deploy requires the local branch to match its upstream exactly." >&2
    echo "Current branch: $branch" >&2
    echo "Upstream: $upstream" >&2
    echo "Behind: $behind, Ahead: $ahead" >&2
    echo "Push/pull first, or pass --sync-mode rsync explicitly." >&2
    exit 1
  fi
}

create_snapshot_branch_ref() {
  require_cmd mktemp

  local git_remote_url="$1"
  local push_target="$2"
  local base_ref="$3"
  local ts
  local snapshot_branch
  local temp_dir
  local cleanup_worktree=0
  local branch_created=0
  local snapshot_commit
  local branch_slug
  local current_branch
  local rsync_args=(
    -a
    --delete
    --exclude=.git
    --exclude=.git/
    --exclude=target/
    --exclude=logs/
    --exclude=deploy-backup/
    --exclude=node_modules/
    --exclude=web-admin/node_modules/
    --exclude=web-admin/dist/
  )

  current_branch="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD)"
  branch_slug="${current_branch//\//-}"
  ts="$(date +%Y%m%d-%H%M%S)"
  snapshot_branch="${SNAPSHOT_BRANCH_PREFIX}/${branch_slug}/${ts}"
  temp_dir="$(mktemp -d /tmp/zero-admin-deploy-snapshot.XXXXXX)"

  cleanup_snapshot() {
    set +e
    if [[ "$cleanup_worktree" -eq 1 ]]; then
      git -C "$REPO_ROOT" worktree remove --force "$temp_dir" >/dev/null 2>&1 || true
    fi
    if [[ "$branch_created" -eq 1 ]]; then
      git -C "$REPO_ROOT" branch -D "$snapshot_branch" >/dev/null 2>&1 || true
    fi
    rm -rf "$temp_dir" >/dev/null 2>&1 || true
  }

  trap cleanup_snapshot RETURN

  git -C "$REPO_ROOT" worktree add -b "$snapshot_branch" "$temp_dir" "$base_ref" >/dev/null
  cleanup_worktree=1
  branch_created=1

  rsync "${rsync_args[@]}" "$REPO_ROOT/" "$temp_dir/"

  git -C "$temp_dir" add -A
  if git -C "$temp_dir" diff --cached --quiet; then
    echo "Auto snapshot requested but no staged changes were produced from the dirty worktree." >&2
    exit 1
  fi

  git -C "$temp_dir" commit -m "chore: deploy snapshot $ts" >/dev/null
  snapshot_commit="$(git -C "$temp_dir" rev-parse HEAD)"
  git -C "$temp_dir" push "$push_target" "HEAD:refs/heads/$snapshot_branch" >/dev/null

  echo "snapshot_branch=$snapshot_branch" >&2
  echo "snapshot_commit=$snapshot_commit" >&2
  printf '%s' "$snapshot_branch"
}

resolve_git_remote_url() {
  if [[ -n "$GIT_REMOTE_URL" ]]; then
    printf '%s' "$GIT_REMOTE_URL"
    return 0
  fi

  git -C "$REPO_ROOT" remote get-url "$GIT_REMOTE"
}

resolve_git_ref() {
  if [[ -n "$GIT_REF" ]]; then
    printf '%s' "$GIT_REF"
    return 0
  fi

  local upstream
  upstream="$(git -C "$REPO_ROOT" rev-parse --abbrev-ref --symbolic-full-name '@{u}' 2>/dev/null || true)"
  if [[ -z "$upstream" ]]; then
    echo "Unable to resolve default git ref because the current branch has no upstream. Pass --git-ref explicitly." >&2
    exit 1
  fi

  printf '%s' "${upstream#*/}"
}

rsync_supports_info_flag() {
  rsync --help 2>&1 | grep -q -- '--info'
}

sync_source_with_rsync() {
  local rsync_args=(
    -az
    --delete
    --exclude=.git/
    --exclude=.agents/
    --exclude=.claude/
    --exclude=.vscode/
    --exclude=.idea/
    --exclude=.run/
    --exclude=.windsurf/
    --exclude=.windsurfrules
    --exclude=_bmad/
    --exclude=_opcos/
    --exclude=deploy-backup/
    --exclude=target/
    --exclude=logs/
    --exclude=node_modules/
    --exclude=web-admin/node_modules/
    --exclude=web-admin/dist/
    --exclude=flutter-mall/
    --exclude=api/admin/etc/admin-api.yaml
    --exclude=api/front/etc/front-api.yaml
    --exclude=rpc/sys/etc/sys.yaml
    --exclude=rpc/ums/etc/ums.yaml
    --exclude=rpc/pms/etc/pms.yaml
    --exclude=rpc/oms/etc/oms.yaml
    --exclude=rpc/sms/etc/sms.yaml
    --exclude=rpc/cms/etc/cms.yaml
    --exclude=rpc/search/etc/search.yaml
    --exclude=consumer/etc/consumer-api.yaml
    --exclude=job/etc/job-api.yaml
  )

  if rsync_supports_info_flag; then
    rsync_args+=(--info=stats1)
  else
    rsync_args+=(--stats)
  fi

  echo "[3/7] Syncing repo source via rsync to $SSH_TARGET:$REMOTE_ROOT"
  rsync "${rsync_args[@]}" "$REPO_ROOT/" "$SSH_TARGET:$REMOTE_ROOT/"
}

sync_source_with_git() {
  local git_remote_url="$1"
  local git_ref="$2"

  echo "[3/7] Syncing repo source via git: $git_remote_url $git_ref"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "bash -s" -- "$REMOTE_ROOT" "$TS" "$git_remote_url" "$git_ref" <<'EOF'
set -euo pipefail

remote_root="$1"
ts="$2"
git_remote_url="$3"
git_ref="$4"
config_backup_dir="$remote_root/deploy-backup/$ts/source-configs"

mkdir -p "$config_backup_dir"
git config --global --add safe.directory "$remote_root" >/dev/null 2>&1 || true

cd "$remote_root"

while IFS= read -r rel; do
  [[ -z "$rel" ]] && continue
  if [[ -f "$rel" ]]; then
    mkdir -p "$config_backup_dir/$(dirname "$rel")"
    cp "$rel" "$config_backup_dir/$rel"
  fi
done <<'CFG'
api/admin/etc/admin-api.yaml
api/front/etc/front-api.yaml
rpc/sys/etc/sys.yaml
rpc/ums/etc/ums.yaml
rpc/pms/etc/pms.yaml
rpc/oms/etc/oms.yaml
rpc/sms/etc/sms.yaml
rpc/cms/etc/cms.yaml
rpc/search/etc/search.yaml
consumer/etc/consumer-api.yaml
job/etc/job-api.yaml
CFG

git fetch --depth=1 "$git_remote_url" "$git_ref"
git checkout -f FETCH_HEAD

while IFS= read -r rel; do
  [[ -z "$rel" ]] && continue
  if [[ -f "$config_backup_dir/$rel" ]]; then
    mkdir -p "$(dirname "$rel")"
    cp "$config_backup_dir/$rel" "$rel"
  fi
done <<'CFG'
api/admin/etc/admin-api.yaml
api/front/etc/front-api.yaml
rpc/sys/etc/sys.yaml
rpc/ums/etc/ums.yaml
rpc/pms/etc/pms.yaml
rpc/oms/etc/oms.yaml
rpc/sms/etc/sms.yaml
rpc/cms/etc/cms.yaml
rpc/search/etc/search.yaml
consumer/etc/consumer-api.yaml
job/etc/job-api.yaml
CFG

echo "checked_out_commit=$(git rev-parse HEAD)"
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      SSH_TARGET="$2"
      shift 2
      ;;
    --remote-root)
      REMOTE_ROOT="$2"
      shift 2
      ;;
    --services)
      SERVICES_CSV="$2"
      shift 2
      ;;
    --skip-services)
      SKIP_SERVICES_CSV="$2"
      shift 2
      ;;
    --migration)
      MIGRATION_PATH="$2"
      shift 2
      ;;
    --sync-mode)
      SYNC_MODE="$2"
      shift 2
      ;;
    --git-remote)
      GIT_REMOTE="$2"
      shift 2
      ;;
    --git-push-remote)
      GIT_PUSH_REMOTE="$2"
      shift 2
      ;;
    --git-remote-url)
      GIT_REMOTE_URL="$2"
      shift 2
      ;;
    --git-ref)
      GIT_REF="$2"
      shift 2
      ;;
    --skip-web)
      SKIP_SERVICES_CSV="${SKIP_SERVICES_CSV:+$SKIP_SERVICES_CSV,}web-admin"
      shift
      ;;
    --skip-admin-api)
      SKIP_SERVICES_CSV="${SKIP_SERVICES_CSV:+$SKIP_SERVICES_CSV,}admin-api"
      shift
      ;;
    --skip-sys-rpc)
      SKIP_SERVICES_CSV="${SKIP_SERVICES_CSV:+$SKIP_SERVICES_CSV,}sys-rpc"
      shift
      ;;
    --skip-smoke)
      RUN_SMOKE=0
      shift
      ;;
    --skip-sync)
      RUN_SYNC=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

set_services_from_csv "$SERVICES_CSV"
apply_skip_services "$SKIP_SERVICES_CSV"

if [[ "${#SELECTED_SERVICES[@]}" -eq 0 && -z "$MIGRATION_PATH" ]]; then
  echo "Nothing to deploy. Select at least one service or pass --migration." >&2
  exit 1
fi

EFFECTIVE_SYNC_MODE="$(resolve_effective_sync_mode)"
if [[ "$EFFECTIVE_SYNC_MODE" == "git" || "$EFFECTIVE_SYNC_MODE" == "snapshot" ]]; then
  require_cmd git
  RESOLVED_GIT_REMOTE_URL="$(resolve_git_remote_url)"
fi

if [[ "$EFFECTIVE_SYNC_MODE" == "git" ]]; then
  require_clean_git_tree
  if [[ -z "$GIT_REF" ]]; then
    ensure_git_branch_synced_with_upstream
  fi
  RESOLVED_GIT_REF="$(resolve_git_ref)"
fi

if [[ "$EFFECTIVE_SYNC_MODE" == "snapshot" ]]; then
  if [[ -n "$GIT_REF" ]]; then
    echo "When worktree is dirty, auto mode manages its own snapshot ref. Use --sync-mode git with a clean tree, or --sync-mode rsync." >&2
    exit 1
  fi
  RESOLVED_GIT_REF="$(create_snapshot_branch_ref "$RESOLVED_GIT_REMOTE_URL" "$GIT_PUSH_REMOTE" HEAD)"
fi

require_cmd ssh
require_cmd scp
require_cmd rsync
require_cmd python3

NEEDS_GO=0
NEEDS_NPM=0
for service in "${SELECTED_SERVICES[@]}"; do
  if service_is_binary "$service"; then
    NEEDS_GO=1
  fi
  if [[ "$service" == "web-admin" ]]; then
    NEEDS_NPM=1
  fi
done

if [[ "$EFFECTIVE_SYNC_MODE" != "git" && "$EFFECTIVE_SYNC_MODE" != "snapshot" && "$NEEDS_GO" -eq 1 ]]; then
  require_cmd go
fi

if [[ "$EFFECTIVE_SYNC_MODE" != "git" && "$EFFECTIVE_SYNC_MODE" != "snapshot" && "$NEEDS_NPM" -eq 1 ]]; then
  require_cmd npm
fi

cd "$REPO_ROOT"

if [[ -n "$MIGRATION_PATH" && "$MIGRATION_PATH" != /* ]]; then
  MIGRATION_PATH="$REPO_ROOT/$MIGRATION_PATH"
fi

if [[ -n "$MIGRATION_PATH" && ! -f "$MIGRATION_PATH" ]]; then
  echo "Migration file not found: $MIGRATION_PATH" >&2
  exit 1
fi

TS="$(date +%Y%m%d-%H%M%S)"
RELEASE_DIR="/tmp/zero-admin-remote-deploy/$TS"
mkdir -p "$RELEASE_DIR"

echo "[1/7] Verifying SSH access: $SSH_TARGET"
ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" "test -d '$REMOTE_ROOT'"

echo "[2/7] Preparing build artifacts"
if [[ "$EFFECTIVE_SYNC_MODE" == "git" || "$EFFECTIVE_SYNC_MODE" == "snapshot" ]]; then
  echo "  - git-backed mode will build artifacts from remote fetched source"
else
  for service in "${SELECTED_SERVICES[@]}"; do
    if ! service_is_binary "$service"; then
      continue
    fi
    echo "  - building $service locally"
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$RELEASE_DIR/$service" "$(service_entrypoint "$service")"
  done

  if service_selected "web-admin"; then
    if [[ ! -d "$REPO_ROOT/web-admin/node_modules" ]]; then
      (cd "$REPO_ROOT/web-admin" && npm install)
    fi
    (cd "$REPO_ROOT/web-admin" && NODE_OPTIONS=--openssl-legacy-provider npm run build)
  fi
fi

case "$EFFECTIVE_SYNC_MODE" in
  git|snapshot)
    sync_source_with_git "$RESOLVED_GIT_REMOTE_URL" "$RESOLVED_GIT_REF"
    ;;
  rsync)
    if local_git_dirty; then
      echo "  - local workspace is dirty, using rsync so uncommitted changes are included"
    fi
    sync_source_with_rsync
    ;;
  none)
    echo "[3/7] Source sync skipped"
    ;;
esac

if [[ -n "$MIGRATION_PATH" ]]; then
  echo "[4/7] Uploading and applying migration: ${MIGRATION_PATH#$REPO_ROOT/}"
  REMOTE_MIGRATION="$REMOTE_ROOT/deploy-backup/$TS/$(basename "$MIGRATION_PATH")"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "mkdir -p '$REMOTE_ROOT/deploy-backup/$TS/db' '$REMOTE_ROOT/deploy-backup/$TS/artifacts'"
  scp -o BatchMode=yes -o StrictHostKeyChecking=no "$MIGRATION_PATH" "$SSH_TARGET:$REMOTE_MIGRATION"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "mysqldump -h '$MYSQL_HOST' -P '$MYSQL_PORT' -u '$MYSQL_USER' '-p$MYSQL_PASSWORD' '$MYSQL_DB' > '$REMOTE_ROOT/deploy-backup/$TS/db/${MYSQL_DB}.pre-migration.sql'"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "mysql -h '$MYSQL_HOST' -P '$MYSQL_PORT' -u '$MYSQL_USER' '-p$MYSQL_PASSWORD' '$MYSQL_DB' < '$REMOTE_MIGRATION'"
else
  echo "[4/7] No migration requested"
fi

echo "[5/7] Uploading deployable artifacts"
ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
  "mkdir -p '$REMOTE_ROOT/deploy-backup/$TS/artifacts' '$REMOTE_ROOT/web-admin'"

if [[ "$EFFECTIVE_SYNC_MODE" == "git" || "$EFFECTIVE_SYNC_MODE" == "snapshot" ]]; then
  remote_go_arg="$REMOTE_GO_BIN"
  remote_npm_arg="$REMOTE_NPM_BIN"
  [[ -z "$remote_go_arg" ]] && remote_go_arg="$REMOTE_BIN_AUTO"
  [[ -z "$remote_npm_arg" ]] && remote_npm_arg="$REMOTE_BIN_AUTO"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "bash -s" -- "$REMOTE_ROOT" "$TS" "$(join_by , "${SELECTED_SERVICES[@]}")" "$remote_go_arg" "$remote_npm_arg" <<'EOF'
set -euo pipefail

remote_root="$1"
ts="$2"
services_csv="$3"
remote_go_bin_override="$4"
remote_npm_bin_override="$5"
artifact_dir="$remote_root/deploy-backup/$ts/artifacts"
backup_dir="$remote_root/deploy-backup/$ts"

if [[ "$remote_go_bin_override" == "__AUTO__" ]]; then
  remote_go_bin_override=""
fi

if [[ "$remote_npm_bin_override" == "__AUTO__" ]]; then
  remote_npm_bin_override=""
fi

IFS=',' read -r -a services <<< "$services_csv"

resolve_go_bin() {
  if [[ -n "$remote_go_bin_override" && -x "$remote_go_bin_override" ]]; then
    printf '%s' "$remote_go_bin_override"
    return 0
  fi
  if command -v go >/dev/null 2>&1; then
    command -v go
    return 0
  fi
  if [[ -x /usr/local/go/bin/go ]]; then
    printf '%s' /usr/local/go/bin/go
    return 0
  fi
  echo "remote go binary not found" >&2
  exit 1
}

resolve_npm_bin() {
  if [[ -n "$remote_npm_bin_override" && -x "$remote_npm_bin_override" ]]; then
    printf '%s' "$remote_npm_bin_override"
    return 0
  fi
  if command -v npm >/dev/null 2>&1; then
    command -v npm
    return 0
  fi
  echo "remote npm binary not found" >&2
  exit 1
}

service_entrypoint() {
  case "$1" in
    admin-api) echo "./api/admin/admin.go" ;;
    front-api) echo "./api/front/front.go" ;;
    sys-rpc) echo "./rpc/sys/sys.go" ;;
    ums-rpc) echo "./rpc/ums/ums.go" ;;
    pms-rpc) echo "./rpc/pms/pms.go" ;;
    oms-rpc) echo "./rpc/oms/oms.go" ;;
    sms-rpc) echo "./rpc/sms/sms.go" ;;
    cms-rpc) echo "./rpc/cms/cms.go" ;;
    search-rpc) echo "./rpc/search/search.go" ;;
    consumer) echo "./consumer/consumer.go" ;;
    job) echo "./job/job.go" ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

GO_BIN=""
NPM_BIN=""

mkdir -p "$artifact_dir" "$backup_dir"
cd "$remote_root"

if printf '%s\n' "${services[@]}" | grep -qv '^web-admin$'; then
  GO_BIN="$(resolve_go_bin)"
  export PATH="$(dirname "$GO_BIN"):$PATH"
  export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
fi

for service in "${services[@]}"; do
  [[ -z "$service" ]] && continue
  if [[ "$service" == "web-admin" ]]; then
    if [[ -z "$NPM_BIN" ]]; then
      NPM_BIN="$(resolve_npm_bin)"
    fi
    build_root="$backup_dir/web-admin-build"
    rm -rf "$build_root"
    mkdir -p "$build_root"
    cp -a "$remote_root/web-admin/." "$build_root/"
    (
      cd "$build_root"
      if [[ ! -d node_modules ]]; then
        "$NPM_BIN" install
      fi
      NODE_OPTIONS=--openssl-legacy-provider "$NPM_BIN" run build
    )
    rm -rf "$remote_root/web-admin/dist.new"
    cp -a "$build_root/dist" "$remote_root/web-admin/dist.new"
    continue
  fi

  echo "  - building $service on remote host"
  "$GO_BIN" build -o "$artifact_dir/$service" "$(service_entrypoint "$service")"
done
EOF
else
  artifact_files=()
  for service in "${SELECTED_SERVICES[@]}"; do
    if service_is_binary "$service"; then
      artifact_files+=("$RELEASE_DIR/$service")
    fi
  done

  if [[ "${#artifact_files[@]}" -gt 0 ]]; then
    scp -o BatchMode=yes -o StrictHostKeyChecking=no "${artifact_files[@]}" \
      "$SSH_TARGET:$REMOTE_ROOT/deploy-backup/$TS/artifacts/"
  fi

  if service_selected "web-admin"; then
    rsync -az --delete "$REPO_ROOT/web-admin/dist/" "$SSH_TARGET:$REMOTE_ROOT/web-admin/dist.new/"
  fi
fi

echo "[6/7] Backing up and swapping remote services"
SERVICE_LIST="$(join_by , "${SELECTED_SERVICES[@]}")"
ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
  "bash -s" -- "$REMOTE_ROOT" "$TS" "$SERVICE_LIST" <<'EOF'
set -euo pipefail

remote_root="$1"
ts="$2"
services_csv="$3"
backup_dir="$remote_root/deploy-backup/$ts"
artifact_dir="$backup_dir/artifacts"
target_root="$remote_root/target"

IFS=',' read -r -a services <<< "$services_csv"

service_config_source() {
  case "$1" in
    admin-api) echo "api/admin/etc/admin-api.yaml" ;;
    front-api) echo "api/front/etc/front-api.yaml" ;;
    sys-rpc) echo "rpc/sys/etc/sys.yaml" ;;
    ums-rpc) echo "rpc/ums/etc/ums.yaml" ;;
    pms-rpc) echo "rpc/pms/etc/pms.yaml" ;;
    oms-rpc) echo "rpc/oms/etc/oms.yaml" ;;
    sms-rpc) echo "rpc/sms/etc/sms.yaml" ;;
    cms-rpc) echo "rpc/cms/etc/cms.yaml" ;;
    search-rpc) echo "rpc/search/etc/search.yaml" ;;
    consumer) echo "consumer/etc/consumer-api.yaml" ;;
    job) echo "job/etc/job-api.yaml" ;;
    *)
      echo "unknown service: $1" >&2
      exit 1
      ;;
  esac
}

service_config_name() {
  case "$1" in
    consumer) echo "consumer-api.yaml" ;;
    job) echo "job-api.yaml" ;;
    *) echo "$1.yaml" ;;
  esac
}

service_legacy_pattern() {
  case "$1" in
    admin-api|sys-rpc)
      local config_name
      config_name="$(service_config_name "$1")"
      echo "./$1 -f ./$config_name"
      ;;
    *)
      echo ""
      ;;
  esac
}

kill_if_running() {
  local pattern="$1"
  [[ -z "$pattern" ]] && return 0
  if pgrep -f "$pattern" >/dev/null 2>&1; then
    pkill -f "$pattern"
    sleep 1
  fi
}

deploy_binary_service() {
  local service="$1"
  local binary="$service"
  local config_name
  local config_source_rel
  local config_source_abs
  local service_target_dir
  local new_pattern
  local legacy_pattern

  config_name="$(service_config_name "$service")"
  config_source_rel="$(service_config_source "$service")"
  config_source_abs="$remote_root/$config_source_rel"
  service_target_dir="$target_root/$service"
  new_pattern="./$service/$binary -f ./$service/$config_name"
  legacy_pattern="$(service_legacy_pattern "$service")"

  mkdir -p "$backup_dir/$service" "$service_target_dir"

  if [[ -f "$service_target_dir/$binary" ]]; then
    cp "$service_target_dir/$binary" "$backup_dir/$service/$binary"
  fi

  if [[ -f "$config_source_abs" && ! -f "$service_target_dir/$config_name" ]]; then
    cp "$config_source_abs" "$service_target_dir/$config_name"
  fi

  kill_if_running "$legacy_pattern"
  kill_if_running "$new_pattern"

  install -m 0755 "$artifact_dir/$binary" "$service_target_dir/$binary"

  (
    cd "$target_root"
    nohup "./$service/$binary" -f "./$service/$config_name" >/dev/null 2>&1 &
  )

  sleep 2
  pgrep -af "$new_pattern" >/dev/null || {
    echo "failed to restart $service" >&2
    exit 1
  }
}

deploy_web() {
  mkdir -p "$backup_dir/web-admin"
  if [[ -d "$remote_root/web-admin/dist" ]]; then
    cp -a "$remote_root/web-admin/dist" "$backup_dir/web-admin/dist"
    rm -rf "$remote_root/web-admin/dist"
  fi
  mv "$remote_root/web-admin/dist.new" "$remote_root/web-admin/dist"
}

mkdir -p "$backup_dir"

for service in "${services[@]}"; do
  [[ -z "$service" ]] && continue
  if [[ "$service" == "web-admin" ]]; then
    deploy_web
  else
    deploy_binary_service "$service"
  fi
done

echo "backup_dir=$backup_dir"
ps -ef | egrep '(admin-api|front-api|sys-rpc|ums-rpc|pms-rpc|oms-rpc|sms-rpc|cms-rpc|search-rpc|consumer|job)' | grep -v grep || true
EOF

if [[ "$RUN_SMOKE" -eq 1 ]]; then
  echo "[7/7] Running smoke checks"
  SMOKE_ARGS=(
    --base-url "$SMOKE_BASE_URL"
    --account "$ADMIN_ACCOUNT"
    --password "$ADMIN_PASSWORD"
  )
  if service_selected "front-api"; then
    SMOKE_ARGS+=(--front-base-url "$FRONT_SMOKE_BASE_URL")
  fi

  python3 "$SCRIPT_DIR/smoke_remote.py" "${SMOKE_ARGS[@]}"
else
  echo "[7/7] Smoke checks skipped"
fi

echo "Deployment complete."
echo "Services: $(join_by , "${SELECTED_SERVICES[@]}")"
echo "Sync mode: $EFFECTIVE_SYNC_MODE"
if [[ "$EFFECTIVE_SYNC_MODE" == "git" || "$EFFECTIVE_SYNC_MODE" == "snapshot" ]]; then
  echo "Git source: $RESOLVED_GIT_REMOTE_URL @ $RESOLVED_GIT_REF"
fi
echo "Release dir: $RELEASE_DIR"
echo "Remote backup: $REMOTE_ROOT/deploy-backup/$TS"
