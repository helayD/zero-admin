---
name: zero-admin-log-viewer
description: |
  Zero-Admin 远程服务器日志查看技能。
  支持查看所有服务的实时日志、错误日志、慢日志、按日期查历史日志。
  触发词：查日志、看日志、tail 日志、error log、服务报错、慢查询日志、日志路径
---

# Zero-Admin 日志查看

## 服务器信息

- **SSH**: `sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56`
- **日志根目录**: `/root/zero-admin/target/logs/`

---

## 服务与日志目录对应关系

| 服务名 | 日志子目录 | 端口 |
|--------|-----------|------|
| `admin-api` | `logs/admin/` | 8888 |
| `front-api` | `logs/front/` | 9999 |
| `sys-rpc` | `logs/sys/` | 8070 |
| `ums-rpc` | `logs/ums/` | 8081 |
| `pms-rpc` | `logs/pms/` | 8082 |
| `oms-rpc` | `logs/oms/` | 8083 |
| `sms-rpc` | `logs/sms/` | 8084 |
| `cms-rpc` | `logs/cms/` | 8086 |
| `search-rpc` | `logs/search/` | 8088 |
| `consumer` | `logs/consumer/` | — |
| `job` | `logs/job/` | — |

每个子目录日志文件类型：
- `access.log` — 当天访问日志（go-zero 自动按日轮转）
- `access.log-YYYY-MM-DD` — 历史访问日志
- `error.log` — 当天错误日志
- `error.log-YYYY-MM-DD` — 历史错误日志
- `slow.log` — 慢请求日志（超 500ms）
- `severe.log` — 严重错误日志
- `stat.log` — 统计日志

顶层重启日志（在 `logs/` 根目录）：
- `job-restart.log`
- `consumer-restart.log`
- `front-api-restart.log`
- `sms-rpc-restart.log`

---

## 数据库连接信息

| 中间件 | 连接方式 | 用户/密码 | 数据库 |
|--------|---------|----------|--------|
| **MySQL** | `127.0.0.1:3306` | `root / 12341qweqfsd2356` | `gozero` |
| **Redis** | `127.0.0.1:16379` | 密码: `123456` | — |
| **MongoDB** | `127.0.0.1:27017` | `admin / admin123` | — |
| **Elasticsearch** | `127.0.0.1:9200` | 无鉴权 | — |
| **Etcd** | `127.0.0.1:2379` | 无鉴权 | — |

> 所有连接信息均在服务器本地（127.0.0.1），需先 SSH 登录后再执行。

### MySQL 常用配合查询

```bash
SSH="sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56"
MYSQL="mysql -h127.0.0.1 -uroot -p'12341qweqfsd2356' gozero"

# 进入 MySQL 交互式终端
eval "$SSH '$MYSQL'"

# 非交互执行单条 SQL（推荐）
eval "$SSH \"$MYSQL -e 'SELECT id, mint_status, source_type, created_at FROM sms_card_instance ORDER BY id DESC LIMIT 20;'\""

# 查卡片实例最近状态
eval "$SSH \"$MYSQL -e 'SELECT id, mint_status, chain_status, transferable, created_at FROM sms_card_instance WHERE mint_status != \"mint_success\" ORDER BY id DESC LIMIT 20;'\""

# 查发卡任务队列
eval "$SSH \"$MYSQL -e 'SELECT id, instance_id, status, retry_count, last_error, created_at FROM sms_card_mint_task ORDER BY id DESC LIMIT 20;'\""

# 查抽卡参与记录
eval "$SSH \"$MYSQL -e 'SELECT id, user_id, activity_id, result_type, result_status, created_at FROM sms_draw_participation_record ORDER BY id DESC LIMIT 20;'\""

# 查订单
eval "$SSH \"$MYSQL -e 'SELECT id, order_sn, status, total_amount, created_at FROM oms_order ORDER BY id DESC LIMIT 10;'\""
```

### Redis 常用配合查询

```bash
SSH="sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56"
REDIS="redis-cli -p 16379 -a 123456"

# 查询 keys（慎用 KEYS *，改用 SCAN）
eval "$SSH '$REDIS SCAN 0 MATCH \"*token*\" COUNT 20'"

# 查看登录 token
eval "$SSH '$REDIS KEYS \"*login*\"'"

# 查 JWT / Session
eval "$SSH '$REDIS TTL <key>'"
```

### 脚本一键执行 SQL

```bash
# 用脚本：bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh db <sql>
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh db "SELECT id, mint_status FROM sms_card_instance ORDER BY id DESC LIMIT 10"
```

---

## 常用查询命令

### 实时追踪日志（tail -f）

```bash
# 追踪 admin-api 错误日志
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'tail -f /root/zero-admin/target/logs/admin/error.log'

# 追踪 sms-rpc 错误日志
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'tail -f /root/zero-admin/target/logs/sms/error.log'

# 追踪 job（发卡任务）日志
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'tail -100 /root/zero-admin/target/logs/job/error.log'
```

### 查看最近 N 行日志

```bash
# 用脚本：bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh <service> <log-type> [lines]
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh admin error 100
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh sms error 50
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh job error 200
```

### 搜索关键词

```bash
# 搜索 admin-api 错误日志中的关键词
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'grep -i "panic\|fatal\|error" /root/zero-admin/target/logs/admin/error.log | tail -50'

# 搜索 mint 相关错误（发卡链路）
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'grep -i "mint\|card" /root/zero-admin/target/logs/sms/error.log | tail -50'
```

### 查看历史日志（指定日期）

```bash
# 查看 pms-rpc 2026-05-20 的错误日志
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'cat /root/zero-admin/target/logs/pms/error.log-2026-05-20'
```

### 查看所有服务的当前错误（一键巡检）

```bash
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh all error 20
```

### 查看慢日志

```bash
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'tail -50 /root/zero-admin/target/logs/sys/slow.log'
```

### 列出日志目录（查看有哪些历史日志文件）

```bash
sshpass -p 'Qianmai1#' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@47.107.224.56 \
  'ls -lh /root/zero-admin/target/logs/<service>/'
```

---

## 使用脚本

提供封装脚本 `scripts/query_logs.sh`，一行命令完成日志查询：

```
用法: query_logs.sh <service|all> <log-type> [lines]

service:  admin | front | sys | ums | pms | oms | sms | cms | search | consumer | job | all
log-type: access | error | slow | severe | restart
lines:    默认 100
```

示例：
```bash
# 查 sms-rpc 最近 50 条错误
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh sms error 50

# 查所有服务的最近 20 条错误（快速巡检）
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh all error 20

# 查 job 服务的重启日志
bash .agents/skills/zero-admin-log-viewer/scripts/query_logs.sh job restart 100
```

---

## 日志格式说明（go-zero）

go-zero 的日志是 JSON 格式，关键字段：
- `"level"`: `"info"` / `"error"` / `"severe"`
- `"ts"`: 时间戳（RFC3339）
- `"msg"` / `"content"`: 消息内容
- `"trace"`: 链路 trace ID
- `"span"`: span ID
- `"caller"`: 调用文件:行号

搜索 error 日志时可以：
```bash
grep '"level":"error"' /root/zero-admin/target/logs/admin/access.log | tail -20
```

---

## 何时使用此 Skill

- 用户说"看一下 xxx 服务的日志"、"查一下报错"、"tail 日志"
- 远程部署后排查服务是否正常启动
- 发卡链路故障排查（结合 zero-admin-card-issuance skill）
- 查看构建/重启后的启动日志
