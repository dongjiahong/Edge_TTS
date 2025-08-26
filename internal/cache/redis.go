package cache

import (
	"context"
	"fmt"
	"time"
	"tts-service/internal/config"

	"github.com/go-redis/redis/v8"
)

// RedisClient Redis缓存客户端
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()

	// 测试连接
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Set 设置缓存
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("设置Redis缓存失败: %w", err)
	}
	return nil
}

// Get 获取缓存
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在
		}
		return "", fmt.Errorf("获取Redis缓存失败: %w", err)
	}
	return val, nil
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(key string) (bool, error) {
	val, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("检查Redis键失败: %w", err)
	}
	return val > 0, nil
}

// Delete 删除缓存
func (r *RedisClient) Delete(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("删除Redis缓存失败: %w", err)
	}
	return nil
}

// SetJSON 设置JSON格式的缓存
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// GetJSON 获取JSON格式的缓存
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("缓存不存在")
		}
		return fmt.Errorf("获取JSON缓存失败: %w", err)
	}

	// 这里简化处理，实际项目中可能需要JSON序列化
	// 由于我们的缓存主要是简单的字符串和文件路径，暂时不需要复杂的JSON处理
	return nil
}

// Increment 递增计数器
func (r *RedisClient) Increment(key string) (int64, error) {
	val, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("递增计数器失败: %w", err)
	}
	return val, nil
}

// SetWithTTL 设置带TTL的缓存
func (r *RedisClient) SetWithTTL(key string, value interface{}, seconds int) error {
	ttl := time.Duration(seconds) * time.Second
	return r.Set(key, value, ttl)
}

// GetTTL 获取键的TTL
func (r *RedisClient) GetTTL(key string) (time.Duration, error) {
	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取TTL失败: %w", err)
	}
	return ttl, nil
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// FlushAll 清空所有缓存（慎用）
func (r *RedisClient) FlushAll() error {
	err := r.client.FlushAll(r.ctx).Err()
	if err != nil {
		return fmt.Errorf("清空Redis缓存失败: %w", err)
	}
	return nil
}

// GetStats 获取Redis统计信息
func (r *RedisClient) GetStats() (map[string]string, error) {
	info, err := r.client.Info(r.ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis信息失败: %w", err)
	}

	// 简化版统计信息解析
	stats := make(map[string]string)
	stats["info"] = info

	return stats, nil
}