#!/bin/bash

# TTS服务启动脚本
echo "🚀 启动 TTS 服务..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "❌ 错误: 未找到配置文件 config.yaml"
    exit 1
fi

# 创建必要的目录
mkdir -p storage
mkdir -p logs
mkdir -p data

echo "📁 创建存储目录完成"

# 检查Redis连接（可选）
if command -v redis-cli &> /dev/null; then
    if redis-cli ping &> /dev/null; then
        echo "✅ Redis连接正常"
    else
        echo "⚠️  警告: Redis未运行，将使用SQLite缓存"
    fi
else
    echo "⚠️  警告: 未安装redis-cli，跳过Redis连接检查"
fi

# 初始化Go模块（如果需要）
if [ ! -f "go.sum" ]; then
    echo "📦 初始化Go依赖..."
    go mod tidy
fi

# 编译项目
echo "🔨 编译项目..."
go build -o tts-service main.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"

# 创建默认API用户（如果数据库为空）
echo "👤 检查默认用户..."

# 启动服务
echo "🌟 启动服务..."
./tts-service -config config.yaml
