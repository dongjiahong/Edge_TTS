package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"tts-service/internal/config"
	"tts-service/internal/db"
	"tts-service/internal/models"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		action     = flag.String("action", "list", "操作类型: list, create, delete")
		name       = flag.String("name", "", "用户名")
		apiKey     = flag.String("key", "", "API Key (create时可选，delete时必须)")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	database, err := db.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	switch *action {
	case "list":
		listUsers(database)
	case "create":
		if *name == "" {
			fmt.Println("创建用户需要提供用户名: -name <username>")
			os.Exit(1)
		}
		createUser(database, *name, *apiKey)
	case "delete":
		if *apiKey == "" {
			fmt.Println("删除用户需要提供API Key: -key <api_key>")
			os.Exit(1)
		}
		deleteUser(database, *apiKey)
	default:
		fmt.Printf("未知操作: %s\n", *action)
		fmt.Println("支持的操作: list, create, delete")
		os.Exit(1)
	}
}

func listUsers(database *db.DB) {
	query := `SELECT id, api_key, name, created_at FROM users ORDER BY created_at DESC`
	rows, err := database.Query(query)
	if err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("用户列表:")
	fmt.Printf("%-5s %-40s %-20s %-20s\n", "ID", "API Key", "Name", "Created At")
	fmt.Println(strings.Repeat("-", 90))

	count := 0
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.APIKey, &user.Name, &user.CreatedAt)
		if err != nil {
			log.Printf("扫描用户数据失败: %v", err)
			continue
		}
		
		// 脱敏显示API Key
		maskedKey := user.APIKey[:8] + "..." + user.APIKey[len(user.APIKey)-8:]
		fmt.Printf("%-5d %-40s %-20s %-20s\n", 
			user.ID, maskedKey, user.Name, user.CreatedAt.Format("2006-01-02 15:04:05"))
		count++
	}
	
	if count == 0 {
		fmt.Println("暂无用户")
	} else {
		fmt.Printf("\n总计: %d 个用户\n", count)
	}
}

func createUser(database *db.DB, name, apiKey string) {
	// 如果没有提供API Key，生成一个
	if apiKey == "" {
		apiKey = generateAPIKey()
	}

	user := &models.User{
		APIKey: apiKey,
		Name:   name,
	}

	err := database.CreateUser(user)
	if err != nil {
		log.Fatalf("创建用户失败: %v", err)
	}

	fmt.Printf("✅ 用户创建成功!\n")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("API Key: %s\n", user.APIKey)
	fmt.Println("\n请妥善保存API Key，后续无法再次查看完整密钥。")
}

func deleteUser(database *db.DB, apiKey string) {
	// 先查询用户是否存在
	user, err := database.GetUserByAPIKey(apiKey)
	if err != nil {
		fmt.Printf("❌ 用户不存在或API Key错误: %v\n", err)
		return
	}

	// 删除用户
	query := `DELETE FROM users WHERE api_key = ?`
	result, err := database.Exec(query, apiKey)
	if err != nil {
		log.Fatalf("删除用户失败: %v", err)
	}

	affected, _ := result.RowsAffected()
	if affected > 0 {
		fmt.Printf("✅ 用户删除成功: %s (ID: %d)\n", user.Name, user.ID)
	} else {
		fmt.Println("❌ 删除失败")
	}
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("生成API Key失败: %v", err)
	}
	return hex.EncodeToString(bytes)
}