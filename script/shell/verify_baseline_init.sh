#!/bin/bash

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

fail() {
  echo "baseline-init check failed: $1" >&2
  exit 1
}

assert_file_contains() {
  local file="$1"
  local pattern="$2"
  local description="$3"

  if ! grep -Fq "$pattern" "$file"; then
    fail "$description"
  fi
}

assert_file_not_contains() {
  local file="$1"
  local pattern="$2"
  local description="$3"

  if grep -Fq "$pattern" "$file"; then
    fail "$description"
  fi
}

assert_path_exists() {
  local path="$1"
  local description="$2"

  if [ ! -e "$path" ]; then
    fail "$description"
  fi
}

required_paths=(
  "$REPO_ROOT/api"
  "$REPO_ROOT/rpc"
  "$REPO_ROOT/consumer"
  "$REPO_ROOT/job"
  "$REPO_ROOT/web-admin"
  "$REPO_ROOT/flutter-mall"
  "$REPO_ROOT/pkg"
  "$REPO_ROOT/docs"
  "$REPO_ROOT/script"
  "$REPO_ROOT/web-admin/package-lock.json"
  "$REPO_ROOT/flutter-mall/pubspec.lock"
)

for path in "${required_paths[@]}"; do
  assert_path_exists "$path" "required baseline path missing: ${path#$REPO_ROOT/}"
done

assert_file_not_contains \
  "$REPO_ROOT/makefile" \
  "./target/web-api/web-api" \
  "make start still references nonexistent web-api target"

assert_file_not_contains \
  "$REPO_ROOT/service_manager.sh" \
  "./target/web-api/web-api" \
  "service_manager start still references nonexistent web-api target"

if ! awk '
  $0 ~ /"pms"\)/ {
    getline
    if ($0 ~ /pms_service/) {
      found = 1
    }
  }
  END {
    exit found ? 0 : 1
  }
' "$REPO_ROOT/service_manager.sh"; then
  fail "service_manager pms branch does not dispatch to pms_service"
fi

assert_file_contains \
  "$REPO_ROOT/makefile" \
  '$(GOBUILD) -o target/pms-rpc/pms-rpc -v ./rpc/pms/pms.go' \
  "make build no longer compiles the existing pms service target"

assert_file_not_contains \
  "$REPO_ROOT/makefile" \
  "goctl@latest" \
  "makefile still installs goctl@latest and can drift generated artifacts"

assert_file_contains \
  "$REPO_ROOT/makefile" \
  "GOCTL_VERSION=v1.9.2" \
  "makefile does not pin goctl to the repository generation baseline"

echo "baseline-init checks passed"
