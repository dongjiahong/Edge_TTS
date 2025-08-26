package tts

import (
	"fmt"
	"os"
	"path/filepath"
	"tts-service/internal/cache"
	"tts-service/internal/config"
	"tts-service/internal/db"
	"tts-service/internal/models"
	"tts-service/internal/utils"
)

// TTSService TTS服务
type TTSService struct {
	db         *db.DB
	config     *config.Config
	edgeClient *EdgeTTSClient
	redis      *cache.RedisClient
}

// NewTTSService 创建新的TTS服务
func NewTTSService(database *db.DB, cfg *config.Config) *TTSService {
	edgeClient := NewEdgeTTSClient(&cfg.EdgeTTS)
	
	// 初始化Redis客户端（可选）
	var redisClient *cache.RedisClient
	if cfg.Redis.Addr != "" {
		var err error
		redisClient, err = cache.NewRedisClient(&cfg.Redis)
		if err != nil {
			fmt.Printf("Redis初始化失败，将使用SQLite缓存: %v\n", err)
		}
	}
	
	return &TTSService{
		db:         database,
		config:     cfg,
		edgeClient: edgeClient,
		redis:      redisClient,
	}
}

// ProcessTTSRequest 处理TTS请求
func (s *TTSService) ProcessTTSRequest(req *models.TTSRequest) (*models.TTSData, error) {
	// 设置默认值
	if req.Voice == "" {
		req.Voice = s.config.TTS.DefaultVoice
	}
	if req.Format == "" {
		req.Format = s.config.TTS.DefaultFormat
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Volume == 0 {
		req.Volume = 1.0
	}

	// 生成文本哈希用于缓存
	textHash := utils.GenerateTextHash(req.Text, req.Voice, req.Format)
	cacheKey := fmt.Sprintf("tts:%s", textHash)
	
	// 首先检查Redis缓存
	if s.redis != nil {
		if audioPath, err := s.redis.Get(cacheKey); err == nil && audioPath != "" {
			// 检查文件是否存在
			if _, err := os.Stat(audioPath); err == nil {
				// Redis缓存命中
				return &models.TTSData{
					AudioURL: s.getAudioURL(audioPath),
					TaskID:   utils.GenerateRequestID(),
				}, nil
			} else {
				// 文件不存在，删除Redis缓存
				s.redis.Delete(cacheKey)
			}
		}
	}
	
	// 检查SQLite缓存
	if cache, err := s.db.GetTTSCache(textHash, req.Voice, req.Format); err == nil && cache != nil {
		// 检查文件是否存在
		if _, err := os.Stat(cache.AudioPath); err == nil {
			// SQLite缓存命中，同时更新Redis缓存
			if s.redis != nil {
				s.redis.SetWithTTL(cacheKey, cache.AudioPath, 3600) // 1小时TTL
			}
			return &models.TTSData{
				AudioURL: s.getAudioURL(cache.AudioPath),
				TaskID:   utils.GenerateRequestID(),
			}, nil
		} else {
			// 文件不存在，删除缓存记录
			// 这里可以添加删除缓存记录的逻辑
		}
	}

	// 调用Edge TTS进行语音合成
	audioData, err := s.edgeClient.Synthesize(req.Text, req.Voice, req.Format, req.Speed, req.Pitch)
	if err != nil {
		return nil, fmt.Errorf("语音合成失败: %w", err)
	}

	// 保存音频文件
	audioPath, err := s.saveAudioFile(audioData, textHash, req.Format)
	if err != nil {
		return nil, fmt.Errorf("保存音频文件失败: %w", err)
	}

	// 保存SQLite缓存记录
	cache := &models.TTSCache{
		TextHash:  textHash,
		Voice:     req.Voice,
		Format:    req.Format,
		AudioPath: audioPath,
	}
	if err := s.db.CreateTTSCache(cache); err != nil {
		// 缓存保存失败不影响主流程，只记录日志
		fmt.Printf("保存SQLite缓存失败: %v\n", err)
	}
	
	// 保存Redis缓存
	if s.redis != nil {
		if err := s.redis.SetWithTTL(cacheKey, audioPath, 3600); err != nil {
			fmt.Printf("保存Redis缓存失败: %v\n", err)
		}
	}

	return &models.TTSData{
		AudioURL: s.getAudioURL(audioPath),
		Size:     int64(len(audioData)),
		TaskID:   utils.GenerateRequestID(),
	}, nil
}

// saveAudioFile 保存音频文件
func (s *TTSService) saveAudioFile(audioData []byte, hash, format string) (string, error) {
	// 确保存储目录存在
	if err := os.MkdirAll(s.config.Storage.Path, 0755); err != nil {
		return "", err
	}

	// 生成文件名
	filename := hash + utils.GetFileExtension(format)
	audioPath := filepath.Join(s.config.Storage.Path, filename)

	// 写入文件
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		return "", err
	}

	return audioPath, nil
}

// getAudioURL 生成音频访问URL
func (s *TTSService) getAudioURL(audioPath string) string {
	filename := filepath.Base(audioPath)
	return fmt.Sprintf("/api/v1/audio/%s", filename)
}

// GetAudioFilePath 根据文件名获取音频文件路径
func (s *TTSService) GetAudioFilePath(filename string) string {
	// 安全检查，防止路径遍历
	cleanFilename := utils.SanitizeFileName(filename)
	return filepath.Join(s.config.Storage.Path, cleanFilename)
}

// CleanupExpiredCache 清理过期缓存
func (s *TTSService) CleanupExpiredCache() error {
	// 删除数据库中的过期记录
	deleted, err := s.db.DeleteExpiredCache(s.config.Storage.CleanupHours)
	if err != nil {
		return fmt.Errorf("清理数据库缓存失败: %w", err)
	}

	fmt.Printf("清理了 %d 条过期缓存记录\n", deleted)

	// TODO: 清理对应的音频文件（可以添加文件系统清理逻辑）

	return nil
}