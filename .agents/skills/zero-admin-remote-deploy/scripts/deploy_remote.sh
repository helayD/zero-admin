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

GIT_REMOTE="${ZERO_ADMIN_GIT_REMOTE:-origin}"
GIT_REMOTE_URL="${ZERO_ADMIN_GIT_REMOTE_URL:-}"
GIT_PUSH_REMOTE="${ZERO_ADMIN_GIT_PUSH_REMOTE:-}"
GIT_REF="${ZERO_ADMIN_GIT_REF:-}"
AUTO_COMMIT_MESSAGE_PREFIX="${ZERO_ADMIN_AUTO_COMMIT_MESSAGE_PREFIX:-chore: deploy sync}"
REMOTE_GO_BIN="${ZERO_ADMIN_REMOTE_GO_BIN:-}"
REMOTE_NPM_BIN="${ZERO_ADMIN_REMOTE_NPM_BIN:-}"
REMOTE_BIN_AUTO="__AUTO__"
AUTO_COMMIT_EXCLUDES=(
  target
  deploy-backup
  logs
  node_modules
  web-admin/node_modules
  web-admin/dist
  flutter-mall
  admin
  cms
  oms
  pms
  search
  sms
  sys
)

DEFAULT_SERVICES="all"
SERVICES_CSV="${ZERO_ADMIN_DEPLOY_SERVICES:-$DEFAULT_SERVICES}"
SKIP_SERVICES_CSV=""
RUN_SMOKE=1
RUN_API_TEST=1
AUTO_MIGRATION=0
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
  --auto-migration             Auto-discover and apply all script/sql/migration_*.sql files
  --skip-api-test              Skip post-deploy Story API tests
  --api-test                   Force run Story API tests (default: on)
  --stories <csv>              Only run specified story tests, e.g. "4-5,4-6,5-1"
  --git-remote <name>          Local git remote used to resolve the fetch URL. Default: origin
  --git-push-remote <target>   Local git push target used before deploy. Default: same as --git-remote
  --git-remote-url <url>       Explicit git URL to fetch on the remote host
  --git-ref <ref>              Deploy an already-pushed ref instead of publishing the current branch. Requires a clean local worktree
  --skip-web                   Backward-compatible shortcut for --skip-services web-admin
  --skip-admin-api             Backward-compatible shortcut for --skip-services admin-api
  --skip-sys-rpc               Backward-compatible shortcut for --skip-services sys-rpc
  --skip-smoke                 Skip post-deploy smoke checks
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
  for item in "${SELECTED_SERVICES[@]-}"; do
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
  for item in "${SELECTED_SERVICES[@]-}"; do
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

print_local_git_status() {
  git -C "$REPO_ROOT" status --short --branch >&2 || true
}

require_clean_git_tree() {
  if local_git_dirty; then
    echo "Deploying an explicit --git-ref requires a clean local worktree. Omit --git-ref if you want the script to commit and push the current branch for you." >&2
    print_local_git_status
    exit 1
  fi
}

current_branch_name() {
  git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD
}

ensure_not_detached_head() {
  local branch
  branch="$(current_branch_name)"
  if [[ "$branch" == "HEAD" ]]; then
    echo "Detached HEAD is not supported for automated deploy. Checkout a branch first, or pass --git-ref explicitly." >&2
    exit 1
  fi
}

rebase_current_branch_onto_push_target() {
  local branch="$1"
  local remote="$2"
  local remote_ref="$remote/$branch"
  local counts
  local behind

  git -C "$REPO_ROOT" fetch "$remote" "$branch" >/dev/null 2>&1 || true

  if ! git -C "$REPO_ROOT" show-ref --verify --quiet "refs/remotes/$remote/$branch"; then
    return 0
  fi

  counts="$(git -C "$REPO_ROOT" rev-list --left-right --count "$remote_ref...HEAD")"
  behind="${counts%% *}"
  if [[ "$behind" != "0" ]]; then
    echo "Rebasing current branch '$branch' onto $remote_ref before deploy" >&2
    git -C "$REPO_ROOT" rebase "$remote_ref" >/dev/null
  fi
}

stash_worktree_for_publish() {
  local ts="$1"
  local before_ref
  local after_ref

  if ! local_git_dirty; then
    return 0
  fi

  before_ref="$(git -C "$REPO_ROOT" rev-parse -q --verify refs/stash 2>/dev/null || true)"
  git -C "$REPO_ROOT" stash push --include-untracked -m "zero-admin deploy temp $ts" >/dev/null
  after_ref="$(git -C "$REPO_ROOT" rev-parse -q --verify refs/stash 2>/dev/null || true)"

  if [[ -n "$after_ref" && "$after_ref" != "$before_ref" ]]; then
    printf '%s' "$after_ref"
  fi
}

restore_stashed_worktree_for_publish() {
  local stash_ref="$1"
  local top_ref

  [[ -z "$stash_ref" ]] && return 0

  top_ref="$(git -C "$REPO_ROOT" rev-parse -q --verify refs/stash 2>/dev/null || true)"
  if [[ "$top_ref" != "$stash_ref" ]]; then
    echo "Skipping automatic stash restore because the stash stack changed while publishing deploy source." >&2
    return 0
  fi

  if ! git -C "$REPO_ROOT" stash pop >/dev/null; then
    echo "Failed to restore local changes after publishing deploy source. Resolve it with: git -C '$REPO_ROOT' stash pop" >&2
    exit 1
  fi
}

publish_current_branch_for_deploy() {
  local ts="$1"
  local allow_auto_commit="$2"
  local branch
  local remote="$GIT_PUSH_REMOTE"
  local commit_message
  local commit_sha
  local exclude
  local publish_stash_ref=""

  ensure_not_detached_head
  branch="$(current_branch_name)"

  if local_git_dirty; then
    if [[ "$allow_auto_commit" != "1" ]]; then
      require_clean_git_tree
    fi

    commit_message="$AUTO_COMMIT_MESSAGE_PREFIX $ts"
    git -C "$REPO_ROOT" add -A
    for exclude in "${AUTO_COMMIT_EXCLUDES[@]}"; do
      git -C "$REPO_ROOT" reset -q HEAD -- "$exclude" >/dev/null 2>&1 || true
    done
    if git -C "$REPO_ROOT" diff --cached --quiet; then
      echo "Local worktree is dirty, but only in excluded build/runtime artifacts. Nothing new will be committed before deploy." >&2
    else
      git -C "$REPO_ROOT" commit -m "$commit_message" >/dev/null
      commit_sha="$(git -C "$REPO_ROOT" rev-parse HEAD)"
      echo "auto_commit=$commit_sha" >&2
    fi
  fi

  publish_stash_ref="$(stash_worktree_for_publish "$ts")"
  trap 'restore_stashed_worktree_for_publish "$publish_stash_ref"' RETURN

  rebase_current_branch_onto_push_target "$branch" "$remote"
  git -C "$REPO_ROOT" push "$remote" "HEAD:refs/heads/$branch" >/dev/null
  echo "published_branch=$branch" >&2
  echo "published_commit=$(git -C "$REPO_ROOT" rev-parse HEAD)" >&2

  trap - RETURN
  restore_stashed_worktree_for_publish "$publish_stash_ref"
  printf '%s' "$branch"
}

resolve_named_remote_url() {
  local remote_name="$1"
  git -C "$REPO_ROOT" remote get-url "$remote_name"
}

resolve_git_ref() {
  if [[ -z "$GIT_REF" ]]; then
    echo "No git ref was provided. Pass --git-ref explicitly." >&2
    exit 1
  fi

  printf '%s' "$GIT_REF"
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

for _attempt in 1 2 3; do
  if git -c http.version=HTTP/1.1 fetch --depth=1 "$git_remote_url" "$git_ref"; then
    break
  fi
  echo "  git fetch attempt $_attempt failed, retrying in 3s..."
  sleep 3
done
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
    --api-test)
      RUN_API_TEST=1
      shift
      ;;
    --skip-api-test)
      RUN_API_TEST=0
      shift
      ;;
    --auto-migration)
      AUTO_MIGRATION=1
      shift
      ;;
    --stories)
      API_TEST_STORIES="$2"
      shift 2
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

if [[ -z "$GIT_PUSH_REMOTE" ]]; then
  GIT_PUSH_REMOTE="$GIT_REMOTE"
fi

if [[ "${#SELECTED_SERVICES[@]}" -eq 0 && -z "$MIGRATION_PATH" ]]; then
  echo "Nothing to deploy. Select at least one service or pass --migration." >&2
  exit 1
fi

TS="$(date +%Y%m%d-%H%M%S)"
require_cmd git
require_cmd ssh
require_cmd scp
require_cmd python3

if [[ -n "$GIT_REMOTE_URL" ]]; then
  RESOLVED_GIT_REMOTE_URL="$GIT_REMOTE_URL"
else
  RESOLVED_GIT_REMOTE_URL="$(resolve_named_remote_url "$GIT_PUSH_REMOTE")"
fi

if [[ -n "$GIT_REF" ]]; then
  require_clean_git_tree
  RESOLVED_GIT_REF="$(resolve_git_ref)"
else
  RESOLVED_GIT_REF="$(publish_current_branch_for_deploy "$TS" 1)"
fi

cd "$REPO_ROOT"

# Collect migration files
MIGRATION_FILES=()

if [[ -n "$MIGRATION_PATH" ]]; then
  if [[ "$MIGRATION_PATH" != /* ]]; then
    MIGRATION_PATH="$REPO_ROOT/$MIGRATION_PATH"
  fi
  if [[ ! -f "$MIGRATION_PATH" ]]; then
    echo "Migration file not found: $MIGRATION_PATH" >&2
    exit 1
  fi
  MIGRATION_FILES+=("$MIGRATION_PATH")
fi

if [[ "$AUTO_MIGRATION" -eq 1 ]]; then
  MIGRATION_SCAN_DIR="$REPO_ROOT/script/sql"
  if [[ -d "$MIGRATION_SCAN_DIR" ]]; then
    while IFS= read -r f; do
      # Deduplicate against explicitly provided --migration
      already=0
      for existing in "${MIGRATION_FILES[@]-}"; do
        [[ "$existing" == "$f" ]] && already=1 && break
      done
      [[ "$already" -eq 0 ]] && MIGRATION_FILES+=("$f")
    done < <(find "$MIGRATION_SCAN_DIR" -maxdepth 1 -name 'migration_*.sql' -type f | sort)
  fi
fi

echo "[1/8] Verifying SSH access: $SSH_TARGET"
ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" "test -d '$REMOTE_ROOT'"

echo "[2/8] Preparing build artifacts"
echo "  - remote host will build artifacts from git-synced source"

sync_source_with_git "$RESOLVED_GIT_REMOTE_URL" "$RESOLVED_GIT_REF"

if [[ ${#MIGRATION_FILES[@]} -gt 0 ]]; then
  echo "[4/8] Uploading and applying ${#MIGRATION_FILES[@]} migration(s)"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "mkdir -p '$REMOTE_ROOT/deploy-backup/$TS/db' '$REMOTE_ROOT/deploy-backup/$TS/artifacts'"

  # Backup DB before any migration
  echo "  - backing up database before migration"
  ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
    "mysqldump -h '$MYSQL_HOST' -P '$MYSQL_PORT' -u '$MYSQL_USER' '-p$MYSQL_PASSWORD' '$MYSQL_DB' > '$REMOTE_ROOT/deploy-backup/$TS/db/${MYSQL_DB}.pre-migration.sql' 2>/dev/null"

  for mig_file in "${MIGRATION_FILES[@]}"; do
    mig_basename="$(basename "$mig_file")"
    REMOTE_MIGRATION="$REMOTE_ROOT/deploy-backup/$TS/$mig_basename"
    echo "  - applying: $mig_basename"
    scp -o BatchMode=yes -o StrictHostKeyChecking=no "$mig_file" "$SSH_TARGET:$REMOTE_MIGRATION"
    ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
      "mysql -h '$MYSQL_HOST' -P '$MYSQL_PORT' -u '$MYSQL_USER' '-p$MYSQL_PASSWORD' '$MYSQL_DB' < '$REMOTE_MIGRATION' 2>/dev/null"
  done
  echo "  - all migrations applied successfully"
else
  echo "[4/8] No migration requested"
fi

echo "[5/8] Uploading deployable artifacts"
ssh -o BatchMode=yes -o StrictHostKeyChecking=no "$SSH_TARGET" \
  "mkdir -p '$REMOTE_ROOT/deploy-backup/$TS/artifacts' '$REMOTE_ROOT/web-admin'"

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

echo "[6/8] Backing up and swapping remote services"
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

  # ---- FIX 1: Sync missing RPC client entries from source YAML to target YAML ----
  sync_missing_rpc_clients "$service" "$config_source_abs" "$service_target_dir/$config_name"

  # ---- FIX 2: Warn on port mismatch between deployed API client config and RPC ListenOn ----
  check_rpc_port_mismatch "$service" "$service_target_dir/$config_name"

  kill_if_running "$legacy_pattern"
  kill_if_running "$new_pattern"

  install -m 0755 "$artifact_dir/$binary" "$service_target_dir/$binary"

  # ---- FIX 3: Restart fallback — try nohup if pgrep check fails ----
  (
    cd "$target_root"
    nohup "./$service/$binary" -f "./$service/$config_name" >/dev/null 2>&1 &
  )

  sleep 3
  if pgrep -af "$new_pattern" >/dev/null 2>&1; then
    echo "  $service started successfully"
  else
    echo "  pgrep check failed, trying direct process check..."
    if pgrep -f "$binary" >/dev/null 2>&1; then
      echo "  $service is running (process found)"
    else
      echo "  WARNING: $service may not have started cleanly. Check logs manually." >&2
    fi
  fi
}

# Extract ListenOn port from an RPC server config YAML
get_rpc_server_port() {
  local rpc_config="$1"
  grep -i "^ListenOn:" "$rpc_config" 2>/dev/null | awk '{print $2}' | cut -d: -f2
}

extract_yaml_top_level_block() {
  local key="$1"
  local yaml_file="$2"

  awk -v key="$key" '
    $0 ~ "^" key ":" {
      capture = 1
    }
    capture && $0 ~ /^[^[:space:]#][^:]*:/ && $0 !~ "^" key ":" {
      exit
    }
    capture {
      print
    }
  ' "$yaml_file" 2>/dev/null || true
}

rpc_client_endpoint_port() {
  local client_name="$1"
  local yaml_file="$2"
  local block

  block="$(extract_yaml_top_level_block "$client_name" "$yaml_file")"
  [[ -z "$block" ]] && return 0

  if ! printf '%s\n' "$block" | grep -q '^[[:space:]]*Endpoints:'; then
    return 0
  fi

  printf '%s\n' "$block" | awk '
    /^[[:space:]]*Endpoints:/ {
      in_endpoints = 1
      next
    }
    in_endpoints && /^[[:space:]]*-/ {
      line = $0
      sub(/.*:/, "", line)
      gsub(/[^0-9].*$/, "", line)
      if (line != "") {
        print line
        exit
      }
    }
    in_endpoints && /^[^[:space:]-]/ {
      exit
    }
  ' 2>/dev/null || true
}

rpc_client_uses_etcd() {
  local client_name="$1"
  local yaml_file="$2"
  local block

  block="$(extract_yaml_top_level_block "$client_name" "$yaml_file")"
  [[ -z "$block" ]] && return 1

  printf '%s\n' "$block" | grep -q '^[[:space:]]*Etcd:'
}

# Get the RPC client name (e.g. SearchRpc) from the api YAML given the server name (e.g. search-rpc)
rpc_client_name() {
  local server="$1"
  case "$server" in
    sys-rpc)     echo "SysRpc" ;;
    ums-rpc)     echo "UmsRpc" ;;
    pms-rpc)     echo "PmsRpc" ;;
    oms-rpc)     echo "OmsRpc" ;;
    sms-rpc)     echo "SmsRpc" ;;
    cms-rpc)     echo "CmsRpc" ;;
    search-rpc)  echo "SearchRpc" ;;
    *)           echo "" ;;
  esac
}

# FIX 1: If source YAML has RPC client entries that target YAML is missing, copy them over.
# This handles the case where a new RPC client was added to ServiceContext but the remote
# target YAML was never updated.
sync_missing_rpc_clients() {
  local service="$1"
  local src_yaml="$2"
  local tgt_yaml="$3"

  if [[ ! -f "$src_yaml" || ! -f "$tgt_yaml" ]]; then
    return 0
  fi

  # Only API gateways have RPC client sections to sync
  case "$service" in
    admin-api|front-api) ;;
    *) return 0 ;;
  esac

  # For each known RPC server, check if the client block exists in source but not in target
  local client_name server_name tgt_port src_port
  for server_name in sys-rpc ums-rpc pms-rpc oms-rpc sms-rpc cms-rpc search-rpc; do
    client_name="$(rpc_client_name "$server_name")"
    [[ -z "$client_name" ]] && continue

    # Does source have this client block?
    if ! grep -q "^${client_name}:" "$src_yaml" 2>/dev/null; then
      continue
    fi

    # Does target have it?
    if grep -q "^${client_name}:" "$tgt_yaml" 2>/dev/null; then
      continue
    fi

    echo "  [WARN] $service target YAML missing ${client_name}, copying from source"

    # Extract the client block from source (indented block under client name)
    local block_start block_end line
    block_start=$(grep -n "^${client_name}:" "$src_yaml" 2>/dev/null | head -1 | cut -d: -f1)
    if [[ -z "$block_start" ]]; then
      continue
    fi

    # Find the end of the block (next top-level key or EOF)
    local total_lines
    total_lines=$(wc -l < "$src_yaml")
    block_end="$total_lines"
    local search_line=$((block_start + 1))
    local indent
    indent=$(sed -n "${block_start}s/^\([[:space:]]*\).*/\1/p" "$src_yaml")
    indent="${#indent}"

    while [[ $search_line -le $total_lines ]]; do
      local line_text
      line_text=$(sed -n "${search_line}p" "$src_yaml")
      if [[ -n "$line_text" ]]; then
        local line_indent
        line_indent=$(echo "$line_text" | sed 's/^\([[:space:]]*\).*/\1/')
        line_indent="${#line_indent}"
        # Top-level key found (indent == 0 or same as first line's indent and next top-level key)
        if [[ $line_indent -eq 0 && "$line_text" =~ ^[a-zA-Z] ]]; then
          block_end=$((search_line - 1))
          break
        fi
      fi
      search_line=$((search_line + 1))
    done

    # Append the block to target YAML
    sed -n "${block_start},${block_end}p" "$src_yaml" >> "$tgt_yaml"
    echo "  [SYNCED] ${client_name} added to $service target YAML"
  done
}

# FIX 2: For API gateways, check that deployed RPC client endpoints match
# the actual ListenOn ports of the corresponding RPC servers.
# Endpoints mode can be checked directly; Etcd mode is skipped on purpose.
check_rpc_port_mismatch() {
  local service="$1"
  local deployed_yaml="$2"

  if [[ ! -f "$deployed_yaml" ]]; then
    return 0
  fi

  case "$service" in
    admin-api|front-api) ;;
    *) return 0 ;;
  esac

  local client_name server_name server_yaml actual_port expected_port
  for server_name in sys-rpc ums-rpc pms-rpc oms-rpc sms-rpc cms-rpc search-rpc; do
    client_name="$(rpc_client_name "$server_name")"
    [[ -z "$client_name" ]] && continue

    # Only Endpoints mode can be directly compared to ListenOn.
    expected_port="$(rpc_client_endpoint_port "$client_name" "$deployed_yaml")"
    if [[ -z "$expected_port" ]]; then
      if rpc_client_uses_etcd "$client_name" "$deployed_yaml"; then
        echo "  [INFO] ${service} ${client_name} uses Etcd discovery, skipping direct port mismatch check"
      fi
      continue
    fi

    # Get actual ListenOn from RPC server config
    server_yaml="$remote_root/rpc/${server_name%-*}/etc/${server_name}.yaml"
    if [[ "$server_name" == "search-rpc" ]]; then
      server_yaml="$remote_root/rpc/search/etc/search.yaml"
    fi
    actual_port=$(get_rpc_server_port "$server_yaml")

    if [[ -n "$actual_port" && "$actual_port" != "$expected_port" ]]; then
      echo "  [WARN] Port mismatch: ${service} ${client_name} → ${expected_port}, but ${server_name} listens on ${actual_port}"
    fi
  done
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
  echo "[7/8] Running smoke checks"
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
  echo "[7/8] Smoke checks skipped"
fi

if [[ "$RUN_API_TEST" -eq 1 ]]; then
  echo "[8/8] Running Story API tests (100% pass required)"
  API_TEST_ARGS=(
    --admin-url "$SMOKE_BASE_URL"
    --front-url "$FRONT_SMOKE_BASE_URL"
  )
  if [[ -n "${API_TEST_STORIES:-}" ]]; then
    API_TEST_ARGS+=(--stories "$API_TEST_STORIES")
  fi

  if ! bash "$SCRIPT_DIR/run_api_tests.sh" "${API_TEST_ARGS[@]}"; then
    echo "Story API tests FAILED. Deployment verification incomplete." >&2
    exit 1
  fi
else
  echo "[8/8] Story API tests skipped"
fi

echo "Deployment complete."
echo "Services: $(join_by , "${SELECTED_SERVICES[@]}")"
echo "Deploy source: git"
echo "Git source: $RESOLVED_GIT_REMOTE_URL @ $RESOLVED_GIT_REF"
echo "Remote backup: $REMOTE_ROOT/deploy-backup/$TS"
