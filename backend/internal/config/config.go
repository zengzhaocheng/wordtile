package config

import (
	"os"
	"time"
)

// Config 汇总全部环境变量配置，均提供开发期默认值。
type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string // 为空则使用进程内缓存（开发模式）
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	AutoMigrate bool   // 开发模式默认开启 GORM AutoMigrate；生产可用 golang-migrate
	SeedDemo    bool   // 首次启动注入演示账号与示例题库
	CORSOrigins string // 逗号分隔；"*" 表示全部放行
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() *Config {
	return &Config{
		Port:        env("PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://wordtile:wordtile@localhost:5432/wordtile?sslmode=disable"),
		RedisAddr:   env("REDIS_ADDR", ""),
		JWTSecret:   env("JWT_SECRET", "dev-secret-change-me"),
		AccessTTL:   15 * time.Minute,
		RefreshTTL:  7 * 24 * time.Hour,
		AutoMigrate: env("AUTO_MIGRATE", "true") == "true",
		SeedDemo:    env("SEED_DEMO", "true") == "true",
		CORSOrigins: env("CORS_ORIGINS", "*"),
	}
}
