package server

import (
	"fmt"
	"net/http"
	"strings"
	"tts-service/internal/db"
	"tts-service/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware API认证中间件
func AuthMiddleware(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    401,
				Message: "需要提供API Key",
				Error:   "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    401,
				Message: "无效的认证格式",
				Error:   "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		apiKey := strings.TrimPrefix(authHeader, bearerPrefix)
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    401,
				Message: "API Key不能为空",
				Error:   "API key is empty",
			})
			c.Abort()
			return
		}

		// 验证API Key
		user, err := database.GetUserByAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    401,
				Message: "无效的API Key",
				Error:   "Invalid API key",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user", user)
		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// ErrorHandlingMiddleware 错误处理中间件
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}