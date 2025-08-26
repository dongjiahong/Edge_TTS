package db

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

// DB 数据库连接
type DB struct {
	*sql.DB
}

// Init 初始化数据库
func Init(dbPath string) (*DB, error) {
	// 连接数据库
	sqlDB, err := sql.Open("sqlite", dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL&_sync=NORMAL&_cache_size=1000000")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	db := &DB{sqlDB}

	// 设置 SQLite 优化参数
	if err := db.optimize(); err != nil {
		return nil, fmt.Errorf("优化数据库失败: %w", err)
	}

	// 创建表
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return db, nil
}

// optimize 优化 SQLite 配置
func (db *DB) optimize() error {
	optimizations := []string{
		"PRAGMA journal_mode = WAL;",        // 写前日志，提高并发
		"PRAGMA synchronous = NORMAL;",      // 平衡性能和安全
		"PRAGMA cache_size = 1000000;",      // 1GB缓存
		"PRAGMA temp_store = memory;",       // 临时数据存内存
		"PRAGMA mmap_size = 268435456;",     // 256MB内存映射
		"PRAGMA foreign_keys = ON;",         // 启用外键约束
	}

	for _, sql := range optimizations {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("执行优化SQL失败 [%s]: %w", sql, err)
		}
	}

	return nil
}

// createTables 创建数据库表
func (db *DB) createTables() error {
	// 用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_key TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// TTS缓存表
	cacheTable := `
	CREATE TABLE IF NOT EXISTS tts_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text_hash TEXT UNIQUE NOT NULL,
		voice TEXT NOT NULL,
		format TEXT NOT NULL,
		audio_path TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_text_hash ON tts_cache(text_hash);",
		"CREATE INDEX IF NOT EXISTS idx_created_at ON tts_cache(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_api_key ON users(api_key);",
	}

	// 执行创建表语句
	if _, err := db.Exec(userTable); err != nil {
		return fmt.Errorf("创建用户表失败: %w", err)
	}

	if _, err := db.Exec(cacheTable); err != nil {
		return fmt.Errorf("创建缓存表失败: %w", err)
	}

	// 创建索引
	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("创建索引失败 [%s]: %w", idx, err)
		}
	}

	return nil
}