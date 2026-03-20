# Story 1.1 基线初始化验收记录

## 摘要

- Story：`1-1-既有仓库基线初始化`
- 日期：`2026-03-20`
- 结果：已在既有 brownfield 仓库内完成基线准备，未创建任何平行 scaffold

## Brownfield 基线盘点

- 已确认根级实现单元继续沿用：`api/`、`rpc/`、`consumer/`、`job/`、`web-admin/`、`flutter-mall/`、`pkg/`、`docs/`、`script/`
- 本次 story 未引入新的 admin/front/mobile 平行工程目录
- 补充的初始化说明仅写入既有 `_opcos/implementation-artifacts/`

## 已修复脚本缺陷

1. 从 [`makefile`](../../makefile) 与 [`service_manager.sh`](../../service_manager.sh) 中移除了不存在的 `target/web-api/web-api` 启动入口
2. 将 [`service_manager.sh`](../../service_manager.sh) 中 `"pms"` 分支修正为调用 `pms_service`
3. 将 [`makefile`](../../makefile) 中的 `goctl` 安装固定为 `v1.9.2`，避免后续 `make gen` 生成物相对仓库基线漂移

## 验证命令

| 命令 | 结果 | 说明 |
| --- | --- | --- |
| `bash script/shell/verify_baseline_init.sh` | 通过 | 新增基线 smoke check，覆盖根级目录、`web-api` 假入口、`pms_service` 分派与 `goctl` 固定版本 |
| `bash -n service_manager.sh` | 通过 | 修复后脚本语法正常 |
| `make -n start` | 通过 | 启动命令列表已不再引用 `web-api` |
| `make deps` | 通过 | `go mod tidy -v` 执行完成，未留下跟踪依赖 diff |
| `make build` | 通过 | 既有后端构建链完成，RPC、网关、job、consumer 二进制均成功生成 |
| `cd web-admin && npm install` | 通过（补充兼容配置后） | 在 npm `11.6.2` 下需要 `web-admin/.npmrc` 中的 `legacy-peer-deps=true` 才能兼容 React 17 依赖树 |
| `cd flutter-mall && flutter pub get` | 通过 | 基于现有 lockfile 的依赖恢复完成 |
| `cd web-admin && npm run lint` | 失败 | 现有代码库仍有 `32 errors`、`461 warnings`，不属于 Story 1.1 改动引入 |
| `cd flutter-mall && flutter analyze` | 失败 | 现有代码库仍有 `17 issues`，以 info/warning 为主，不属于 Story 1.1 改动引入 |

## 残余基线阻塞

- `web-admin` lint 仍存在既有问题，例如：
  - `src/pages/pms/ProductSku/data.d.ts:14` 解析错误
  - `src/components/common/UploadFileComponents.tsx:58` 变量遮蔽
  - `src/pages/system/*` 下的类型空格与变量遮蔽问题
- `flutter-mall` analyze 仍存在既有问题，例如：
  - `lib/config/service_url.dart:1` dangling library doc comment
  - `lib/utils/http_util.dart:108` / `126` / `139` 的 `rethrow` 建议
  - `lib/welcome.dart:2` 未使用导入与异步 `BuildContext` 告警
- `make` / `go build` 过程中仍会出现 `ld: warning: ignoring duplicate libraries: '-ltaos'`，但本 story 中未阻断依赖准备与产物生成

## 生成链与边界确认

- 未手改任何生成网关路由、protobuf 文件或 `rpc/*/gen/query/*` 产物
- 生成源边界保持不变：
  - API 定义：`api/*/doc/api/**/*.api`
  - RPC 定义：`rpc/*/*.proto`、`rpc/*/proto/*.proto`
  - GORM 生成入口：`rpc/*/gen/generator.go`
- 后续 story 仍应按 Dev Notes 中约定，继续落在既有 backend、web-admin 与 flutter-mall 结构内扩展

## 结论

- Story 1.1 已完成既有 brownfield 仓库的基线初始化准备
- 启动脚本中的关键阻断项已经清除
- 后端、Web Admin 安装链路、Flutter 依赖恢复链路均已在现有仓库布局内可执行
- 剩余前端 / Flutter 质量问题应作为后续清理项处理，不建议在本基线 story 中继续扩写范围
