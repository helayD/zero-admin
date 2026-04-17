---
name: zero-admin-remote-deploy
description: Deploy the zero-admin repository to the private test server at 47.107.224.56. Use when the user asks for remote deployment, test-environment rollout, smoke verification, or applying a related SQL migration before deployment.
---

# Zero Admin Remote Deploy

## Overview

Use this skill when the user wants the current `zero-admin` workspace deployed to the private remote environment on `47.107.224.56`.

The deployment flow has exactly **one** publish path: deploy only from a clean, committed, already-pushed git ref.

- Before any remote action, confirm the local worktree is clean and the intended branch/tag/commit is already on GitHub.
- For Codex usage of this skill, do **not** rely on `deploy_remote.sh` auto-commit or auto-publish behavior from a dirty worktree.
- `--git-ref` is the required deploy variant for this skill because it makes the target ref explicit and reviewable.

This keeps every deploy traceable to GitHub and prevents accidental release of uncommitted or implicitly committed local changes.

## Local Git Gate (Pre-Deploy Gate)

Before running SSH, disk checks, migrations, smoke tests, or the deploy script, inspect the local git state first.

```bash
git status --short --branch
```

Rules:

- If the worktree is dirty: stop immediately, show the user the status, and ask them whether to commit/push first or want help preparing a commit. Do **not** start any remote deploy action yet.
- If the target ref is not pushed: ask the user to confirm the branch/tag/commit to publish, then push first.
- Only after the local ref is clean and already available on GitHub may this skill proceed to the remote deploy steps.

## Disk Health Check (Pre-Deploy Gate)

**Server: 47.107.224.56 — After cleanup baseline: 81G / 99G, 86% used, 14 GB free. Baseline is healthy.**

Before any deploy, run the disk check. If disk is above 85%, surface the warning to the user and do not proceed without explicit confirmation.

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

### Disk Cleanup Commands (non-destructive, run with confirmation)

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
- Default deploy source: a clean branch/tag/commit that is already pushed to GitHub
- Required explicit git ref: `--git-ref <ref>`

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

The default deploy scope is the full project. Use `--services <csv>` when you intentionally want a partial rollout.

## Deploy Pipeline (2 Gates + 8 Steps)

**Step 0 — Local Git Gate (precondition, non-negotiable)**

Run before everything else. If the worktree is dirty or the intended ref is not yet pushed, stop and get user confirmation first.

**Step 1 — Disk Health Check (precondition, non-negotiable)**

Run only after Step 0 succeeds. Abort deploy if root fs usage > 85% and report the exact usage.

```bash
ssh -o ConnectTimeout=10 root@47.107.224.56 'df -h /'
```

If usage > 85%: surface a warning with the exact numbers, reclaim space (see Disk Health Check above), verify again, then proceed only after explicit confirmation.

2. **Verify SSH** — confirm remote host is reachable
3. **Prepare** — identify build scope
4. **Sync source** — `git fetch` on remote (with `http.version=HTTP/1.1`) for the already-pushed ref
5. **Migration** — upload & execute DDL/seed SQL (manual `--migration` or `--auto-migration`)
6. **Build** — compile Go services and/or build web-admin on the remote host
7. **Swap** — backup old binaries, install new, restart services
8. **Smoke test** — basic admin + front API health checks
9. **Story API test** — run all `script/shell/api-test/` story tests, **require 100% pass rate**

## Workflow

1. Inspect changed files and choose a safe deploy scope.
2. Inspect `git status --short --branch` and stop if the worktree is dirty. Do not start any remote deploy action until the user confirms the commit/push plan.
3. Push the intended branch/tag/commit to GitHub first, then deploy by explicit `--git-ref`.
4. Run `deploy_remote.sh` with the right `--git-ref`, `--services`, and optional `--migration` / `--auto-migration`.
5. Read the script output for:
   - deploy source
   - git source/ref used on the server
   - backup path
   - restarted processes
   - smoke results
   - **Story API test results (must be 100% pass)**
6. If smoke or API tests fail, inspect the failing endpoint or remote process before retrying another rollout.

## Recommended Commands

Local git gate (run before every deploy):

```bash
git status --short --branch
```

Disk health check (run only after the git gate passes):

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/check_disk.sh
```

Auto-clean + re-check:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/check_disk.sh --auto-clean
```

Push the current branch first:

```bash
git push origin <branch>
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

Deploy with auto-migration (scans root `script/sql/migration_*.sql` only):

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
- Before any remote action, inspect `git status --short --branch`.
- If the local worktree is dirty, stop and ask the user whether to commit/push first. Do not SSH to the server, do not run disk checks, and do not start the deploy script yet.
- For this skill, never rely on `deploy_remote.sh` auto-commit or auto-publish behavior from a dirty worktree.
- Always deploy by explicit `--git-ref`, and that ref must already exist on GitHub.
- The script preserves remote source YAMLs for runtime configs during git sync and does not overwrite target runtime YAMLs if they already exist.
- The script always creates a timestamped backup under `/root/zero-admin/deploy-backup/<timestamp>`.
- The script can fetch from a configurable git remote URL. This matters because the remote test machine may not point at the same `origin` as the local workspace.
- The default behavior is project-wide deployment. Use `--services` only when you intentionally want a partial rollout.
- `--auto-migration` only scans `script/sql/` at the repo root. Migrations under nested folders such as `script/sql/sms/` are not auto-discovered by that flag.
- If smoke fails after restart, inspect the actual failing endpoint and relevant remote process before doing another rollout.

## Internal Safeguards (auto-enabled)

### Target YAML RPC Client Sync
When deploying `admin-api` or `front-api`, if the source `etc/*.yaml` has a new RPC client block (e.g. `SearchRpc`) that the remote `target/*/...yaml` is missing, the script automatically copies that block from source to target. This prevents startup failures due to missing ServiceContext dependencies.

### RPC Port Mismatch Detection
Before deploying, the script cross-checks each API service's RPC client `Endpoints` against the actual `ListenOn` port in the corresponding RPC server config. On mismatch, a `[WARN]` is printed so the operator catches port misconfigurations before they cause runtime gRPC connection errors.

### Restart Fallback
If `pgrep` verification fails after `nohup` restart, the script performs a secondary process check via `pgrep -f`. If the process is confirmed running, deployment proceeds. This handles cases where the service starts but the `pgrep` pattern doesn't match (e.g. different binary path).

## Resources

### scripts/

- `deploy_remote.sh`: main deployment entrypoint with git-based remote sync, service scoping, migration (manual + auto-discovery), target YAML RPC client sync, port mismatch detection, restart fallback, backup, smoke hooks, and Story API test gate. This skill must invoke it only with an already-pushed explicit `--git-ref`.
- `smoke_remote.py`: admin smoke plus optional front smoke with clearer invalid-response diagnostics.
- `run_api_tests.sh`: Story API test runner. Auto-discovers test scripts under `script/shell/api-test/<story-id>/test_*.sh`. Maps story prefixes to correct base URLs (4-* → admin, 5-* → front). Requires 100% pass rate.
- `check_disk.sh`: remote disk health check. Run after the local git gate passes and before every deploy. Exits with warning (85%) or abort (90%). Supports `--auto-clean` to reclaim space (apt clean, docker prune, snap prune, tmp cleanup).

### Story API Tests (`script/shell/api-test/`)

Each story has its own directory with shell test scripts:

- `4-5/test_4_5_api.sh` — 购物车与确认单促销试算 (front-api)
- `4-6/test_4_6_api.sh` — 配置生效状态与作用域下发校验 (admin-api)
- `5-1/test_5_1_api.sh` — 商品加购与商户归属校验 (front-api)
- `5-2/test_5_2_api.sh` — 购物车编辑、删除与批量结算 (front-api)

To add a new story test: create `script/shell/api-test/<story-id>/test_<story-id>_api.sh`. It will be auto-discovered on the next deploy.

### Migration SQL (`script/sql/`)

Files matching `migration_*.sql` under `script/sql/` are auto-discovered by `--auto-migration`. This scan is not recursive. Naming convention: `migration_<YYYYMMDD>_<description>.sql`.
