# Project Deployment Memory

Purpose: Record project-specific deployment and remote environment information for the `zero-admin` repository only.

## Scope

- This file is project memory for `/Users/helay/Documents/GitHub/zero-admin`.
- Do not treat these credentials or endpoints as global Codex memory for other projects.
- User explicitly confirmed this is a private/self-hosted project and allowed plaintext recording here.

## Historical Source

- Source file: [`_opcos/project-context.md`](/Users/helay/Documents/GitHub/zero-admin/_opcos/project-context.md)
- Historical commit containing full deployment details: `4d22c35fae23872b1b45b0a57d1fe022b5ca4560`
- Historical date: `2026-03-15`

## Remote Deployment Details From Historical Project Context

### SSH Host

- Host: `47.107.224.56`
- Port: `22`
- Username: `root`
- Password: `Qianmai1#`
- SSH command: `ssh root@47.107.224.56`

### Infrastructure

- MySQL
  - Host: `47.107.224.56`
  - Port: `3306`
  - Database: `gozero`
  - Username: `root`
  - Password: `12341qweqfsd2356`
- Redis
  - Host: `47.107.224.56`
  - Port: `16379`
  - Password: `123456`
- MongoDB
  - Host: `47.107.224.56`
  - Port: `27017`
  - Username: `admin`
  - Password: `admin123`
- RabbitMQ
  - Host: `47.107.224.56`
  - Port: `5672`
  - Username: `test`
  - Password: `test`
- Elasticsearch
  - Host: `47.107.224.56`
  - Port: `9200`
  - Cluster: `opencoze`
- Etcd
  - Port: `2379`

### Gateways

- Admin API: `http://47.107.224.56:8888`
- Front API: `http://47.107.224.56:9999`
- Swagger: `http://47.107.224.56:8888/swagger`
- Admin Web: `http://47.107.224.56:8000`
- Flutter production API: `http://47.107.224.56:9999`

### RPC Services

- `sys-rpc`: `8070`
- `ums-rpc`: `8081`
- `pms-rpc`: `8082`
- `oms-rpc`: `8083`
- `sms-rpc`: `8084`
- `cms-rpc`: `8085`
- `search-rpc`: `8088`

### Application Accounts

- Admin Web
  - Username: `admin`
  - Password: `123456`

## Current Repository Config Difference

The current repository also contains a different remote MySQL configuration in multiple `k8s.yaml` files:

- Example file: [`rpc/sys/etc/k8s.yaml`](/Users/helay/Documents/GitHub/zero-admin/rpc/sys/etc/k8s.yaml)
- Host: `110.41.179.89`
- Port: `30395`
- Database: `gozero`
- Username: `root`
- Password: `r-wz9wop62956dh5k9ed`

This means the project currently has at least two remembered remote-environment sources:

1. Historical project context remote host: `47.107.224.56`
2. Current k8s config remote MySQL host: `110.41.179.89:30395`

Do not assume these are the same environment without verification.

## Usage Notes

- Prefer project-local config and deployment files over outdated generator defaults such as hardcoded localhost DSNs.
- If database generation or deployment work fails, verify whether the intended target is:
  - historical host `47.107.224.56`, or
  - current `k8s.yaml` host `110.41.179.89:30395`
- Before regenerating GORM models or RPC assets, confirm the active environment source of truth.
