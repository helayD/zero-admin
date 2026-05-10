# Story 10.11: FISCO BCOS 3.x 真实上链商业化接入（替换免费链 mock 底座）

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 平台运营方,
I want 把 `pkg/fisco` 从「阶段一免费链 mock 底座」升级为真实 FISCO BCOS 3.x 合约调用,
so that 提货卡发放能以零链上费用的方式真正完成链上确权、可在区块浏览器查验、满足监管对「已上链」的合规要求，而不再暴露 buildFreeChainReceipt 这种本地伪造逻辑。

## Acceptance Criteria

1. **Given** FISCO BCOS 3.x 单组网四节点已在 47.107.224.56 上部署完成、`CardToken` ERC721 合约已部署、`pkg/fisco` 已替换为真实 Go SDK 实现  **When** 一张新购买型提货卡走完整发放链路（`ensureOrderPurchaseAssetTx` → `EnsureCardMintTask` → MQ → `ExecuteTask` → `Chain.MintToken`）  **Then** 合约方法 `mint(address,uint256,string)` 会被真实调用并上链，`sms_card_mint_task.chain_tx_id` 落库为真实 64 位 tx hash（`^0x[0-9a-f]{64}$`），`token_id` 落库为合约返回的 uint256 字符串，`chain_status=success`，`mint_status=mint_success`，且可在 FISCO 浏览器 / `curl JSON-RPC getTransactionByHash` 查到同一笔交易。

2. **Given** 真实链调用失败（节点连接超时 / gas 超限 / nonce 冲突 / 合约 revert / 证书认证失败）  **When** 任务执行  **Then** 错误必须被映射到既有 `ErrorCodeMintExecuteFailed` 错误体系的明确子类（`node_unreachable` / `tx_timeout` / `contract_revert` / `tls_handshake_failed`），`last_error_reason` 写入人类可读摘要（脱敏）、`retry_count` 累加、`next_retry_at` 按指数退避、达到 `max_retry_count` 升级 `manual_review`，**绝不能出现"免费链能力未启用"或其它 mock 时代的错误串**。

3. **Given** 同一 `idempotency_key` 的任务被 MQ 重复投递 / job 补偿重复捡 / 管理员重复点 Retry  **When** 任意一路触发链上铸造  **Then** 合约级只会产生一笔有效 mint（通过合约 `mapping(bytes32 => uint256) mintedByIdempotencyKey` 或 off-chain `QueryMintToken(key)` 先查后写）、`sms_card_mint_task.chain_tx_id` 不会被覆盖为第二笔、`sms_card_instance.token_id` 唯一。

4. **Given** 当前 3 张遗留卡 `asset_instance_id in (970012, 970016, 970017)` 处于 `manual_review` + 历史错误码 `receipt_reconcile_required: 免费链能力未启用` / `mint_prerequisite_rejected`  **When** Story 10.11 上线后执行清理步骤  **Then** 这 3 张卡被重置为 `pending_dispatch`、自动扫描重派、走真实链路 mint 成功，C 端提货卡详情页不再出现「补偿中 / asset_created / 免费链 / 待发放」等内部术语，时间线显示「已到账」。

5. **Given** FISCO 节点暂时不可达（生产事故 / 运维窗口）  **When** 新订单生成并创建发放任务  **Then** 系统不会因链不可达而阻塞订单主链路（订单仍然创建、`sms_card_instance` 仍然建账），任务表 `task_status=pending_dispatch`、`mint_status=mint_compensating`、由 sms-rpc 自循环扫描或恢复后的 job 自动重新派发，链恢复后自动追发成功。

6. **Given** sms-rpc / consumer / job 三个服务都需要持有同一条 FISCO 链客户端  **When** 任一服务启动  **Then** 根据 `Blockchain.Primary: fisco` 读取相同的 `Fisco.*` 配置块（含 `CaCert/SdkCert/SdkKey` 证书路径）构造同一种 `chainclient.ChainClient` 实现，日志首行打印 `buildChainClient: chainType="fisco_bcos_3x" nodeAddr=...`，与 mock 时代的 `chainType()="free_chain"` 明确区分。

## AC 覆盖矩阵

- **AC1（真实 SDK 集成 + 真实 mint + 真实回执）**：Task 2.1-2.4、3.1-3.5、4.1-4.6、5.1-5.4、10.1-10.3
- **AC2（错误映射 + 重试补偿 + 禁 mock 错误串）**：Task 4.7-4.9、5.5-5.8、8.1-8.4
- **AC3（三层幂等 + 合约级去重 + 回执先查后写）**：Task 3.6-3.7、5.9-5.10、9.1-9.3
- **AC4（历史 3 张卡闭环 + C 端文案脱敏）**：Task 11.1-11.5
- **AC5（节点不可达时订单主链路不阻塞 + 自循环追发）**：Task 5.11-5.13、10.4-10.5
- **AC6（三服务配置统一 + 启动日志可观测）**：Task 1.4、4.10、6.1-6.4

## 背景与关键风险（必须优先阅读）

> **⚠️ CRITICAL - 本 Story 必须复用 Story 10.4 已经落地的「任务表 + consumer + job + 三层幂等 + 补偿治理」框架，绝不能另起一套执行器。** `pkg/digitalcardmint/service.go` 的 `ExecuteTask / DispatchTask / ScanDueTasks / RetryTask / FreezeTask / EscalateTask` 已是唯一执行器，10.11 只允许替换 `ChainClient` 实现（`pkg/fisco/client.go`），不允许绕开 `digitalcardmint.Service` 去直接调 FISCO SDK。
>
> **⚠️ CRITICAL - `tech-spec-multi-chain-adapter-refactor.md` 明确把「FISCO BCOS 节点部署和运维」划为 Out of Scope。** 本 Story 正是为了补齐这一部分，必须把节点部署、合约部署、证书分发、Go SDK 接入、灰度切换完整落地，而不是再留一个「阶段三」尾巴。
>
> **⚠️ CRITICAL - `docs/blockchain-strategy.md` 已经决策 FISCO BCOS 作为「商业化首发基础存证底座」，不是临时选型。** 因此本 Story 不能把 FISCO 当作可选方案做一半，也不能在代码里保留「如果 FISCO 不通就回落到 mock」这类 fallback 逻辑（mock 只能存在于单元测试）。
>
> **⚠️ CRITICAL - C 端监管约束（AGENTS.md 已固化）：Flutter 严禁展示 chainType / chainStatus / chainTxId / tokenId / 上链 / 链路 等字段或关键词。** 本 Story 在 C 端不能因为"真实上链了"就放开这些字段；`mint_status_text` 等允许字段若包含底层链路文案，Flutter 必须做二次脱敏转为「发放/到账/处理结果」口径。B 端（web-admin）可见完整链信息不受此约束。
>
> **⚠️ CRITICAL - 幂等键 `idempotency_key` 必须同时做到合约级和 off-chain 级。** 合约内用 `mapping(bytes32 keccak256(idempotencyKey) => uint256 tokenId)` 去重；off-chain 在 `ExecuteTask` 调链前先 `QueryMintToken(idempotencyKey)`，有回执直接走 `transitionSuccess` 不再发交易。缺任一层，在"MQ 重放 + 管理员 Retry + job 扫描"三路竞争下都会出现重复 mint。
>
> **⚠️ CRITICAL - FISCO BCOS 3.x 默认开启 SSL 双向认证（TLS 1.3）。** `ca.crt / sdk.crt / sdk.key` 三个证书文件必须通过受控渠道分发到三台服务器（sms-rpc / consumer / job 共用），禁止把证书内容硬编码到 yaml；yaml 只存路径。证书文件权限必须 `0400`。
>
> **⚠️ CRITICAL - 交易回执不是"发出去就成功"。** FISCO 3.x `sendTransaction` 返回 tx hash 不代表上链成功，必须后续 `getTransactionReceipt(hash)` 轮询到 `status=0x0` 才算真正成功。必须在 `fisco.client.MintToken` 里实现"发 tx → 轮询 receipt → 解析 event 日志提取 tokenId"完整三步，不能只取 tx hash 就返回。
>
> **⚠️ CRITICAL - 当前仓库 MySQL `sms_card_mint_task.token_id` 是 `VARCHAR(128)`，合约 tokenId 是 `uint256` 最大 78 位十进制、32 字节 hex。** 必须确认 128 字节足够（78 < 128 ✅），同时约定 token_id 落库格式为"十进制字符串"（方便浏览器直接查询），避免 hex/dec 混用。合约 event 解析也按十进制字符串输出。

## 明确非目标（本 Story 不做）

- 不做至信链 / 蚂蚁链 / BSN 的并行双写；本 Story 只完成 FISCO BCOS 3.x 真实接入，双写在未来的 Story 10.12+ 规划。
- 不做合约的业务功能扩展（转赠 / 销毁 / 抵押 / 分片），只做最小商业化 mint + transfer + getToken 三个接口满足发放需求；转赠等能力由 Story 10.7 在合约层再扩展。
- 不做 FISCO BCOS 节点的高可用集群（跨机房 / 跨可用区）；首发阶段就 47.107.224.56 单机 4 节点，满足商业化首发 QPS（≤10 mint/s）。
- 不替换 `pkg/mq/rabbitmq_util.go` 的 MQ 实现，也不改 `sms_card_mint_task` 表结构；只新增字段若必须，须通过 `script/sql/` 迁移脚本走。
- 不改动 Flutter 端任何链相关展示逻辑；C 端文案脱敏（AC4）通过修改 api/front 返回字段实现，不让客户端直接拿到链字段。
- 不升级 go-zero / GORM / RabbitMQ 客户端；本 Story 保持现有版本矩阵。
- 不在本 Story 引入新的监控栈；已有 Prometheus agent 保持现状，FISCO 节点监控在 Story 10.12 再补。

## 实施顺序（必须按序推进）

### Phase 1 — 基础设施准备（节点 + 合约，不改代码）

#### Task 1. 部署 FISCO BCOS 3.x 节点

- [ ] 1.1 在 47.107.224.56 准备目录 `/opt/fisco`，用户 `fisco`，日志轮转策略
- [ ] 1.2 下载官方 `build_chain.sh`（对应 FISCO BCOS 3.8.0 或当时 latest-stable），生成 1 组网 4 节点（Port 20200-20203 RPC, 20300-20303 P2P）
- [ ] 1.3 `./build_chain.sh -l "127.0.0.1:4" -p 30300,20200,8545 -T` 启动，验证 `curl -X POST http://127.0.0.1:20200 -d '{"jsonrpc":"2.0","method":"getGroupList","params":[],"id":1}'` 返回 `["group0"]`
- [ ] 1.4 生成并收集 SDK 证书：`nodes/127.0.0.1/sdk/{ca.crt,sdk.crt,sdk.key}`，权限 `0400`，路径规划为 `/opt/fisco/sdk/`
- [ ] 1.5 配置 systemd 开机自启（`fisco-bcos-node0.service` 到 `node3.service`）

#### Task 2. 编写并部署 CardToken 合约

- [ ] 2.1 新建 `contracts/CardToken.sol`（FISCO BCOS 3.x Solidity 0.8.x 兼容），实现：
  - ERC721 base（`ownerOf / tokenURI / transferFrom / approve`）
  - `mint(address to, string calldata idempotencyKey, string calldata metadataURI) external returns (uint256 tokenId)` — 仅 `MINTER_ROLE` 可调
  - `mapping(bytes32 => uint256) private _mintedByKey` — `mint()` 首行 `require(_mintedByKey[keccak256(bytes(idempotencyKey))] == 0, "duplicate mint")`
  - `getTokenByKey(string calldata idempotencyKey) external view returns (uint256)` — 支持 off-chain 先查后写
  - `event CardMinted(address indexed to, uint256 indexed tokenId, string idempotencyKey, string metadataURI)` — 供 SDK 解析 tokenId
- [ ] 2.2 单元测试（Foundry 或 Remix）覆盖：首次 mint / 重复 idempotencyKey revert / 非 MINTER_ROLE revert / getTokenByKey 返回正确值
- [ ] 2.3 使用 `console2.sh` 或 Java SDK 部署合约到 group0，记录合约地址 `CONTRACT_ADDR`
- [ ] 2.4 导出 `CardToken.abi` 到 `/opt/fisco/contracts/CardToken.abi`，提交一份副本到仓库 `pkg/fisco/contracts/CardToken.abi`（去除敏感信息）
- [ ] 2.5 grant `MINTER_ROLE` 给 sms-rpc 运行时私钥对应地址，记录 txhash

#### Task 3. 私钥与证书分发

- [ ] 3.1 为 sms-rpc 生成专用以太坊兼容私钥（`secp256k1`），写入 `/opt/fisco/keys/sms-rpc.key` 权限 `0400` 归属 `root`
- [ ] 3.2 把该私钥对应地址授予 `MINTER_ROLE`（Task 2.5）
- [ ] 3.3 consumer / job 若需独立签名（当前设计不需要，execute 集中在 sms-rpc）则也生成专用私钥；否则复用 sms-rpc 的私钥通过受控部署
- [ ] 3.4 制定「私钥轮换 SOP」草稿写入 `docs/blockchain-strategy.md` Appendix C
- [ ] 3.5 确认当前 Config 加载方式（明文 yaml）在生产环境可以接受（非金融级私钥，无大额资金），后续若要升级到 KMS / Vault 另起 Story

### Phase 2 — Go SDK 集成（核心代码改动）

#### Task 4. 依赖与 Config 扩展

- [ ] 4.1 `go get github.com/FISCO-BCOS/go-sdk/v3@latest`（确认兼容 3.x）；更新 `go.mod` / `go.sum`
- [ ] 4.2 `rpc/sms/internal/config/config.go` 的 `Fisco struct` 扩展字段：
  ```go
  Fisco struct {
      NodeAddr       string // 例: 127.0.0.1:20200
      GroupID        string // FISCO 3.x 是字符串: "group0"
      ChainID        string // "chain0"
      ContractAddr   string // 0x...
      ContractABI    string // 内联 JSON 或相对路径
      PrivateKey     string // 私钥（或 path:// 前缀指向文件）
      CaCertPath     string // /opt/fisco/sdk/ca.crt
      SdkCertPath    string // /opt/fisco/sdk/sdk.crt
      SdkKeyPath     string // /opt/fisco/sdk/sdk.key
      TimeoutSeconds int64
      ReceiptPoll    struct {
          IntervalMs int64 // 轮询间隔，默认 500ms
          MaxAttempts int  // 最大轮询次数，默认 30（15 秒）
      }
      Enabled bool
  }
  ```
- [ ] 4.3 consumer / job 的 `config.go` 同步扩展（保持三服务 Fisco 字段完全一致）
- [ ] 4.4 `sms-rpc.yaml / consumer-api.yaml / job-api.yaml` 示例更新：把 `Fisco.*` 字段占位符补全（真实值由运维 overlay 注入，不提交到 git）
- [ ] 4.5 `script/` 下新增 `fisco-overlay.yaml.example` 示范生产值填法

#### Task 5. 重写 `pkg/fisco/client.go` 为真实实现

- [ ] 5.1 新建 `pkg/fisco/client_real.go`：`fiscoRealClient struct` 持有 `*sdk.Client`、合约地址、合约 ABI（预解析）、私钥、超时配置
- [ ] 5.2 `NewClient(cfg Config)` 根据 `cfg.Enabled + cfg.NodeAddr != ""` 选择：配置完整 → `fiscoRealClient`；否则 → `disabledClient`（保留既有 disabled 语义，但错误文案改为 "FISCO BCOS 客户端未配置"，不要再出现「免费链」字样）
- [ ] 5.3 删除 `pkg/fisco/client.go` 中所有 `fiscoClient` / `receipts map` / `buildFreeChainReceipt` / `shortDigest` / `cloneMintTokenResponse` mock 逻辑；保留 `disabledClient`（错误文案更新）
- [ ] 5.4 `fiscoRealClient.MintToken(ctx, req)` 实现完整流程：
  1. 组装 metadata JSON：`{"assetNo":...,"activityId":...,"templateId":...,"memberId":...,"scope":...,"traceId":...}`
  2. 用私钥签名构造 tx：`contract.Mint(opts, toAddress, req.IdempotencyKey, metadataJSON)`
  3. 发送 tx，拿到 tx hash
  4. 启动 receipt 轮询循环（`IntervalMs / MaxAttempts` 从 config 取）
  5. 解析 receipt：`status == 0x0` 才算成功；`status != 0x0` 映射为 `contract_revert` 错误
  6. 从 receipt.Logs 中解析 `CardMinted(address,uint256,string,string)` event，提取 tokenId（uint256 → decimal string）
  7. 构造 `chainclient.MintTokenResponse`：`TokenID=tokenIdDec`, `ChainTxID=txHash`, `ChainStatus="success"`, `ReceiptSummary=fmt.Sprintf("block=%d gasUsed=%d",...)`, `ReceiptJSON=` 完整 receipt JSON，`ConfirmedAt=time.Now()`
- [ ] 5.5 `fiscoRealClient.QueryMintToken(ctx, req)` 实现：优先调用合约 `getTokenByKey(idempotencyKey)`；tokenId==0 → 返回 `chainclient.ErrReceiptNotFound`；否则用 `GetTransactionByHashInMap` 回查 tx hash（或记录在本地 `sms_chain_receipt_cache` 表，下一 Story 再做缓存，本 Story 先每次查链）构造 `MintTokenResponse`
- [ ] 5.6 `ChainType()` 返回 `"fisco_bcos_3x"`（不要 `free_chain`），让上游日志一眼能区分真假链
- [ ] 5.7 新增 `fisco/errors.go`：定义错误映射函数 `classifyFiscoError(err error) (code string, reason string)`：
  - `connection refused / i/o timeout` → `node_unreachable`
  - `tls: handshake failure` → `tls_handshake_failed`
  - `gas required exceeds allowance` → `gas_insufficient`
  - `nonce too low` / `nonce too high` → `nonce_conflict`
  - `execution reverted: duplicate mint` → `duplicate_mint_rejected`（AC3 幂等信号，上游直接走 QueryMintToken fallback）
  - `execution reverted: ...` 其他 → `contract_revert`
  - receipt 轮询超时 → `tx_timeout`
- [ ] 5.8 `pkg/digitalcardmint/service.go` 的 `markExecuteFailure` 链路使用 `classifyFiscoError` 返回的 code 写入 `last_error_code`；AC2 中要求的「绝不出现免费链错误串」通过 lint 脚本保障（Task 8.4）
- [ ] 5.9 `pkg/fisco/client_real_test.go` 用 `ginkgo + gomega` 或 `testify` 配合 fake FISCO JSON-RPC server 覆盖：mint 成功 / duplicate mint / tls error / tx timeout 四个分支
- [ ] 5.10 旧的 `pkg/fisco/client_test.go` 若包含 `TestNewClientDisabled` 保留；包含 mock fisco 行为的测试全部删除或改为 real-client 覆盖
- [ ] 5.11 验证链不可达时 `ExecuteTask` 不会 panic，只会走 `markExecuteFailure` + `next_retry_at` + 返回错误（AC5）
- [ ] 5.12 验证 sms-rpc 在 FISCO 不可达时仍能正常启动（`NewClient` 即使连不上节点也不应 panic，链健康检查延后到首次调用；failing fast 风格不适合我们的异步场景）
- [ ] 5.13 `/health` 或 Prometheus 指标暴露 FISCO 连接状态（可选，本 Story 仅要求日志层可见）

#### Task 6. sms-rpc / consumer / job ServiceContext 接线

- [ ] 6.1 `rpc/sms/internal/svc/service_context.go.buildChainClient` 在 `case "fisco"` 分支调用新的 `fisco.NewClient(fisco.Config{...CaCertPath/SdkCertPath/SdkKeyPath/ABI 完整字段...})`
- [ ] 6.2 启动日志必须打印 `buildChainClient: chainType="fisco_bcos_3x" nodeAddr=... contract=0x...`（AC6）
- [ ] 6.3 `consumer/internal/svc/service_context.go` 同步改造
- [ ] 6.4 `job/internal/svc/service_context.go` 同步改造
- [ ] 6.5 三服务一致性测试：任一服务启动都能产生同一条 `ChainClient` 实现，链调用结果一致

### Phase 3 — 质量保证 + 上线治理

#### Task 7. 幂等与回执缓存

- [ ] 7.1 `digitalcardmint.Service.ExecuteTask` 调链前先调 `Chain.QueryMintToken(ctx, {IdempotencyKey})`，有回执直接走 `transitionSuccess`，跳过 `MintToken`
- [ ] 7.2 Off-chain 幂等与 on-chain `_mintedByKey` 双重保障
- [ ] 7.3 `last_receipt_json` 大小上限 4KB，超出截断（FISCO receipt 平均 1~2KB，安全）
- [ ] 7.4 `sms_card_mint_task.chain_tx_id` / `token_id` 一旦落库为非空值，后续 `UPDATE` 不得覆盖（SQL 层加 `AND (chain_tx_id = '' OR chain_tx_id IS NULL)` 条件）
- [ ] 7.5 单测覆盖：同一 IdempotencyKey 两次 ExecuteTask 只产生一笔 mint，第二次直接走 QueryMintToken 返回已有 receipt

#### Task 8. 安全与脱敏

- [ ] 8.1 `last_error_reason` 写入数据库前必须脱敏：屏蔽私钥片段（32 字节 hex）、TLS 证书片段（含 BEGIN/END 关键字的整段）
- [ ] 8.2 `api/front` 层 `OrderDigitalCardItem` 的 `mint_status_text` 若后端返回包含「链 / 上链 / chain / token / mint / receipt / FISCO / BCOS」关键词一律替换为「处理中/到账/待到账」口径（AC4）
- [ ] 8.3 Flutter 层 `digital_card_detail.dart` 做第二层保险：即使 api 层漏了，UI 再过一次关键词黑名单（复用 Story 10.10 已定义的 `chainKeywordBlacklist`）
- [ ] 8.4 CI 新增 `scripts/check-no-free-chain.sh`：`grep -rn "免费链能力未启用\|free_chain\|buildFreeChainReceipt\|mock fisco" pkg/ rpc/ api/ --include="*.go"`，非零退出即 fail，防回退

#### Task 9. 灰度切换与回滚

- [ ] 9.1 灰度阶段 1：只在 sms-rpc 开 Fisco，consumer/job 保持原状；验证 1 个新订单 → 1 张卡 mint 成功
- [ ] 9.2 灰度阶段 2：consumer + job 同步切换；验证自循环扫描 + 超时补偿走真链
- [ ] 9.3 灰度阶段 3：SQL 重置 970012/970016/970017 三张遗留卡为 pending_dispatch、清 last_error / last_execute_at / token_id / chain_tx_id，让自循环扫描自动重派，走真链 mint 成功（AC4）
- [ ] 9.4 回滚脚本 `script/rollback-fisco-to-disabled.sh`：把 `Fisco.Enabled=false` 并重启三服务，新任务会进入 pending_dispatch 等 FISCO 恢复，不丢数据
- [ ] 9.5 运维手册 `docs/fisco-operations.md` 覆盖：节点重启 / 证书轮换 / 合约升级 / 监控告警 / 日常查询

#### Task 10. 端到端验证

- [ ] 10.1 本地冒烟：`make build && go test ./pkg/fisco/... -count=1 -v`
- [ ] 10.2 远程冒烟：部署到 47.107.224.56 → 手机下单 → 观察 sms-rpc 日志出现 `mint tx hash=0x...` → 在 FISCO JSON-RPC 查 txhash 确认链上存在 → C 端显示「已到账」
- [ ] 10.3 幂等验证：手动 `grpcurl` 连续调 `CardMintAdminService/RetryTask` 5 次同一 taskID，确认只出现 1 笔 tx，后续 4 次都短路到 QueryMintToken
- [ ] 10.4 节点宕机验证：`systemctl stop fisco-bcos-node0` 后下单，观察任务停在 pending_dispatch / 错误码 `node_unreachable`；启回节点观察自循环扫描自动追发成功
- [ ] 10.5 三张遗留卡闭环（AC4）：执行 Task 9.3 的 SQL + 观察 30 秒后自循环扫描日志，确认三张卡从 manual_review 走向 mint_success，手机端刷新提货卡详情页确认不再出现 asset_created / 补偿中 / 待发放 文案，时间线只显示「已到账」

#### Task 11. 现存 3 张卡闭环（AC4 落地专项）

- [ ] 11.1 SQL 重置：
  ```sql
  UPDATE sms_card_mint_task
     SET task_status='pending_dispatch', mint_status='mint_compensating',
         manual_required=0, frozen=0, retry_count=0,
         last_error_code='', last_error_reason='',
         last_execute_at=NULL, next_retry_at=NULL,
         token_id='', chain_tx_id=''
   WHERE asset_instance_id IN (970012, 970016, 970017);
  ```
- [ ] 11.2 确认 sms_card_instance 对应记录 `mint_status` 也从 `mint_manual_review` 改回 `mint_compensating`（避免 UI 锁死）
- [ ] 11.3 等 sms-rpc 自循环扫描（Story 10.10 已落地的 startMintTaskSelfScan，每 30s 一次）捡到并真链 mint
- [ ] 11.4 验证 sms_card_mint_task 三条记录 `token_id/chain_tx_id` 落到真实值、`chain_status='success'`
- [ ] 11.5 手机端打开这三张卡详情页，截屏确认时间线只出现产品语：收到订单 / 已支付 / 已到账（或对应脱敏口径）

## 技术约束与依赖

- **语言 & 框架**：Go 1.25、go-zero 1.9.3、GORM、gRPC（沿用）
- **新增依赖**：`github.com/FISCO-BCOS/go-sdk/v3` 3.x 兼容版
- **合约工具链**：Solidity 0.8.x、Foundry（或 Remix）、FISCO 官方 `console2` 部署工具
- **基础设施**：47.107.224.56 服务器、FISCO BCOS 3.x 单组网四节点、单条 group0、Port 20200-20203
- **数据库**：不新增表、不新增字段；仅复用 `sms_card_mint_task.token_id / chain_tx_id / chain_status / last_receipt_*`
- **并行 Story**：本 Story 与 Story 10.7（提货卡转赠）合约层有交集，若 10.7 先上线，本 Story 合约实现需包含 `transferFrom` 适配（ERC721 自带，天然兼容）
- **依赖 Story**：10.3（卡片实例）、10.4（任务表与补偿）、multi-chain-adapter-refactor（接口抽象）。这三项必须已完成且稳定运行，本 Story 才能启动。
- **阻塞 Story**：10.5 资产审计检索需要真实 chain_tx_id 做跳转查证，本 Story 不完成 10.5 的链上跳转能力会缺失。

## 验证策略

- **单元测试**：`pkg/fisco/client_real_test.go` 全分支覆盖（Task 5.9）
- **集成测试**：本地起一个 FISCO 3.x 节点（docker-compose），全链路跑一次 `ExecuteTask → MintToken → receipt → transitionSuccess`
- **冒烟测试**：部署到 47.107.224.56 手机端端到端（Task 10.2、11.5）
- **回归测试**：确保 Story 10.4 / 10.10 的既有能力（`RetryTask / FreezeTask / ScanDueTasks / startMintTaskSelfScan`）全部通过

## 不确定点（预留问题清单）

- [ ] Q1: `github.com/FISCO-BCOS/go-sdk/v3` 对 FISCO 3.8 的兼容性是否稳定？需要 spike 1 天验证。
- [ ] Q2: 合约部署用 Java console2 还是 Go SDK 部署？本 Story 倾向 console2，便于运维复用；但最终 ABI 还需 Go SDK 能解析。
- [ ] Q3: tokenId 十进制字符串 vs hex，落库 128 字节够用但查询体验？倾向十进制字符串（FISCO 浏览器默认十进制）。
- [ ] Q4: `_mintedByKey` 合约 storage 膨胀问题，长期千万级 mint 后 gas 成本？单链首发问题不大，Story 10.13+ 考虑分片或合约升级。

## Related

- Epic 10（提货卡产品线）
- Story 10.3（卡片实例建账）
- Story 10.4（蚂蚁链 token 发放与补偿 —— 本 Story 的执行框架父级）
- Story 10.5（资产展示与审计检索 —— 本 Story 提供 chain_tx_id 来源）
- Story 10.7（提货卡转赠 —— 未来合约层扩展）
- Story 10.10（提货卡后台配置中心 / 商业化版）
- Tech Spec: `_opcos/implementation-artifacts/tech-spec-multi-chain-adapter-refactor.md`
- 战略文档: `docs/blockchain-strategy.md`

## 变更日志

- 2026-05-10：初稿（Cascade）—— 背景：Story 10.4 已落地蚂蚁链真实 HTTP 接入，但 `pkg/fisco` 仍是「阶段一免费链 mock 底座」；客户不接受 mock + 蚂蚁链年费 12w+ 过高，决策走 FISCO BCOS 免费链商业化真实接入。本 Story 补齐 tech-spec-multi-chain-adapter-refactor 明确 Out of Scope 的「FISCO 节点部署 + 真实 SDK 接入」部分。
