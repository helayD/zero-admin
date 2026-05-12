---
name: flutter-tmux
description: 使用 tmux 启动、热刷新、重启和停止 Flutter 应用。适用于在真机/模拟器上调试 flutter-mall 项目。触发词：flutter run, flutter hot reload, flutter restart, flutter stop, 热刷新, 重启flutter
allowed-tools: Bash, Read
---

# Flutter tmux 管理

通过 tmux session 管理 Flutter debug 进程，支持启动、热刷新、热重启、查看输出和停止。

## 环境

- **项目路径**: `/Users/helay/Documents/GitHub/zero-admin/flutter-mall`
- **tmux session 名**: `flutter`

## 操作命令

### 1. 启动 Flutter

如果 session 不存在则创建，已存在则提示：

```bash
tmux has-session -t flutter 2>/dev/null && echo "flutter session already running" || tmux new-session -d -s flutter "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run"
```

启动后等待 ~20s 编译安装，用步骤 4 检查输出确认是否就绪。

### 2. 热刷新 (Hot Reload)

保留应用状态，仅刷新 widget 树：

```bash
tmux send-keys -t flutter r
```

### 3. 热重启 (Hot Restart)

丢弃应用状态，重新执行 `main()`：

```bash
tmux send-keys -t flutter R
```

### 4. 查看当前输出

```bash
tmux capture-pane -t flutter -p | tail -30
```

### 5. 停止 Flutter 并销毁 session

```bash
tmux send-keys -t flutter q && sleep 2 && tmux kill-session -t flutter 2>/dev/null
```

### 6. 强制重启（先杀再启）

```bash
tmux kill-session -t flutter 2>/dev/null; sleep 1; tmux new-session -d -s flutter "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run"
```

## 快捷键对照

| 按键 | 作用 |
|------|------|
| `r`  | Hot Reload（保留状态） |
| `R`  | Hot Restart（重置状态） |
| `q`  | 退出 flutter run |
| `d`  | Detach（分离，不终止） |

## 注意事项

- 如需指定设备：在 `flutter run` 后加 `-d <device_id>`
- 查看已连接设备：`flutter devices`
- 启动后第一次编译较慢（Gradle assembleDebug ~15s），后续热刷新秒级生效
