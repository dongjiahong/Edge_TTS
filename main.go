package main

import (
	"flag"
	"log"
	"os"
	"tts-service/internal/config"
	"tts-service/internal/db"
	"tts-service/internal/server"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建必要的目录
	if err := os.MkdirAll(cfg.Storage.Path, 0755); err != nil {
		log.Fatalf("创建存储目录失败: %v", err)
	}
	
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 初始化数据库
	database, err := db.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 启动服务器
	srv := server.New(cfg, database)
	if err := srv.Start(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}