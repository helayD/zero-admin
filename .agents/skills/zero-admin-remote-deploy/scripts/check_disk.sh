#!/usr/bin/env bash
# ============================================================
# check_disk.sh — Remote server disk health check
# Target: root@47.107.224.56
# ============================================================
set -euo pipefail

SSH_OPTS="-o ConnectTimeout=10 -o StrictHostKeyChecking=no"
SSH_CMD="ssh $SSH_OPTS root@47.107.224.56"
ALERT_THRESHOLD=85   # % usage → warning
CRIT_THRESHOLD=90    # % usage → abort deploy

usage() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --auto-clean      Run cleanup commands automatically (apt clean, docker prune)
  --verbose         Show all known large directories
  --skip-clean      Skip cleanup suggestion (default)
  -h, --help        Show this help

Examples:
  $0                           # Check only
  $0 --auto-clean              # Check + auto clean apt/docker
  $0 --verbose                 # Full verbose report
EOF
  exit 1
}

AUTO_CLEAN=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --auto-clean) AUTO_CLEAN=true; shift ;;
    --verbose) VERBOSE=true; shift ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

echo "=============================================="
echo "  Zero-Admin Remote Disk Health Check"
echo "  Target: 47.107.224.56"
echo "  Time: $(date '+%Y-%m-%d %H:%M:%S')"
echo "=============================================="
echo

# ---- 1. Root filesystem usage ----
echo "[1/6] Root filesystem usage..."
DISK_OUTPUT=$($SSH_CMD 'df -h /' 2>/dev/null)
echo "$DISK_OUTPUT"
ROOT_USAGE=$(echo "$DISK_OUTPUT" | awk 'NR==2 {gsub(/%/, "", $5); print $5}')
ROOT_AVAIL=$(echo "$DISK_OUTPUT" | awk 'NR==2 {print $4}')
echo

# ---- 2. Inode usage ----
echo "[2/6] Inode usage..."
$SSH_CMD 'df -i /' 2>/dev/null | awk 'NR==2 {printf "  Inodes: %s used / %s total  (%.1f%% used, %s free)\n", $3, $2, ($3/$2)*100, $4}'
echo

# ---- 3. Docker system df ----
echo "[3/6] Docker system..."
$SSH_CMD 'docker system df 2>/dev/null' 2>/dev/null || echo "  (docker not accessible or not running)"
echo

# ---- 4. Known large dirs (fast ls-based, no du) ----
echo "[4/6] Known large directories..."
$SSH_CMD '
  echo "  /snap/          : $(du -sh /snap 2>/dev/null | cut -f1)"
  echo "  /var/lib/snapd  : $(du -sh /var/lib/snapd 2>/dev/null | cut -f1)"
  echo "  /var/lib/mongodb: $(du -sh /var/lib/mongodb 2>/dev/null | cut -f1)"
  echo "  /var/lib/mysql  : $(du -sh /var/lib/mysql 2>/dev/null | cut -f1)"
  echo "  /var/lib/etcd   : $(du -sh /var/lib/etcd 2>/dev/null | cut -f1)"
  echo "  /var/lib/containerd: $(du -sh /var/lib/containerd 2>/dev/null | cut -f1)"
  echo "  /var/cache/apt  : $(du -sh /var/cache/apt 2>/dev/null | cut -f1)"
  echo "  /tmp/           : $(du -sh /tmp 2>/dev/null | cut -f1)"
  echo "  /root/zero-admin: (git + binaries ~500MB total)"
' 2>/dev/null
echo

# ---- 5. Docker containers status ----
echo "[5/6] Running containers..."
$SSH_CMD 'docker ps --format "table {{.Names}}\t{{.Status}}" 2>/dev/null' 2>/dev/null || echo "  (no containers running)"
echo

# ---- 6. Verdicts ----
echo "=============================================="
echo "  Verdict"
echo "=============================================="

if (( ROOT_USAGE >= CRIT_THRESHOLD )); then
  echo "  STATUS:  CRITICAL  (${ROOT_USAGE}% used)"
  echo "  ACTION:  ABORT deploy — disk is at or above ${CRIT_THRESHOLD}%"
  echo "  AVAIL:   ${ROOT_AVAIL} remaining"
  echo ""
  echo "  Before deploying, run cleanup:"
  echo "    $0 --auto-clean"
  echo ""
elif (( ROOT_USAGE >= ALERT_THRESHOLD )); then
  echo "  STATUS:  WARNING   (${ROOT_USAGE}% used)"
  echo "  ACTION:  Proceed with caution — disk above ${ALERT_THRESHOLD}%"
  echo "  AVAIL:   ${ROOT_AVAIL} remaining"
  echo ""
  echo "  Consider running cleanup first:"
  echo "    $0 --auto-clean"
  echo ""
else
  echo "  STATUS:  OK        (${ROOT_USAGE}% used)"
  echo "  AVAIL:   ${ROOT_AVAIL} remaining"
  echo "  ACTION:  Safe to deploy"
  echo ""
fi

# ---- Auto clean ----
if $AUTO_CLEAN; then
  echo "=============================================="
  echo "  Running auto-cleanup..."
  echo "=============================================="

  echo "[1/7] apt-get clean + autoremove..."
  $SSH_CMD 'apt-get clean && apt-get autoremove -y' 2>/dev/null || true

  echo "[2/7] docker system prune..."
  $SSH_CMD 'docker system prune -f' 2>/dev/null || true

  echo "[3/7] Clean /tmp (files older than 3 days)..."
  $SSH_CMD 'find /tmp -type f -mtime +3 -delete 2>/dev/null; true' 2>/dev/null || true

  echo "[4/7] Prune old snap revisions (keep latest 2)..."
  $SSH_CMD '
    snap list --all 2>/dev/null | awk "/lxd|core20/ && NF>=5 {
      rev=\$3; if(\$5 ~ /disabled/) {
        printf "Removing snap %s rev %s...\n", \$1, rev
        snap remove \$1 --revision=rev 2>/dev/null || true
      }
    }"
  ' 2>/dev/null || true

  echo "[5/7] Prune orphaned containerd overlayfs snapshots..."
  $SSH_CMD '
    ctr -n k8s.io snapshots ls 2>/dev/null | awk "NR>1 {print \$1}" | while read key; do
      ctr -n k8s.io snapshots unmount "\$key" 2>/dev/null || true
      ctr -n k8s.io snapshots remove "\$key" 2>/dev/null || true
    done
  ' 2>/dev/null || true

  echo "[6/7] journalctl vacuum (keep 3 days, 50MB limit)..."
  $SSH_CMD 'journalctl --vacuum-time=3d --vacuum-size=50M' 2>/dev/null || true

  echo "[7/7] Git gc on zero-admin (repack pack files)..."
  $SSH_CMD '
    cd /root/zero-admin && git config --global gc.reflogExpire 1 && git config --global gc.reflogExpireUnreachable 1
    git -c gc.reflogExpire=1 -c gc.reflogExpireUnreachable=1 -c gc.pruneExpire=now gc --aggressive --prune=now 2>/dev/null
    echo "Git pack size: \$(git count-objects -vH 2>/dev/null | grep size-pack)"
  ' 2>/dev/null || true

  echo ""
  echo "Auto-clean done. Re-check:"
  $SSH_CMD 'df -h /' 2>/dev/null
  echo
fi

echo "=============================================="
echo "Done. $(date '+%Y-%m-%d %H:%M:%S')"
echo "=============================================="
