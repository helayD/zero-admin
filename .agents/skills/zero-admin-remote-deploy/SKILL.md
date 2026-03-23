---
name: zero-admin-remote-deploy
description: Deploy the zero-admin repository to the private test server at 47.107.224.56. Use when the user asks for remote deployment, test-environment rollout, smoke verification, or applying a related SQL migration before deployment.
---

# Zero Admin Remote Deploy

## Overview

Use this skill when the user wants the current `zero-admin` workspace deployed to the private remote environment on `47.107.224.56`.

The deployment flow has exactly **one** publish path: push code to GitHub first, then let the server deploy from that git ref.

- The default path publishes the current branch if needed, rebases it onto the push target, pushes it to GitHub, and only then lets the server deploy that ref.
- `--git-ref` is the strict variant when you intentionally want to deploy an already-pushed ref from a clean local worktree.

This keeps every deploy traceable to GitHub and ensures the server is syncing the same code that was just published.

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
- Default deploy source: current branch published to GitHub, then remote git checkout of that ref
- Optional explicit git ref: `--git-ref <ref>`

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

## Workflow

1. Inspect changed files and choose a safe deploy scope.
2. Let the script publish the current branch to GitHub first, unless you are intentionally deploying an already-pushed ref with `--git-ref`.
3. Run `deploy_remote.sh` with the right `--services`, optional `--migration`, and optional `--git-ref`.
4. Read the script output for:
   - deploy source
   - pushed branch/commit
   - git source/ref used on the server
   - backup path
   - restarted processes
   - smoke results
5. If smoke fails, inspect the failing endpoint or remote process before retrying another rollout.

## Recommended Commands

Default deploy:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh
```

Deploy an already-pushed branch, tag, or commit via remote git sync:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --git-ref <branch|tag|commit>
```

Deploy a narrowed service slice:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --services admin-api,front-api,pms-rpc,oms-rpc,sms-rpc,cms-rpc,search-rpc,consumer,web-admin
```

Deploy the full project:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --services all
```

Deploy with migration:

```bash
bash .agents/skills/zero-admin-remote-deploy/scripts/deploy_remote.sh \
  --migration script/sql/<domain>/<migration-file>.sql
```

Smoke test only:

```bash
python3 .agents/skills/zero-admin-remote-deploy/scripts/smoke_remote.py \
  --base-url http://47.107.224.56:8000 \
  --front-base-url http://47.107.224.56:9999
```

## Guardrails

- There is only one deploy path: GitHub-backed git sync. Do not use or introduce local file-copy deployment variants for this skill.
- The default path publishes the current branch to GitHub before the server deploy starts.
- If you want to deploy the current work, do not skip the publish step. The whole point of the skill is that the server syncs the code that was just published.
- `--git-ref` is only for intentionally deploying an already-pushed ref, and it should be used from a clean local workspace.
- The script preserves remote source YAMLs for runtime configs during git sync and does not overwrite target runtime YAMLs if they already exist.
- The script always creates a timestamped backup under `/root/zero-admin/deploy-backup/<timestamp>`.
- The script can fetch from a configurable git remote URL. This matters because the remote test machine may not point at the same `origin` as the local workspace.
- The default behavior is project-wide deployment. Use `--services` only when you intentionally want a partial rollout.
- If smoke fails after restart, inspect the actual failing endpoint and relevant remote process before doing another rollout.

## Resources

### scripts/

- `deploy_remote.sh`: main deployment entrypoint with push-first GitHub sync, service scoping, migration, backup, restart, and smoke hooks.
- `smoke_remote.py`: admin smoke plus optional front smoke with clearer invalid-response diagnostics.
