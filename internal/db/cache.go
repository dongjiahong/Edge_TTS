package db

import (
	"database/sql"
	"fmt"
	"tts-service/internal/models"
)

// CreateTTSCache 创建TTS缓存记录
func (db *DB) CreateTTSCache(cache *models.TTSCache) error {
	query := `INSERT INTO tts_cache (text_hash, voice, format, audio_path) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, cache.TextHash, cache.Voice, cache.Format, cache.AudioPath)
	if err != nil {
		return fmt.Errorf("创建TTS缓存失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取缓存ID失败: %w", err)
	}

	cache.ID = int(id)
	return nil
}

// GetTTSCache 获取TTS缓存
func (db *DB) GetTTSCache(textHash, voice, format string) (*models.TTSCache, error) {
	query := `SELECT id, text_hash, voice, format, audio_path, created_at 
			  FROM tts_cache 
			  WHERE text_hash = ? AND voice = ? AND format = ?`
	
	var cache models.TTSCache
	err := db.QueryRow(query, textHash, voice, format).Scan(
		&cache.ID,
		&cache.TextHash,
		&cache.Voice,
		&cache.Format,
		&cache.AudioPath,
		&cache.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到缓存，返回nil而不是错误
		}
		return nil, fmt.Errorf("查询TTS缓存失败: %w", err)
	}

	return &cache, nil
}

// DeleteExpiredCache 删除过期的缓存记录
func (db *DB) DeleteExpiredCache(hours int) (int64, error) {
	query := `DELETE FROM tts_cache WHERE created_at < datetime('now', '-' || ? || ' hours')`
	result, err := db.Exec(query, hours)
	if err != nil {
		return 0, fmt.Errorf("删除过期缓存失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取影响行数失败: %w", err)
	}

	return affected, nil
}

// GetCacheStats 获取缓存统计信息
func (db *DB) GetCacheStats() (map[string]interface{}, error) {
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM tts_cache`
	if err := db.QueryRow(countQuery).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("获取缓存总数失败: %w", err)
	}

	var oldestDate sql.NullString
	oldestQuery := `SELECT MIN(created_at) FROM tts_cache`
	if err := db.QueryRow(oldestQuery).Scan(&oldestDate); err != nil {
		return nil, fmt.Errorf("获取最早缓存时间失败: %w", err)
	}

	stats := map[string]interface{}{
		"total_count": totalCount,
		"oldest_date": nil,
	}

	if oldestDate.Valid {
		stats["oldest_date"] = oldestDate.String
	}

	return stats, nil
}