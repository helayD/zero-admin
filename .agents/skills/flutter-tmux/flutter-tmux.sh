#!/bin/bash
# Flutter Tmux 启动脚本
# 用法: ./flutter-tmux.sh [start|stop|restart|status|logs]

FLUTTER_DIR="/Users/helay/Documents/GitHub/zero-admin/flutter-mall"
SESSION_NAME="flutter-run"

case "$1" in
    start)
        # 启动 Flutter 应用
        tmux new-session -d -s $SESSION_NAME 2>/dev/null || true
        tmux send-keys -t $SESSION_NAME "cd $FLUTTER_DIR && flutter run" Enter
        echo "✅ Flutter 应用已在 tmux 会话 '$SESSION_NAME' 中启动"
        echo "📱 查看日志: tmux attach -t $SESSION_NAME"
        echo "🛑 停止应用: $0 stop"
        ;;
    stop)
        # 停止 Flutter 应用
        tmux send-keys -t $SESSION_NAME "q" Enter 2>/dev/null || true
        sleep 2
        tmux kill-session -t $SESSION_NAME 2>/dev/null || true
        echo "✅ Flutter 应用已停止"
        ;;
    restart)
        # 重启 Flutter 应用
        $0 stop
        sleep 1
        $0 start
        ;;
    status)
        # 查看状态
        if tmux has-session -t $SESSION_NAME 2>/dev/null; then
            echo "✅ Flutter 应用正在运行"
            echo "📱 会话: $SESSION_NAME"
            echo "📋 查看日志: $0 logs"
            echo "🛑 停止应用: $0 stop"
        else
            echo "❌ Flutter 应用未运行"
            echo "🚀 启动应用: $0 start"
        fi
        ;;
    logs)
        # 查看日志
        if tmux has-session -t $SESSION_NAME 2>/dev/null; then
            tmux capture-pane -t $SESSION_NAME -p | tail -50
        else
            echo "❌ Flutter 应用未运行"
        fi
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "命令说明:"
        echo "  start   - 启动 Flutter 应用"
        echo "  stop    - 停止 Flutter 应用"
        echo "  restart - 重启 Flutter 应用"
        echo "  status  - 查看运行状态"
        echo "  logs    - 查看应用日志"
        exit 1
        ;;
esac
