package models

import (
	"time"
)

// User 用户模型（简化版，仅用于API Key管理）
type User struct {
	ID        int       `json:"id" db:"id"`
	APIKey    string    `json:"api_key" db:"api_key"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TTSCache TTS缓存模型
type TTSCache struct {
	ID        int       `json:"id" db:"id"`
	TextHash  string    `json:"text_hash" db:"text_hash"`
	Voice     string    `json:"voice" db:"voice"`
	Format    string    `json:"format" db:"format"`
	AudioPath string    `json:"audio_path" db:"audio_path"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TTSRequest TTS请求模型
type TTSRequest struct {
	Text   string  `json:"text" binding:"required"`
	Voice  string  `json:"voice"`
	Format string  `json:"format"`
	Speed  float64 `json:"speed"`
	Pitch  int     `json:"pitch"`
	Volume float64 `json:"volume"`
	Style  string  `json:"style"`
	SSML   bool    `json:"ssml"`
}

// TTSResponse TTS响应模型
type TTSResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *TTSData    `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// TTSData TTS数据模型
type TTSData struct {
	AudioURL string  `json:"audio_url"`
	Duration float64 `json:"duration,omitempty"`
	Size     int64   `json:"size,omitempty"`
	TaskID   string  `json:"task_id"`
}

// OpenAITTSRequest OpenAI兼容的TTS请求模型
type OpenAITTSRequest struct {
	Model          string  `json:"model" binding:"required"`
	Input          string  `json:"input" binding:"required"`
	Voice          string  `json:"voice" binding:"required"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

// ErrorResponse 错误响应模型
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}