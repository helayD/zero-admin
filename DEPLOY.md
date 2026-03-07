# Zero-Admin 部署文档

## 项目简介

Zero-Admin 是一个基于 [go-zero](https://go-zero.dev/) 微服务框架的后台管理系统，包含 7 个 RPC 服务、2 个 API 网关、1 个定时任务服务和 1 个消息消费者服务。

## 目录结构

```
zero-admin/
├── api/
│   ├── admin/          # 后台管理 API（端口 8888）
│   └── front/          # 前台用户 API（端口 9999）
├── rpc/
│   ├── sys/            # 系统服务 RPC（端口 8070）
│   ├── ums/            # 会员服务 RPC（端口 8081）
│   ├── pms/            # 商品服务 RPC（端口 8082）
│   ├── oms/            # 订单服务 RPC（端口 8083）
│   ├── sms/            # 营销服务 RPC（端口 8084）
│   ├── cms/            # 内容服务 RPC（端口 8086）
│   └── search/         # 搜索服务 RPC（端口 8088）
├── job/                # 定时任务（端口 8089）
├── consumer/           # 消息消费者（端口 8087）
├── script/
│   └── sql/            # 数据库初始化脚本
└── makefile            # 构建脚本（Linux/macOS）
```

## 服务端口总览

| 服务名称     | 类型 | 端口  | 说明         |
|-------------|------|-------|-------------|
| sys-rpc     | RPC  | 8070  | 系统管理     |
| ums-rpc     | RPC  | 8081  | 会员管理     |
| pms-rpc     | RPC  | 8082  | 商品管理     |
| oms-rpc     | RPC  | 8083  | 订单管理     |
| sms-rpc     | RPC  | 8084  | 营销管理     |
| cms-rpc     | RPC  | 8086  | 内容管理     |
| search-rpc  | RPC  | 8088  | 搜索服务     |
| admin-api   | API  | 8888  | 后台管理网关  |
| front-api   | API  | 9999  | 前台用户网关  |
| job         | API  | 8089  | 定时任务     |
| consumer    | API  | 8087  | 消息消费者    |

## 环境要求

### 基础环境

- **Go** >= 1.21（已测试 Go 1.26.1）
- **Git**

### 中间件依赖

| 中间件          | 默认端口 | 用途             | 必需 |
|----------------|---------|-----------------|------|
| MySQL 8.0      | 3306    | 主数据库          | ✅   |
| Etcd 3.5+      | 2379    | 服务注册与发现     | ✅   |
| Redis 7+       | 6379    | 缓存             | ✅   |
| MongoDB 5.0+   | 27017   | 会员/商品文档存储   | ✅   |
| RabbitMQ 3.x   | 5672    | 消息队列          | ✅   |
| Elasticsearch 8 | 9200   | 全文搜索          | ✅   |

---

## 部署步骤

### 1. 安装 Go 环境

#### Windows（Chocolatey）
```powershell
# 以管理员身份运行
choco install golang -y
```

#### Windows（手动安装）
从 https://go.dev/dl/ 下载 `.msi` 安装包，双击安装。

#### Linux
```bash
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

安装后验证：
```bash
go version
```

### 2. 启动中间件（Docker 方式）

```bash
# MySQL
docker run -d --name zero-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=12341qweqfsd2356 \
  -e MYSQL_DATABASE=gozero \
  mysql:8.0

# Etcd
docker run -d --name zero-etcd \
  -p 2379:2379 \
  -e ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 \
  -e ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 \
  quay.io/coreos/etcd:v3.5.9

# Redis（密码按配置文件设定）
docker run -d --name zero-redis \
  -p 6379:6379 \
  redis:7 --requirepass "123456"

# MongoDB
docker run -d --name zero-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin123 \
  mongo:5.0

# RabbitMQ
docker run -d --name zero-rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=guest \
  -e RABBITMQ_DEFAULT_PASS=guest \
  rabbitmq:3-management

# Elasticsearch
docker run -d --name zero-es \
  -p 9200:9200 \
  -e discovery.type=single-node \
  -e xpack.security.enabled=false \
  -e ES_JAVA_OPTS="-Xms256m -Xmx256m" \
  elasticsearch:8.11.0
```

> **注意**：如果 RabbitMQ 需要多个用户（如 `test/test`），启动后执行：
> ```bash
> docker exec zero-rabbitmq rabbitmqctl add_user test test
> docker exec zero-rabbitmq rabbitmqctl set_permissions -p / test '.*' '.*' '.*'
> ```

### 3. 初始化数据库

等待 MySQL 完全启动后（约 10-15 秒），导入所有 SQL 脚本：

```bash
# Linux/macOS
cat script/sql/sys/*.sql script/sql/ums/*.sql script/sql/cms/*.sql \
    script/sql/oms/*.sql script/sql/pms/*.sql script/sql/sms/*.sql \
    script/sql/pub/*.sql | \
    docker exec -i zero-mysql mysql -uroot -p12341qweqfsd2356 gozero
```

```powershell
# Windows (通过 WSL)
wsl bash -c "cat /mnt/c/<项目路径>/script/sql/*/*.sql | docker exec -i zero-mysql mysql -uroot -p12341qweqfsd2356 gozero"
```

### 4. 修改配置文件

各服务的配置文件位于各模块的 `etc/` 目录下，根据实际中间件地址修改：

| 配置文件路径                  | 需要配置的项                              |
|------------------------------|------------------------------------------|
| `rpc/sys/etc/sys.yaml`      | MySQL、Etcd、Redis、JWT                   |
| `rpc/ums/etc/ums.yaml`      | MySQL、Etcd、Redis、MongoDB、RabbitMQ、JWT |
| `rpc/pms/etc/pms.yaml`      | MySQL、Etcd、Redis、MongoDB、RabbitMQ      |
| `rpc/oms/etc/oms.yaml`      | MySQL、Etcd、RabbitMQ                     |
| `rpc/sms/etc/sms.yaml`      | MySQL、Etcd                              |
| `rpc/cms/etc/cms.yaml`      | MySQL、Etcd                              |
| `rpc/search/etc/search.yaml`| Etcd、Elasticsearch                      |
| `api/admin/etc/admin-api.yaml`| Etcd、Redis、JWT                        |
| `api/front/etc/front-api.yaml`| Etcd、Redis、RabbitMQ、JWT、支付宝配置    |
| `job/etc/job-api.yaml`      | Etcd                                     |
| `consumer/etc/consumer-api.yaml`| Etcd、RabbitMQ、JWT                    |

### 5. 编译项目

#### Linux/macOS
```bash
export GOPROXY=https://goproxy.cn,direct
go mod tidy
make build
```

#### Windows (PowerShell)
```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
go mod tidy

# 创建输出目录
$services = @(
    @{name="sys-rpc";    src="./rpc/sys/sys.go";         config="rpc/sys/etc/sys.yaml"},
    @{name="ums-rpc";    src="./rpc/ums/ums.go";         config="rpc/ums/etc/ums.yaml"},
    @{name="oms-rpc";    src="./rpc/oms/oms.go";         config="rpc/oms/etc/oms.yaml"},
    @{name="pms-rpc";    src="./rpc/pms/pms.go";         config="rpc/pms/etc/pms.yaml"},
    @{name="cms-rpc";    src="./rpc/cms/cms.go";         config="rpc/cms/etc/cms.yaml"},
    @{name="sms-rpc";    src="./rpc/sms/sms.go";         config="rpc/sms/etc/sms.yaml"},
    @{name="search-rpc"; src="./rpc/search/search.go";   config="rpc/search/etc/search.yaml"},
    @{name="admin-api";  src="./api/admin/admin.go";     config="api/admin/etc/admin-api.yaml"},
    @{name="front-api";  src="./api/front/front.go";     config="api/front/etc/front-api.yaml"},
    @{name="job";        src="./job/job.go";             config="job/etc/job-api.yaml"},
    @{name="consumer";   src="./consumer/consumer.go";   config="consumer/etc/consumer-api.yaml"}
)

foreach ($s in $services) {
    New-Item -ItemType Directory -Force -Path "target/$($s.name)" | Out-Null
    Copy-Item $s.config "target/$($s.name)/$($s.name).yaml" -Force
    Write-Host "Building $($s.name)..."
    go build -o "target/$($s.name)/$($s.name).exe" $s.src
}
```

### 6. 启动服务

#### Linux/macOS
```bash
make start
```

#### Windows (PowerShell)
```powershell
# 先启动 RPC 服务
$rpcServices = @("sys-rpc","ums-rpc","oms-rpc","pms-rpc","cms-rpc","sms-rpc","search-rpc")
foreach ($name in $rpcServices) {
    Start-Process -FilePath ".\target\$name\$name.exe" `
        -ArgumentList "-f",".\target\$name\$name.yaml" -WindowStyle Hidden
}

# 等待 RPC 服务注册到 Etcd
Start-Sleep -Seconds 5

# 启动 API 和其他服务
$apiServices = @("admin-api","front-api","job","consumer")
foreach ($name in $apiServices) {
    Start-Process -FilePath ".\target\$name\$name.exe" `
        -ArgumentList "-f",".\target\$name\$name.yaml" -WindowStyle Hidden
}
```

### 7. 停止服务

#### Linux/macOS
```bash
make stop
```

#### Windows (PowerShell)
```powershell
$names = @("sys-rpc","ums-rpc","oms-rpc","pms-rpc","cms-rpc","sms-rpc",
           "search-rpc","admin-api","front-api","job","consumer")
foreach ($name in $names) {
    Stop-Process -Name $name -Force -ErrorAction SilentlyContinue
}
```

---

## 验证部署

### 检查进程状态
```powershell
# Windows
Get-Process -Name sys-rpc,ums-rpc,oms-rpc,pms-rpc,cms-rpc,sms-rpc,search-rpc,admin-api,front-api,job,consumer -ErrorAction SilentlyContinue | Format-Table ProcessName, Id -AutoSize
```
```bash
# Linux
ps aux | grep -E "sys-rpc|ums-rpc|oms-rpc|pms-rpc|cms-rpc|sms-rpc|search-rpc|admin-api|front-api|job|consumer"
```

### 测试 admin-api 登录接口

```bash
curl -X POST http://localhost:8888/api/sys/user/login \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"123456"}'
```

预期返回：
```json
{
  "code": "000000",
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### 测试 front-api 登录接口

```bash
curl -X POST http://localhost:9999/api/member/login \
  -H "Content-Type: application/json" \
  -d '{"mobile":"13800138000","password":"123456"}'
```

> 如果返回 `"账号不存在"` 说明服务正常，只是没有该会员数据。

---

## Docker 部署

### 构建镜像
```bash
make image
```

### 启动容器
```bash
make run
```

### Kubernetes 部署
```bash
make kubectl
```

---

## 日志

日志文件位于项目根目录 `logs/` 下，按服务分目录存放：

```
logs/
├── sys/        # sys-rpc 日志
├── ums/        # ums-rpc 日志
├── pms/        # pms-rpc 日志
├── oms/        # oms-rpc 日志
├── sms/        # sms-rpc 日志
├── cms/        # cms-rpc 日志
├── search/     # search-rpc 日志
├── admin/      # admin-api 日志
├── front/      # front-api 日志
├── job/        # job 日志
└── consumer/   # consumer 日志
```

每个目录包含：
- `access.log` — 访问日志
- `error.log` — 错误日志
- `severe.log` — 严重错误日志
- `slow.log` — 慢请求日志

---

## 常见问题

### 1. `bytedance/sonic` 编译报错 `undefined: GoMapIterator`

**原因**：`sonic` 版本与高版本 Go（>= 1.26）不兼容。

**解决**：
```bash
go get github.com/bytedance/sonic@latest
go mod tidy
```

### 2. RabbitMQ 连接报 `403 username or password not allowed`

**原因**：RabbitMQ 没有对应的用户。

**解决**：
```bash
docker exec <rabbitmq容器名> rabbitmqctl add_user guest guest
docker exec <rabbitmq容器名> rabbitmqctl set_permissions -p / guest '.*' '.*' '.*'
```

### 3. MongoDB 认证失败 `AuthenticationFailed`

**原因**：配置文件中的 MongoDB 密码与实际不匹配。

**解决**：修改 `rpc/ums/etc/ums.yaml` 和 `rpc/pms/etc/pms.yaml` 中 `Mongo.Datasource` 的密码。

### 4. 端口冲突

**原因**：其他服务占用了相同端口。

**解决**：
```powershell
# 查看端口占用
netstat -ano | findstr :<端口号>
# 结束占用进程
taskkill /PID <进程ID> /F
```

### 5. Swagger 文档

admin-api 和 front-api 均支持 Swagger 文档，配置文件中 `Swagger.IsTest` 设为 `true` 时启用：
- admin-api: 静态文件路径 `api/admin/static`
- front-api: 静态文件路径 `api/front/static`
