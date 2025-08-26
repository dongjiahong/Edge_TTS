# TTS Service - 通用文本转语音服务

基于微软 Edge TTS 的通用文本转语音服务，提供 REST API 和 OpenAI 兼容接口。

## ✨ 特性

- 🎯 **基于 Edge TTS**: 高质量的语音合成
- 🚀 **高性能**: Go 语言开发，支持并发处理
- 💾 **双重缓存**: SQLite + Redis 智能缓存系统
- 🔗 **OpenAI 兼容**: 支持 OpenAI TTS API 格式
- 🌐 **多语言支持**: 支持中英文等多种语言
- 📦 **本地部署**: 简单的本地化部署方案
- 🔒 **API 认证**: 基于 API Key 的认证机制

## 🏗️ 技术架构

```
Client → HTTP API → Go Server → Redis缓存 → Edge TTS WebSocket
                         ↓
                    SQLite数据库
```

- **后端**: Go + Gin Framework
- **数据库**: SQLite (轻量级，零配置)
- **缓存**: Redis (可选) + SQLite
- **语音引擎**: Microsoft Edge TTS
- **部署**: 本地化部署，单一可执行文件

## 🚀 快速开始

### 1. 初始化项目

```bash
# 克隆或下载项目后，运行初始化脚本
./scripts/init.sh
```

初始化脚本会：
- 创建必要的目录结构
- 生成默认配置文件
- 安装 Go 依赖
- 编译项目
- 生成初始 API Key

### 2. 启动服务

```bash
# 启动服务
./scripts/start.sh

# 或者直接运行
./tts-service -config config.yaml
```

### 3. 测试服务

```bash
# 健康检查
curl http://localhost:2828/api/v1/health

# 获取语音列表
curl http://localhost:2828/api/v1/voices
```

## 📖 API 文档

### 基础 TTS 接口

```bash
curl -X POST http://localhost:2828/api/v1/tts/synthesize \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "你好，这是一个测试",
    "voice": "zh-CN-XiaoxiaoNeural",
    "format": "mp3",
    "speed": 1.0
  }'
```

**响应示例:**
```json
{
  "code": 200,
  "message": "合成成功",
  "data": {
    "audio_url": "/api/v1/audio/xxxxx.mp3",
    "size": 204800,
    "task_id": "abc123"
  }
}
```

### OpenAI 兼容接口

```bash
curl -X POST http://localhost:2828/api/v1/audio/speech \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1",
    "input": "Hello world",
    "voice": "alloy",
    "response_format": "mp3"
  }' --output speech.mp3
```

### 可用语音

- **中文**: `zh-CN-XiaoxiaoNeural`, `zh-CN-YunxiNeural`
- **英文**: `en-US-JennyNeural`, `en-US-GuyNeural`
- **OpenAI映射**: `alloy`, `echo`, `fable`, `onyx`, `nova`, `shimmer`

### 支持格式

- `mp3` - MP3 音频格式 (默认)
- `wav` - WAV 音频格式
- `ogg` - OGG 音频格式

## 🛠️ 配置说明

`config.yaml` 配置文件：

```yaml
server:
  port: 2828          # 服务端口
  host: "0.0.0.0"     # 监听地址

database:
  path: "./tts.db"    # SQLite 数据库文件

redis:
  addr: "localhost:6379"  # Redis 地址 (可选)
  password: ""            # Redis 密码
  db: 0                   # Redis 数据库

storage:
  path: "./storage"       # 音频文件存储目录
  cleanup_hours: 24       # 缓存清理时间 (小时)

tts:
  default_voice: "zh-CN-XiaoxiaoNeural"  # 默认语音
  default_format: "mp3"                  # 默认格式
```

## 👤 用户管理

### 创建新用户

```bash
./scripts/manage-user.sh create "用户名"
```

### 查看用户列表

```bash
./scripts/manage-user.sh list
```

### 删除用户

```bash
./scripts/manage-user.sh delete "api_key"
```

## 🔧 运维管理

### 启动/停止服务

```bash
# 启动
./scripts/start.sh

# 停止
./scripts/stop.sh
```

### 日志查看

```bash
# 查看实时日志
tail -f logs/tts.log

# 查看错误日志
grep "ERROR" logs/tts.log
```

### 缓存清理

服务会自动清理过期缓存 (24小时)，也可以手动清理：

```bash
# 删除音频文件
rm -rf storage/*

# 重启服务会自动清理数据库缓存
```

## 📊 性能优化

### SQLite 优化
- WAL 模式: 提高并发读写性能
- 内存缓存: 1GB 内存缓存加速查询
- 索引优化: 针对查询模式优化索引

### Redis 缓存 (可选)
- 热数据缓存: 1小时 TTL
- 减少数据库查询
- 提高响应速度

### 并发处理
- Goroutine 池: 高效处理并发请求
- WebSocket 连接复用: 减少连接开销
- 智能缓存: 避免重复合成

## 🐛 常见问题

### Q: Redis 连接失败怎么办？
A: 服务会自动回退到 SQLite 缓存，不影响正常使用。可以注释掉 `config.yaml` 中的 Redis 配置。

### Q: Edge TTS 连接失败？
A: 检查网络连接，确保可以访问 `speech.platform.bing.com`。

### Q: 音频文件过大？
A: 可以调整 `storage.cleanup_hours` 配置，定期清理缓存文件。

### Q: 如何增加新的语音？
A: 修改 `internal/server/handlers.go` 中的 `GetVoices` 方法，添加新的语音选项。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

基于 [yy4382/read-aloud](https://github.com/yy4382/read-aloud) 项目的架构设计思路，感谢原作者的开源贡献。
