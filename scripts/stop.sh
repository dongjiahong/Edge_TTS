#!/bin/bash

# TTS服务停止脚本
echo "🛑 停止 TTS 服务..."

# 查找TTS服务进程
PID=$(pgrep -f "tts-service")

if [ -z "$PID" ]; then
    echo "ℹ️  TTS服务未运行"
else
    echo "🔍 找到TTS服务进程: $PID"
    
    # 优雅停止
    echo "📤 发送SIGTERM信号..."
    kill -TERM $PID
    
    # 等待5秒
    sleep 5
    
    # 检查进程是否仍在运行
    if kill -0 $PID 2>/dev/null; then
        echo "⚡ 强制停止服务..."
        kill -KILL $PID
        sleep 2
    fi
    
    if ! kill -0 $PID 2>/dev/null; then
        echo "✅ TTS服务已停止"
    else
        echo "❌ 无法停止TTS服务"
        exit 1
    fi
fi