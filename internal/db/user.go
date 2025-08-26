package db

import (
	"database/sql"
	"fmt"
	"tts-service/internal/models"
)

// CreateUser 创建用户
func (db *DB) CreateUser(user *models.User) error {
	query := `INSERT INTO users (api_key, name) VALUES (?, ?)`
	result, err := db.Exec(query, user.APIKey, user.Name)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取用户ID失败: %w", err)
	}

	user.ID = int(id)
	return nil
}

// GetUserByAPIKey 通过API Key获取用户
func (db *DB) GetUserByAPIKey(apiKey string) (*models.User, error) {
	query := `SELECT id, api_key, name, created_at FROM users WHERE api_key = ?`
	
	var user models.User
	err := db.QueryRow(query, apiKey).Scan(
		&user.ID,
		&user.APIKey,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetUserByID 通过ID获取用户
func (db *DB) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, api_key, name, created_at FROM users WHERE id = ?`
	
	var user models.User
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.APIKey,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}