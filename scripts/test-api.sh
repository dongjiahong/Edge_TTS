#!/bin/bash
# API测试脚本

API_BASE="http://localhost:2828/api/v1"
API_KEY="b74a50d601132f6ebf83ae60da6aea2a87cae548762a82ea77470e5a4527aab9"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 TTS API 测试脚本"
echo ""

# 检查服务是否运行
echo "1️⃣  检查服务状态..."
response=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE/health")
if [ "$response" -eq 200 ]; then
    echo -e "${GREEN}✅ 服务运行正常${NC}"
else
    echo -e "${RED}❌ 服务未运行 (HTTP: $response)${NC}"
    echo "请先启动服务: ./scripts/start.sh"
    exit 1
fi

# 获取语音列表
echo ""
echo "2️⃣  测试获取语音列表..."
response=$(curl -s "$API_BASE/voices")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 语音列表获取成功${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
else
    echo -e "${RED}❌ 语音列表获取失败${NC}"
fi

# 检查API Key
if [ "$API_KEY" = "your_api_key_here" ]; then
    echo ""
    echo -e "${YELLOW}⚠️  请先设置API Key:${NC}"
    echo "1. 运行 ./scripts/manage-user.sh create test-user 创建用户"
    echo "2. 将获得的API Key替换脚本中的 your_api_key_here"
    exit 0
fi

# 测试基础TTS接口
echo ""
echo "3️⃣  测试基础TTS接口..."
tts_response=$(curl -s -X POST "$API_BASE/tts/synthesize" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你好，这是一个测试",
    "voice": "zh-CN-XiaoxiaoNeural",
    "format": "mp3",
    "speed": 1.0
  }')

if echo "$tts_response" | grep -q "audio_url"; then
    echo -e "${GREEN}✅ TTS接口测试成功${NC}"
    echo "$tts_response" | jq . 2>/dev/null || echo "$tts_response"
    
    # 提取音频URL并测试下载
    audio_url=$(echo "$tts_response" | jq -r '.data.audio_url' 2>/dev/null || echo "")
    if [ -n "$audio_url" ] && [ "$audio_url" != "null" ]; then
        echo ""
        echo "4️⃣  测试音频文件下载..."
        audio_response=$(curl -s -w "%{http_code}" -o /tmp/test_audio.mp3 "http://localhost:2828$audio_url")
        if [ "$audio_response" -eq 200 ] && [ -f /tmp/test_audio.mp3 ]; then
            file_size=$(wc -c < /tmp/test_audio.mp3)
            if [ "$file_size" -gt 0 ]; then
                echo -e "${GREEN}✅ 音频文件下载成功 (大小: ${file_size} bytes)${NC}"
                echo "音频文件保存到: /tmp/test_audio.mp3"
            else
                echo -e "${RED}❌ 音频文件为空${NC}"
            fi
        else
            echo -e "${RED}❌ 音频文件下载失败 (HTTP: $audio_response)${NC}"
        fi
    fi
else
    echo -e "${RED}❌ TTS接口测试失败${NC}"
    echo "$tts_response"
fi

# 测试OpenAI兼容接口
echo ""
echo "5️⃣  测试OpenAI兼容接口..."
openai_response=$(curl -s -w "%{http_code}" -X POST "$API_BASE/audio/speech" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1",
    "input": "Hello world, this is a test",
    "voice": "alloy",
    "response_format": "mp3"
  }' -o /tmp/test_openai_audio.mp3)

if [ "$openai_response" -eq 200 ] && [ -f /tmp/test_openai_audio.mp3 ]; then
    file_size=$(wc -c < /tmp/test_openai_audio.mp3)
    if [ "$file_size" -gt 0 ]; then
        echo -e "${GREEN}✅ OpenAI接口测试成功 (大小: ${file_size} bytes)${NC}"
        echo "音频文件保存到: /tmp/test_openai_audio.mp3"
    else
        echo -e "${RED}❌ OpenAI接口返回空文件${NC}"
    fi
else
    echo -e "${RED}❌ OpenAI接口测试失败 (HTTP: $openai_response)${NC}"
fi

# 获取模型列表
echo ""
echo "6️⃣  测试获取模型列表..."
models_response=$(curl -s -H "Authorization: Bearer $API_KEY" "$API_BASE/models")
if [ $? -eq 0 ] && echo "$models_response" | grep -q "tts-1"; then
    echo -e "${GREEN}✅ 模型列表获取成功${NC}"
    echo "$models_response" | jq . 2>/dev/null || echo "$models_response"
else
    echo -e "${RED}❌ 模型列表获取失败${NC}"
fi

echo ""
echo "🎉 测试完成!"
echo ""
echo "如果你听到音频播放，说明一切正常："
echo "播放测试音频: "
if command -v afplay >/dev/null 2>&1; then
    echo "  afplay /tmp/test_audio.mp3"
elif command -v mpg123 >/dev/null 2>&1; then
    echo "  mpg123 /tmp/test_audio.mp3"
elif command -v ffplay >/dev/null 2>&1; then
    echo "  ffplay /tmp/test_audio.mp3"
else
    echo "  使用你喜欢的音频播放器播放 /tmp/test_audio.mp3"
fi
