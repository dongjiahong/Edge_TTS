package server

import (
	"fmt"
	"tts-service/internal/config"
	"tts-service/internal/db"
	"tts-service/internal/tts"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	config     *config.Config
	db         *db.DB
	ttsService *tts.TTSService
	router     *gin.Engine
}

// New 创建新的服务器实例
func New(cfg *config.Config, database *db.DB) *Server {
	// 设置Gin模式
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建TTS服务
	ttsService := tts.NewTTSService(database, cfg)

	server := &Server{
		config:     cfg,
		db:         database,
		ttsService: ttsService,
		router:     gin.New(),
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 添加中间件
	s.router.Use(LoggingMiddleware())
	s.router.Use(ErrorHandlingMiddleware())
	s.router.Use(CORSMiddleware())

	// 创建处理器
	ttsHandler := NewTTSHandler(s.ttsService)
	openaiHandler := NewOpenAIHandler(s.ttsService)

	// 公开路由（无需认证）
	public := s.router.Group("/api/v1")
	{
		public.GET("/health", ttsHandler.HealthCheck)
		public.GET("/voices", ttsHandler.GetVoices)
		public.GET("/audio/:filename", ttsHandler.ServeAudio)
	}

	// 需要认证的路由
	private := s.router.Group("/api/v1")
	private.Use(AuthMiddleware(s.db))
	{
		// 基础TTS接口
		private.POST("/tts/synthesize", ttsHandler.Synthesize)
		
		// OpenAI兼容接口
		private.POST("/audio/speech", openaiHandler.CreateSpeech)
		private.GET("/models", openaiHandler.GetModels)
		private.GET("/voices/openai", openaiHandler.GetVoicesOpenAI)
	}

	// 根路径
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "TTS Service API",
			"version": "1.0.0",
			"docs":    "/api/v1/health",
		})
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("🚀 TTS服务启动成功！\n")
	fmt.Printf("📡 监听地址: http://%s\n", addr)
	fmt.Printf("🔍 健康检查: http://%s/api/v1/health\n", addr)
	fmt.Printf("📚 API文档: http://%s/\n", addr)
	
	return s.router.Run(addr)
}

// GetRouter 获取路由器（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}