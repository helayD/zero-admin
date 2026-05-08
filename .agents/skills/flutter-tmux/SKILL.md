# Flutter Tmux Skill

在 tmux 中启动和管理 Flutter 应用。

## 使用场景

当需要在后台运行 Flutter 应用时使用此技能。

## 命令

### 启动 Flutter 应用

```bash
# 创建新的 tmux 会话
tmux new-session -d -s flutter-run

# 进入 Flutter 项目目录并启动应用
tmux send-keys -t flutter-run "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter
```

### 查看 Flutter 日志

```bash
# 查看 tmux 会话输出
tmux capture-pane -t flutter-run -p | tail -50
```

### 停止 Flutter 应用

```bash
# 发送 'q' 停止应用
tmux send-keys -t flutter-run "q" Enter

# 或者直接杀死会话
tmux kill-session -t flutter-run
```

### 重新启动 Flutter 应用

```bash
# 先停止
tmux send-keys -t flutter-run "q" Enter
sleep 2

# 再启动
tmux send-keys -t flutter-run "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter
```

## 设备选择

启动后会提示选择设备：
- `1` - Android 设备
- `2` - iOS 设备
- `3` - macOS 桌面
- `4` - Chrome 浏览器

## 常见问题

### 会话已存在

如果提示 `duplicate session`，使用现有会话：

```bash
tmux send-keys -t flutter-run "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter
```

### 应用崩溃

查看日志：

```bash
tmux capture-pane -t flutter-run -p | tail -100
```

### 强制停止

```bash
tmux kill-session -t flutter-run
```

## 快捷命令

```bash
# 一键启动（如果会话不存在）
tmux new-session -d -s flutter-run 2>/dev/null || true
tmux send-keys -t flutter-run "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter

# 一键停止
tmux kill-session -t flutter-run 2>/dev/null || true

# 一键重启
tmux kill-session -t flutter-run 2>/dev/null || true
sleep 1
tmux new-session -d -s flutter-run
tmux send-keys -t flutter-run "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter
```

## 相关技能

- `flutter-run` - 直接在前台运行 Flutter
- `flutter-test` - 运行 Flutter 测试
- `flutter-build` - 构建 Flutter 应用
