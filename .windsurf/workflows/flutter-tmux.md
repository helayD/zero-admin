---
description: 使用 tmux 启动、热刷新、重启和停止 Flutter 应用。适用于在真机/模拟器上调试 flutter-mall 项目。触发词：flutter run, flutter hot reload, flutter restart, flutter stop, 热刷新, 重启flutter
---

# Flutter tmux 管理

通过 tmux session 管理 Flutter debug 进程，支持热刷新、重启、停止等操作。

## 环境

- **项目路径**: `/Users/helay/Documents/GitHub/zero-admin/flutter-mall`
- **tmux session 名**: `flutter`
- **Flutter SDK**: `/Users/helay/SDKs/flutter_3.41.5_stable`

## 命令

### 1. 启动 Flutter（如果 session 不存在则创建）

// turbo
```bash
tmux has-session -t flutter 2>/dev/null && echo "session already exists" || tmux new-session -d -s flutter "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run"
```

启动后等待 ~20s 编译安装完成，用以下命令检查状态：

// turbo
```bash
sleep 5 && tmux capture-pane -t flutter -p | tail -15
```

### 2. 热刷新 (Hot Reload)

// turbo
```bash
tmux send-keys -t flutter r
```

### 3. 热重启 (Hot Restart)

// turbo
```bash
tmux send-keys -t flutter R
```

### 4. 查看当前输出

// turbo
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

## 注意事项

- `r` = hot reload（保留状态，仅刷新 widget 树）
- `R` = hot restart（丢弃状态，重新执行 main()）
- `q` = 退出 flutter run
- 如需指定设备：在 `flutter run` 后加 `-d <device_id>`
- 查看已连接设备：`flutter devices`
