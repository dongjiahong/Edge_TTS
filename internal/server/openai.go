package server

import (
	"net/http"
	"tts-service/internal/models"
	"tts-service/internal/tts"

	"github.com/gin-gonic/gin"
)

// OpenAIHandler OpenAI兼容接口处理器
type OpenAIHandler struct {
	ttsService *tts.TTSService
}

// NewOpenAIHandler 创建新的OpenAI处理器
func NewOpenAIHandler(ttsService *tts.TTSService) *OpenAIHandler {
	return &OpenAIHandler{
		ttsService: ttsService,
	}
}

// CreateSpeech OpenAI兼容的语音合成接口
func (h *OpenAIHandler) CreateSpeech(c *gin.Context) {
	var req models.OpenAITTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "请求参数错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// 验证必要参数
	if req.Input == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "input字段不能为空",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if req.Voice == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "voice字段不能为空",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// 转换OpenAI请求为内部TTS请求
	ttsReq := h.convertOpenAIRequest(&req)

	// 处理TTS请求
	result, err := h.ttsService.ProcessTTSRequest(ttsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "语音合成失败: " + err.Error(),
				"type":    "server_error",
			},
		})
		return
	}

	// 获取音频文件路径
	audioPath := h.ttsService.GetAudioFilePath(result.AudioURL[len("/api/v1/audio/"):])
	
	// 设置响应头并直接返回音频文件
	c.Header("Content-Type", h.getContentType(ttsReq.Format))
	c.Header("Transfer-Encoding", "chunked")
	
	// 直接提供文件下载
	c.File(audioPath)
}

// convertOpenAIRequest 将OpenAI请求转换为内部TTS请求
func (h *OpenAIHandler) convertOpenAIRequest(req *models.OpenAITTSRequest) *models.TTSRequest {
	// OpenAI语音映射到Edge TTS语音
	voice := h.mapOpenAIVoice(req.Voice)
	
	// 默认音频格式
	format := "mp3"
	if req.ResponseFormat != "" {
		format = req.ResponseFormat
	}
	
	// 默认语速
	speed := 1.0
	if req.Speed > 0 {
		speed = req.Speed
	}

	return &models.TTSRequest{
		Text:   req.Input,
		Voice:  voice,
		Format: format,
		Speed:  speed,
		Pitch:  0,
		Volume: 1.0,
		Style:  "default",
		SSML:   false,
	}
}

// mapOpenAIVoice 将OpenAI语音名称映射到Edge TTS语音
func (h *OpenAIHandler) mapOpenAIVoice(openaiVoice string) string {
	voiceMap := map[string]string{
		// OpenAI语音 -> Edge TTS语音
		"alloy":   "en-US-JennyNeural",
		"echo":    "en-US-GuyNeural", 
		"fable":   "en-US-DavisNeural",
		"onyx":    "en-US-JasonNeural",
		"nova":    "en-US-SaraNeural",
		"shimmer": "en-US-AriaNeural",
	}

	if edgeVoice, exists := voiceMap[openaiVoice]; exists {
		return edgeVoice
	}

	// 如果没有映射，直接使用原始语音名（可能是Edge TTS的语音名）
	return openaiVoice
}

// getContentType 根据格式获取Content-Type
func (h *OpenAIHandler) getContentType(format string) string {
	switch format {
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	case "opus":
		return "audio/ogg"
	case "aac":
		return "audio/aac"
	case "flac":
		return "audio/flac"
	case "pcm":
		return "audio/pcm"
	default:
		return "audio/mpeg"
	}
}

// GetModels 获取可用模型列表（OpenAI兼容）
func (h *OpenAIHandler) GetModels(c *gin.Context) {
	models := gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"id":       "tts-1",
				"object":   "model",
				"created":  1677610602,
				"owned_by": "openai-internal",
				"permission": []gin.H{},
				"root":     "tts-1",
				"parent":   nil,
			},
			{
				"id":       "tts-1-hd",
				"object":   "model", 
				"created":  1677610602,
				"owned_by": "openai-internal",
				"permission": []gin.H{},
				"root":     "tts-1-hd",
				"parent":   nil,
			},
		},
	}

	c.JSON(http.StatusOK, models)
}

// GetVoicesOpenAI 获取OpenAI兼容的语音列表
func (h *OpenAIHandler) GetVoicesOpenAI(c *gin.Context) {
	voices := []gin.H{
		{
			"id":          "alloy",
			"name":        "Alloy",
			"description": "A balanced, neutral voice",
			"language":    "en-US",
			"gender":      "neutral",
		},
		{
			"id":          "echo", 
			"name":        "Echo",
			"description": "A clear, expressive voice",
			"language":    "en-US",
			"gender":      "male",
		},
		{
			"id":          "fable",
			"name":        "Fable",
			"description": "A warm, storytelling voice", 
			"language":    "en-US",
			"gender":      "neutral",
		},
		{
			"id":          "onyx",
			"name":        "Onyx",
			"description": "A deep, authoritative voice",
			"language":    "en-US", 
			"gender":      "male",
		},
		{
			"id":          "nova",
			"name":        "Nova",
			"description": "A bright, energetic voice",
			"language":    "en-US",
			"gender":      "female",
		},
		{
			"id":          "shimmer",
			"name":        "Shimmer", 
			"description": "A soft, elegant voice",
			"language":    "en-US",
			"gender":      "female",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"voices": voices,
	})
}