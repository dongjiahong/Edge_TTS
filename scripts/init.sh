#!/bin/bash

# TTS服务初始化脚本
echo "🎯 初始化 TTS 服务..."

# 注意：不再使用generate_api_key函数，直接通过用户管理工具创建

# 创建必要的目录
echo "📁 创建目录结构..."
mkdir -p storage
mkdir -p logs
mkdir -p data


# 创建示例配置文件（如果不存在）
if [ ! -f "config.yaml" ]; then
    echo "📝 创建配置文件..."
    cat > config.yaml << 'EOF'
server:
  port: 2828
  host: "0.0.0.0"

database:
  path: "./tts.db"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

storage:
  path: "./storage"
  max_size: "1GB"
  cleanup_hours: 24

tts:
  engines:
    - edge
  default_voice: "zh-CN-XiaoxiaoNeural"
  default_format: "mp3"
  
edge_tts:
  endpoint: "wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1"
  user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.66 Safari/537.36 Edg/103.0.1264.44"

logging:
  level: "info"
  file: "./logs/tts.log"
EOF
    echo "✅ 配置文件创建完成"
fi

# 安装Go依赖
if [ -f "go.mod" ]; then
    echo "📦 安装Go依赖..."
    go mod tidy
    if [ $? -eq 0 ]; then
        echo "✅ 依赖安装完成"
    else
        echo "❌ 依赖安装失败"
        exit 1
    fi
fi

# 编译项目
echo "🔨 编译项目..."
go build -o tts-service main.go
if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 编译用户管理工具
echo "🔨 编译用户管理工具..."
go build -o user-manager cmd/user-manager/main.go
if [ $? -ne 0 ]; then
    echo "❌ 用户管理工具编译失败"
    exit 1
fi

# 创建初始用户
echo "👤 创建初始用户..."
INITIAL_USER_OUTPUT=$(./user-manager -action create -name "admin" 2>&1)
if [ $? -eq 0 ]; then
    API_KEY=$(echo "$INITIAL_USER_OUTPUT" | grep "API Key:" | awk '{print $3}')
    echo "✅ 初始用户创建成功"
    echo ""
    echo "🔑 初始API Key (请保存): $API_KEY"
    echo ""
else
    echo "❌ 初始用户创建失败"
    echo "$INITIAL_USER_OUTPUT"
    exit 1
fi

# 创建使用说明
cat > README_USAGE.md << EOF
# TTS服务使用说明

## 启动服务

\`\`\`bash
./scripts/start.sh
# 或者
./tts-service -config config.yaml
\`\`\`

## 停止服务

\`\`\`bash
./scripts/stop.sh
\`\`\`

## API接口

### 基础TTS接口

\`\`\`bash
curl -X POST http://localhost:2828/api/v1/tts/synthesize \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "text": "你好，这是一个测试",
    "voice": "zh-CN-XiaoxiaoNeural",
    "format": "mp3",
    "speed": 1.0
  }'
\`\`\`

### OpenAI兼容接口

\`\`\`bash
curl -X POST http://localhost:2828/api/v1/audio/speech \\
  -H "Authorization: Bearer $API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "tts-1",
    "input": "Hello world",
    "voice": "alloy",
    "response_format": "mp3"
  }' --output speech.mp3
\`\`\`

## 健康检查

\`\`\`bash
curl http://localhost:2828/api/v1/health
\`\`\`

## 获取语音列表

\`\`\`bash
curl http://localhost:2828/api/v1/voices
\`\`\`

## 初始API Key

$API_KEY

请将此API Key保存在安全的地方，并在HTTP请求的Authorization头中使用：
Authorization: Bearer $API_KEY

## 注意事项

1. 请确保Redis服务正在运行（可选，如果没有Redis将使用SQLite缓存）
2. 首次运行会自动创建SQLite数据库
3. 音频文件存储在 ./storage 目录下
4. 日志文件存储在 ./logs 目录下
5. 可以通过修改 config.yaml 来调整配置

EOF

echo "📚 使用说明已生成: README_USAGE.md"
echo ""
echo "🎉 初始化完成！"
echo ""
echo "下一步:"
echo "1. 启动服务: ./scripts/start.sh"
echo "2. 测试API: curl http://localhost:2828/api/v1/health"
echo "3. 查看完整使用说明: cat README_USAGE.md"
