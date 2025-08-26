package server

import (
	"net/http"
	"os"
	"path/filepath"
	"tts-service/internal/models"
	"tts-service/internal/tts"

	"github.com/gin-gonic/gin"
)

// TTSHandler TTS处理器
type TTSHandler struct {
	ttsService *tts.TTSService
}

// NewTTSHandler 创建新的TTS处理器
func NewTTSHandler(ttsService *tts.TTSService) *TTSHandler {
	return &TTSHandler{
		ttsService: ttsService,
	}
}

// Synthesize 处理TTS合成请求
func (h *TTSHandler) Synthesize(c *gin.Context) {
	var req models.TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证请求参数
	if req.Text == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "文本内容不能为空",
			Error:   "text field is required",
		})
		return
	}

	// 处理TTS请求
	result, err := h.ttsService.ProcessTTSRequest(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    500,
			Message: "语音合成失败",
			Error:   err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, models.TTSResponse{
		Code:    200,
		Message: "合成成功",
		Data:    result,
	})
}

// ServeAudio 提供音频文件服务
func (h *TTSHandler) ServeAudio(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    400,
			Message: "文件名不能为空",
			Error:   "filename is required",
		})
		return
	}

	// 获取音频文件路径
	audioPath := h.ttsService.GetAudioFilePath(filename)
	
	// 检查文件是否存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Code:    404,
			Message: "音频文件不存在",
			Error:   "audio file not found",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", h.getContentType(filepath.Ext(filename)))
	c.Header("Cache-Control", "public, max-age=3600")
	
	// 提供文件服务
	c.File(audioPath)
}

// HealthCheck 健康检查
func (h *TTSHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "TTS Service",
		"version": "1.0.0",
	})
}

// GetVoices 获取可用语音列表（示例）
func (h *TTSHandler) GetVoices(c *gin.Context) {
	voices := []map[string]interface{}{
		{
			"name":        "zh-CN-XiaoxiaoNeural",
			"language":    "zh-CN",
			"gender":      "female",
			"description": "中文女声",
		},
		{
			"name":        "zh-CN-YunxiNeural", 
			"language":    "zh-CN",
			"gender":      "male",
			"description": "中文男声",
		},
		{
			"name":        "en-US-JennyNeural",
			"language":    "en-US",
			"gender":      "female", 
			"description": "英语女声",
		},
		{
			"name":        "en-US-GuyNeural",
			"language":    "en-US",
			"gender":      "male",
			"description": "英语男声",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    voices,
	})
}

// getContentType 根据文件扩展名获取Content-Type
func (h *TTSHandler) getContentType(ext string) string {
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".m4a":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	default:
		return "audio/mpeg"
	}
}