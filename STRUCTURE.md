# 项目结构说明

```
tts-service/
├── README.md                 # 项目说明文档
├── STRUCTURE.md             # 项目结构说明 (本文件)
├── go.mod                   # Go模块文件
├── go.sum                   # Go依赖锁定文件
├── config.yaml              # 主配置文件
├── main.go                  # 程序入口文件
│
├── internal/                # 内部包目录
│   ├── config/             # 配置管理
│   │   └── config.go       # 配置结构和加载逻辑
│   │
│   ├── db/                 # 数据库相关
│   │   ├── db.go          # 数据库初始化和连接
│   │   ├── user.go        # 用户数据操作
│   │   └── cache.go       # 缓存数据操作
│   │
│   ├── models/            # 数据模型
│   │   └── models.go      # 所有数据结构定义
│   │
│   ├── tts/               # TTS核心服务
│   │   ├── tts.go         # TTS服务主逻辑
│   │   └── edge_tts.go    # Edge TTS客户端实现
│   │
│   ├── cache/             # 缓存服务
│   │   └── redis.go       # Redis客户端封装
│   │
│   ├── server/            # HTTP服务器
│   │   ├── server.go      # 服务器主程序
│   │   ├── handlers.go    # 基础API处理器
│   │   ├── middleware.go  # 中间件
│   │   └── openai.go      # OpenAI兼容接口
│   │
│   └── utils/             # 工具函数
│       └── utils.go       # 通用工具函数
│
├── cmd/                   # 命令行工具
│   └── user-manager/      # 用户管理工具
│       └── main.go        # 用户管理程序入口
│
├── scripts/               # 脚本目录
│   ├── init.sh           # 初始化脚本
│   ├── start.sh          # 启动脚本  
│   ├── stop.sh           # 停止脚本
│   ├── manage-user.sh    # 用户管理脚本
│   └── test-api.sh       # API测试脚本
│
├── storage/              # 音频文件存储目录 (运行时创建)
│   └── *.mp3            # 生成的音频文件
│
├── logs/                 # 日志目录 (运行时创建)
│   └── tts.log          # 应用日志
│
├── tts.db               # SQLite数据库 (运行时创建)
├── tts-service          # 编译后的主程序 (运行时生成)
└── README_USAGE.md      # 使用说明 (init.sh生成)
```

## 核心模块说明

### 1. 主程序 (main.go)
- 程序入口点
- 负责配置加载、数据库初始化、服务器启动
- 命令行参数解析

### 2. 配置管理 (internal/config/)
- 配置文件解析 (YAML格式)
- 配置结构定义
- 默认值设置

### 3. 数据库层 (internal/db/)
- SQLite 数据库初始化和优化
- 用户数据 CRUD 操作
- TTS 缓存数据管理
- 数据库连接池管理

### 4. 数据模型 (internal/models/)
- 用户模型 (User)
- TTS缓存模型 (TTSCache)
- 请求/响应模型 (TTSRequest, TTSResponse)
- OpenAI兼容模型 (OpenAITTSRequest)

### 5. TTS 核心服务 (internal/tts/)
- **tts.go**: TTS服务主逻辑
  - 请求处理和参数验证
  - 缓存查询和管理
  - 音频文件存储
  - 服务协调

- **edge_tts.go**: Edge TTS客户端
  - WebSocket连接管理
  - SSML生成和发送
  - 音频数据接收和处理
  - 协议实现

### 6. 缓存服务 (internal/cache/)
- Redis客户端封装
- 缓存键管理
- TTL设置和管理
- 错误处理和回退

### 7. HTTP 服务器 (internal/server/)
- **server.go**: 服务器主程序
  - 路由设置
  - 中间件注册
  - 服务器启动和配置

- **handlers.go**: 基础API处理器
  - TTS合成接口
  - 音频文件服务
  - 健康检查
  - 语音列表

- **middleware.go**: 中间件
  - API认证
  - CORS处理
  - 日志记录
  - 错误处理

- **openai.go**: OpenAI兼容接口
  - OpenAI格式请求转换
  - 语音映射
  - 模型列表
  - 兼容性处理

### 8. 工具函数 (internal/utils/)
- 文本哈希生成
- 文件名处理
- SSML生成
- 通用辅助函数

### 9. 命令行工具 (cmd/)
- **user-manager**: 用户管理工具
  - 用户创建、删除、列表
  - API Key生成
  - 数据库直接操作

### 10. 脚本工具 (scripts/)
- **init.sh**: 项目初始化
- **start.sh**: 服务启动
- **stop.sh**: 服务停止  
- **manage-user.sh**: 用户管理
- **test-api.sh**: API测试

## 数据流

### TTS请求处理流程
1. HTTP请求 → 中间件认证 → 处理器
2. 参数验证 → 缓存查询 (Redis → SQLite)
3. 缓存未命中 → Edge TTS WebSocket调用
4. 音频数据接收 → 文件存储 → 缓存更新
5. 返回音频URL → 客户端获取音频文件

### 缓存策略
1. **L1缓存 (Redis)**: 热数据，1小时TTL
2. **L2缓存 (SQLite)**: 持久化缓存，24小时清理
3. **文件缓存**: 实际音频文件，配合数据库记录

### WebSocket通信
1. 建立连接 → 发送音频配置
2. 发送SSML文本 → 接收音频数据流
3. 解析二进制数据 → 提取音频内容
4. 连接复用和错误处理

## 扩展点

### 1. 新增TTS引擎
在 `internal/tts/` 目录下添加新的引擎实现，遵循相同的接口规范。

### 2. 新增音频格式
在 `internal/utils/` 中添加格式支持，更新Content-Type映射。

### 3. 新增认证方式
在 `internal/server/middleware.go` 中扩展认证逻辑。

### 4. 性能优化
- 连接池优化
- 缓存策略调整
- 异步处理
- 批量处理

### 5. 监控和日志
- 添加指标收集
- 结构化日志
- 性能监控
- 错误追踪

这个架构设计实现了高内聚、低耦合的原则，便于维护和扩展。