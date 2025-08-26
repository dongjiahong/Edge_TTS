#!/bin/bash

# 用户管理脚本

# 编译用户管理工具
echo "🔨 编译用户管理工具..."
go build -o user-manager cmd/user-manager/main.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

# 显示使用方法
show_usage() {
    echo "用法:"
    echo "  $0 list                    # 列出所有用户"
    echo "  $0 create <name>          # 创建新用户"
    echo "  $0 delete <api_key>       # 删除用户"
    echo ""
    echo "示例:"
    echo "  $0 list"
    echo "  $0 create test-user"
    echo "  $0 delete abc123..."
}

# 检查参数
if [ $# -eq 0 ]; then
    show_usage
    exit 1
fi

ACTION=$1

case $ACTION in
    "list")
        echo "📋 查询用户列表..."
        ./user-manager -action list
        ;;
    "create")
        if [ -z "$2" ]; then
            echo "❌ 错误: 请提供用户名"
            echo "用法: $0 create <用户名>"
            exit 1
        fi
        echo "👤 创建新用户: $2"
        ./user-manager -action create -name "$2"
        ;;
    "delete")
        if [ -z "$2" ]; then
            echo "❌ 错误: 请提供API Key"
            echo "用法: $0 delete <api_key>"
            exit 1
        fi
        echo "🗑️  删除用户..."
        ./user-manager -action delete -key "$2"
        ;;
    *)
        echo "❌ 未知操作: $ACTION"
        show_usage
        exit 1
        ;;
esac

# 清理编译文件
rm -f user-manager