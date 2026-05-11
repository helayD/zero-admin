---
name: zero-admin-remote-deploy
description: Deploy the zero-admin repository to the private test server at 47.107.224.56. Use when the user asks for remote deployment, test-environment rollout, smoke verification, or applying a related SQL migration before deployment.
---

# Zero Admin Remote Deploy

## Overview

Use this skill when the user wants the current `zero-admin` workspace deployed to the private remote environment on `47.107.224.56`.

The target experience for this skill is **zero-interaction automatic deploy**:

- When the user says "发布到远程/测试环境", assume they want the current workspace published end-to-end unless they explicitly narrow the scope.
- Complete the normal path automatically: inspect changes, run focused local verification, commit if needed, push, clean disk if needed, deploy, smoke test, and run Story API tests.
- Do **not** pause for routine confirmations during that path. Only stop on a hard blocker such as failing local tests, failed push, failed cleanup that still leaves disk >= 90%, failed migration, failed build, or failed smoke/API tests.

The deployment flow still has exactly **one** publish path: GitHub-backed git sync. The server must deploy an explicit git ref, but the skill should prepare that ref automatically when the user wants to deploy current local work.

## Automation Defaults

When the user asks to deploy current work, follow these defaults automatically:

1. Inspect `git status --short --branch`.
2. If the worktree is dirty, run the most relevant local verification that can be inferred from changed files.
3. If verification passes, auto-commit the current workspace with a concise deploy-oriented commit message.
4. If the branch is ahead or the commit is new, push it automatically to GitHub.
5. Resolve the deploy ref from the pushed branch/commit and pass it explicitly to `deploy_remote.sh --git-ref`.
6. Run disk preflight with automatic cleanup if needed.
7. Infer a minimal safe deploy scope from changed files; if inference is ambiguous, fall back to `all`.
8. Auto-include migrations when changed files indicate repo SQL migrations are part of the change set.

No confirmation pause is needed for those defaults.

## Disk Health Check (Automatic Pre-Deploy Gate)

**Server: 47.107.224.56 — After cleanup baseline: 81G / 99G, 86% used, 14 GB free. Baseline is healthy.**

Before any deploy, run the disk check. If disk is above 85%, automatically run safe cleanup and re-check before deciding whether to continue.

```bash
ssh -o ConnectTimeout=10 root@47.107.224.56 'df -h / && echo "---INODES---" && df -i /'
```

### Known Disk Consumers (measured 2026-03-29, post-cleanup)

| Path | Usage | Actionable |
|---|---|---|
| `/` (root fs) | 81G / 99G, 86% | ✅ OK baseline |
| `/root/zero-admin/.git/` | 209 MB (was 4.74 GB) | `git gc` periodically |
| `/usr/share/elasticsearch/` | ~1.3 GB | Do not remove — ES running |
| `/var/lib/snapd/` | 596 MB | Prune disabled snap revisions |
| `/root/go/pkg/mod/` | 0 MB (cleaned) | Go build cache — on demand |
| `/root/.npm/_cacache/` | 0 MB (cleaned) | npm cache — on demand |
| `/root/.windsurf-server/` | 0 MB (cleaned) | Windsurf — on demand |
| `/var/log/journal/` | 57 MB (was 385 MB) | `journalctl --vacuum-time=3d` |
| `/var/lib/mongodb/` | 421 MB | Data dir — do not touch |
| `/var/lib/mysql/` | 258 MB | Data dir — do not touch |
| `/var/lib/etcd/` | 123 MB | K8s data — do not touch |
| `/var/lib/containerd/` | ~380 MB | Prune orphaned snapshots |
| `docker system df` | ~500 MB | `docker system prune -f` safe |

### Disk Cleanup Commands (non-destructive, auto-runnable)

```bash
# apt cache
ssh root@47.107.224.56 'apt-get clean && apt-get autoremove -y'

# Docker prune
ssh root@47.107.224.56 'docker system prune -f'

# Journal vacuum (keep 3 days, 50MB cap)
ssh root@47.107.224.56 'journalctl --vacuum-time=3d --vacuum-size=50M'

# Git gc on zero-admin (reclaim ~4 GB over time as packs fragment)
ssh root@47.107.224.56 'cd /root/zero-admin && git -c gc.pruneExpire=now gc --aggressive --prune=now'

# Containerd orphaned snapshots
ssh root@47.107.224.56 'ctr -n k8s.io snapshots ls | awk "NR>1 {print \$1}" | while read k; do ctr -n k8s.io snapshots remove "$k" 2>/dev/null || true; done'

# Snap prune
ssh root@47.107.224.56 'snap list --all | awk "/lxd|core20/ && \$5 ~ /disabled/ {print \$1,\$3}" | while read n r; do snap remove "$n" --revision="$r" 2>/dev/null; done'
```

## When To Use

- The user says "远程部署", "发到测试环境", "部署到 47.107.224.56", or asks to verify the deployed environment.
- Backend files under `api/`, `rpc/`, `consumer/`, or `job/` changed and need to be tested remotely.
- Frontend files under `web-admin/` changed and need to be published to the remote admin site.
- A repo SQL migration needs to be applied together with the deployment.

## Defaults

- SSH target: `root@47.107.224.56`
- Remote root: `/root/zero-admin`
- Admin smoke base URL: `http://47.107.224.56:8000`
- Front smoke base URL: `http://47.107.224.56:9999`
- Admin smoke account: `admin / 123456`
- MySQL backup/apply host on the remote server: `127.0.0.1`
- Default deploy services: `all`
- Default deploy source: current branch HEAD after automatic commit/push when needed
- Deploy command should still use explicit git ref: `--git-ref <ref>`

## Service Matrix

The deploy script now understands the real service layout of this repo:

- `admin-api`
- `front-api`
- `sys-rpc`
- `ums-rpc`
- `pms-rpc`
- `oms-rpc`
- `sms-rpc`
- `cms-rpc`
- `search-rpc`
- `consumer`
- `job`
- `web-admin`

The skill should infer the smallest safe scope from changed files first, and only fall back to full-project deployment when the touched area is broad or ambiguous.

## Deploy Pipeline (Fully Automatic)

**Step 0 — Local Publish Gate**

Run before everything else. Inspect git state, run focused verification, auto-commit if needed, and push automatically.

**Step 1 — Disk Health Gate**

Run only after Step 0 succeeds. If root fs usage is >= 85%, run safe cleanup automatically and re-check. Continue automatically when post-clean usage is < 90%; abort only when it remains >= 90%.

```bash
ssh -o ConnectTimeout=10 root@47.107.224.56 'df -h /'
```

If usage > 85%: reclaim space automatically, verify again, and proceed without asking unless the post-clean result is still at or above the hard stop threshold.

2. **Verify SSH** — confirm remote host is reachable
3. **Prepare** — identify build scope
4. **Sync source** — `git fetch` on remote (with `http.version=HTTP/1.1`) for the already-pushed ref
5. **Migration** — upload & execute DDL/seed SQL (manual `--migration` or `--auto-migration`)
6. **Build** — compile Go services and/or build web-admin on the remote host
7. **Swap** — backup old binaries, install new, restart services
8. **Smoke test** — basic admin + front API health checks
9. **Story API test** — run all `script/shell/api-test/` story tests, **require 100% pass rate**

## Workflow

1. Inspect changed files and infer the deploy scope plus likely migration/test impact.
2. Run focused local verification automatically.
3. Auto-commit and push current branch when local changes are present or the branch is ahead.
4. Run `check_disk.sh --auto-clean-if-needed`.
5. Run `deploy_remote.sh` with the resolved `--git-ref`, inferred `--services`, and auto migration flags when applicable.
6. Read the script output for:
   - deploy source
   - git source/ref used on the server
   - backup path
   - restarted processes
   - smoke results
   - **Story API test results (must be 100% pass)**
7. If smoke or API tests fail, inspect the failing endpoint or remote process and treat the deploy as failed. Report findings instead of asking routine confirmation questions.

## Recommended Commands

Automatic disk gate:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/check_disk.sh --auto-clean-if-needed
```

Automatic deploy of current branch:

```bash
branch="$(git rev-parse --abbrev-ref HEAD)"
git push origin "$branch"
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh --git-ref "$branch"
```

Deploy an already-pushed branch, tag, or commit via remote git sync:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit>
```

Deploy a narrowed service slice:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --services admin-api,front-api,pms-rpc,oms-rpc,sms-rpc,cms-rpc,search-rpc,consumer,web-admin
```

Deploy the full project:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --services all
```

Deploy with explicit migration:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --migration script/sql/migration_20260326_deploy_seed.sql
```

Deploy with auto-migration (recursive under `script/sql/`):

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --auto-migration
```

Deploy with specific story API tests only:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --stories 4-5,4-6,5-1,5-2
```

Skip API tests (smoke only):

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit> \
  --skip-api-test
```

Run Story API tests standalone:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/run_api_tests.sh \
  --admin-url http://47.107.224.56:8000 \
  --front-url http://47.107.224.56:9999
```

Smoke test only:

```bash
python3 .agents/skills/zero-admin-remote-deploy/scripts/smoke_remote.py \
  --base-url http://47.107.224.56:8000 \
  --front-base-url http://47.107.224.56:9999
```

## Guardrails

- There is only one deploy path: GitHub-backed git sync. Do not use or introduce local file-copy deployment variants for this skill.
- Treat "发布到远程/测试环境" as authorization to execute the full normal deploy path automatically.
- Do not stop for routine confirmation at local git gate, push, disk cleanup, service-scope inference, or migration auto-discovery.
- Always ensure any remote deploy still uses an explicit git ref, even when that ref was just prepared automatically from the current branch.
- The script preserves remote source YAMLs for runtime configs during git sync and does not overwrite target runtime YAMLs if they already exist.
- The script always creates a timestamped backup under `/root/zero-admin/deploy-backup/<timestamp>`.
- The script can fetch from a configurable git remote URL. This matters because the remote test machine may not point at the same `origin` as the local workspace.
- Infer the smallest safe deploy scope when possible; fall back to project-wide deployment when uncertain.
- `--auto-migration` recursively scans `script/sql/**/migration_*.sql`, so nested domain migrations are eligible for automatic inclusion.
- If smoke fails after restart, inspect the actual failing endpoint and relevant remote process before doing another rollout.
- Hard blockers are allowed to stop the flow. Examples: local tests fail, git push fails, disk usage remains >= 90% after auto-clean, migration apply fails, remote build fails, smoke fails, or Story API tests fail.

## Internal Safeguards (auto-enabled)

### Target YAML RPC Client Sync
When deploying `admin-api` or `front-api`, if the source `etc/*.yaml` has a new RPC client block (e.g. `SearchRpc`) that the remote `target/*/...yaml` is missing, the script automatically copies that block from source to target. This prevents startup failures due to missing ServiceContext dependencies.

### Runtime Config Values Are NOT Synced
配置**值**不会被部署脚本同步——`etc/*.yaml` 在 git checkout 前会被备份、之后再 restore，目的是保护远程上线后的环境特定值。

实际影响：本地仓库改 `Share.AllowedDomains` / `Auth.AccessSecret` / `Mysql.Datasource` / 任何端口/域名/密钥，远程不会自动更新。

**修改远程运行时配置的正确流程**：
1. SSH 直改远程目标文件，例如 `/root/zero-admin/target/front-api/front-api.yaml`
2. `pkill -f '<service>/<service>'` 停旧进程
3. `cd /root/zero-admin/target && setsid nohup ./<svc>/<svc> -f ./<svc>/<svc>.yaml > ./logs/<svc>.log 2>&1 < /dev/null &` 重启
4. 验证：`ss -tlnp \| grep <port>` 与 `tail logs/<svc>.log`

**已踩过的坑**：`Share.AllowedDomains` 默认是 `example.com` 占位符，导致 Story 10.7 分享链接 `https://example.com/h5/...` 在 Flutter / H5 都无法打开，必须手动改远程为 `47.107.224.56:9999`（详见 2026-05 修复记录）。

### RPC Port Mismatch Detection
Before deploying, the script cross-checks each API service's RPC client `Endpoints` against the actual `ListenOn` port in the corresponding RPC server config. On mismatch, a `[WARN]` is printed so the operator catches port misconfigurations before they cause runtime gRPC connection errors.

### Restart Fallback
If `pgrep` verification fails after `nohup` restart, the script performs a secondary process check via `pgrep -f`. If the process is confirmed running, deployment proceeds. This handles cases where the service starts but the `pgrep` pattern doesn't match (e.g. different binary path).

## Resources

### scripts/

- `deploy_remote.sh`: main deployment entrypoint with git-based remote sync, service scoping, migration (manual + auto-discovery), target YAML RPC client sync, port mismatch detection, restart fallback, backup, smoke hooks, and Story API test gate. The skill should prepare the git ref first, then invoke this script with an explicit `--git-ref`.
- `smoke_remote.py`: admin smoke plus optional front smoke with clearer invalid-response diagnostics.
- `run_api_tests.sh`: Story API test runner. Auto-discovers test scripts under `script/shell/api-test/<story-id>/test_*.sh`. Maps story prefixes to correct base URLs (4-* → admin, 5-* → front). Requires 100% pass rate.
- `check_disk.sh`: remote disk health check. Run after local publish prep and before every deploy. Supports `--auto-clean-if-needed` for zero-interaction flows and returns non-zero when post-clean disk usage is still at or above the hard stop threshold.

### Story API Tests (`script/shell/api-test/`)

Each story has its own directory with shell test scripts:

- `4-5/test_4_5_api.sh` — 购物车与确认单促销试算 (front-api)
- `4-6/test_4_6_api.sh` — 配置生效状态与作用域下发校验 (admin-api)
- `5-1/test_5_1_api.sh` — 商品加购与商户归属校验 (front-api)
- `5-2/test_5_2_api.sh` — 购物车编辑、删除与批量结算 (front-api)

To add a new story test: create `script/shell/api-test/<story-id>/test_<story-id>_api.sh`. It will be auto-discovered on the next deploy.

### Migration SQL (`script/sql/`)

Files matching `migration_*.sql` under `script/sql/` are auto-discovered by `--auto-migration`. This scan is recursive. Naming convention: `migration_<YYYYMMDD>_<description>.sql`.
