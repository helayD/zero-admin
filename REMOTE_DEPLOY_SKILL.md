# 远程部署 Skill

## 服务器信息

| 项目 | 值 |
|---|---|
| 服务器 IP | 47.107.224.56 |
| SSH 用户 | root |
| 数据库 | MySQL 8.x |
| 数据库名 | gozero |
| 数据库用户 | root |

## 连接方式

```bash
# SSH 登录
ssh root@47.107.224.56

# MySQL 命令行（需要输入密码）
mysql -u root -p gozero
```

## 项目部署路径

远程服务器上的项目路径（假设与本地一致）：
- `/opt/zero-admin/` 或 `/root/zero-admin/`（需确认）

## 常用部署命令

### 1. 编译后端服务

```bash
# 在本地编译
go build -o admin-api ./api/admin/admin.go
go build -o front-api ./api/front/front.go
go build -o sms-rpc ./rpc/sms/sms.go
go build -o pms-rpc ./rpc/pms/pms.go
go build -o oms-rpc ./rpc/oms/oms.go
go build -o consumer ./consumer/consumer.go
go build -o job-server ./job/job.go

# 或通过 Makefile（如果有）
make build-all
```

### 2. 上传到远程服务器

```bash
# 使用 scp 上传二进制文件
scp admin-api root@47.107.224.56:/opt/zero-admin/
scp front-api root@47.107.224.56:/opt/zero-admin/
# ... 其他服务
```

### 3. 重启远程服务

```bash
# SSH 到远程后执行
ssh root@47.107.224.56

# 停止旧服务
systemctl stop zero-admin-api  # 或 supervisorctl stop ...

# 替换二进制
cp /opt/zero-admin/admin-api.new /opt/zero-admin/admin-api

# 启动服务
systemctl start zero-admin-api
# 或使用 nohup
nohup /opt/zero-admin/admin-api -f /opt/zero-admin/etc/admin-api.yaml > /var/log/zero-admin/admin-api.log 2>&1 &
```

### 4. 数据库 Migration

```bash
# 远程执行 SQL migration
ssh root@47.107.224.56 "mysql -u root -p'{DB_PASSWORD}' gozero < /opt/zero-admin/script/sql/sms/migration_xxx.sql"
```

### 5. 前端部署（Web Admin）

```bash
# 本地构建
cd web-admin
npm run build

# 上传到远程 Nginx 目录
scp -r dist/* root@47.107.224.56:/usr/share/nginx/html/admin/

# 或重启 Nginx
ssh root@47.107.224.56 "nginx -s reload"
```

## 配置检查清单

部署后需要确认的配置项：

- [ ] `api/admin/etc/admin-api.yaml` - Admin API 监听地址和端口
- [ ] `api/front/etc/front-api.yaml` - Front API 监听地址和端口
- [ ] `rpc/sms/etc/sms.yaml` - SMS RPC Etcd 注册地址
- [ ] `rpc/pms/etc/pms.yaml` - PMS RPC Etcd 注册地址
- [ ] `rpc/oms/etc/oms.yaml` - OMS RPC Etcd 注册地址
- [ ] `consumer/etc/consumer.yaml` - RabbitMQ 连接配置
- [ ] `job/etc/job.yaml` - Job 调度配置
- [ ] 数据库连接配置中的 `host` 应为 `localhost` 或 `127.0.0.1`

## 服务健康检查

```bash
# 检查各服务端口是否监听
ss -tlnp | grep -E '8080|8081|8082|9090'

# 检查进程
ps aux | grep zero-admin

# 检查日志
tail -f /var/log/zero-admin/*.log
```

## 回滚策略

1. 部署前备份旧二进制：`cp admin-api admin-api.bak.$(date +%Y%m%d%H%M%S)`
2. 数据库变更前导出相关表：`mysqldump -u root -p gozero sms_card_instance > backup.sql`
3. 发现问题立即停止服务并回退到备份版本

## 注意事项

1. **Go 版本**：远程服务器需安装 Go 1.22+ 或预编译二进制
2. **Etcd**：确保远程 Etcd 集群运行正常
3. **RabbitMQ**：确保 RabbitMQ 服务运行且队列正确声明
4. **Nginx**：Web Admin 前端需 Nginx 做静态文件服务
5. **防火墙**：确认服务器防火墙开放所需端口
6. **环境变量**：生产环境配置不应提交到 Git，应使用环境变量或单独的配置文件
