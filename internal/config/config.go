package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Storage  StorageConfig  `yaml:"storage"`
	TTS      TTSConfig      `yaml:"tts"`
	EdgeTTS  EdgeTTSConfig  `yaml:"edge_tts"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type StorageConfig struct {
	Path         string `yaml:"path"`
	MaxSize      string `yaml:"max_size"`
	CleanupHours int    `yaml:"cleanup_hours"`
}

type TTSConfig struct {
	Engines       []string `yaml:"engines"`
	DefaultVoice  string   `yaml:"default_voice"`
	DefaultFormat string   `yaml:"default_format"`
}

type EdgeTTSConfig struct {
	Endpoint  string `yaml:"endpoint"`
	UserAgent string `yaml:"user_agent"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}