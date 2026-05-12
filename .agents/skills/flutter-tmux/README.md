# Flutter Tmux Skill

在 tmux 中启动和管理 Flutter 应用的技能。

## 功能

- 在后台运行 Flutter 应用
- 查看应用日志
- 启动/停止/重启应用
- 查看运行状态

## 使用方法

### 使用 skill

```
/flutter-tmux start   # 启动应用
/flutter-tmux stop    # 停止应用
/flutter-tmux restart # 重启应用
/flutter-tmux status  # 查看状态
/flutter-tmux logs    # 查看日志
```

### 使用脚本

```bash
# 直接运行脚本
./.agents/skills/flutter-tmux/flutter-tmux.sh start
./.agents/skills/flutter-tmux/flutter-tmux.sh stop
./.agents/skills/flutter-tmux/flutter-tmux.sh restart
./.agents/skills/flutter-tmux/flutter-tmux.sh status
./.agents/skills/flutter-tmux/flutter-tmux.sh logs
```

### 使用 tmux 命令

```bash
# 启动
tmux new-session -d -s flutter
tmux send-keys -t flutter "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter

# 查看日志
tmux attach -t flutter

# 停止
tmux send-keys -t flutter "q" Enter
tmux kill-session -t flutter
```

## 文件结构

```
.agents/skills/flutter-tmux/
├── SKILL.md           # 技能说明
├── flutter-tmux.sh    # 启动脚本
└── README.md          # 本文件
```

## 依赖

- tmux
- Flutter SDK
- 连接的设备（Android/iOS/macOS/Chrome）

## 常见问题

### 会话已存在

如果提示 `duplicate session`，使用现有会话：

```bash
tmux send-keys -t flutter "cd /Users/helay/Documents/GitHub/zero-admin/flutter-mall && flutter run" Enter
```

### 应用崩溃

查看日志：

```bash
tmux capture-pane -t flutter -p | tail -100
```

### 强制停止

```bash
tmux kill-session -t flutter
```

## 相关技能

- `flutter` - 直接在前台运行 Flutter
- `flutter-test` - 运行 Flutter 测试
- `flutter-build` - 构建 Flutter 应用
