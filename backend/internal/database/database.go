package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/wordtile/backend/internal/model"
)

// Connect 带重试地连接 PostgreSQL（docker-compose 启动顺序兜底）。
func Connect(dsn string, retries int) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				if pingErr = sqlDB.Ping(); pingErr == nil {
					return db, nil
				}
			}
			err = pingErr
		}
		lastErr = err
		log.Printf("waiting for database (%d/%d): %v", i+1, retries, err)
		time.Sleep(2 * time.Second)
	}
	return nil, lastErr
}

// AutoMigrate 开发模式建表；生产环境请使用 migrations/ 下的 golang-migrate 脚本。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Deck{},
		&model.Entry{},
		&model.Commit{},
		&model.Conflict{},
		&model.Comment{},
		&model.Like{},
		&model.GameSession{},
	)
}
