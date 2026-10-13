package main

import (
	"log"
	"os"

	"github.com/wordtile/backend/internal/config"
	"github.com/wordtile/backend/internal/database"
	"github.com/wordtile/backend/internal/handler"
	"github.com/wordtile/backend/internal/repository"
	"github.com/wordtile/backend/internal/router"
	"github.com/wordtile/backend/internal/service"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		_ = os.Setenv("GIN_MODE", "release")
	}

	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL, 15)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	if cfg.AutoMigrate {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("自动迁移失败: %v", err)
		}
	}
	if cfg.SeedDemo {
		if err := database.SeedDemo(db); err != nil {
			log.Printf("注入演示数据失败（可忽略）: %v", err)
		}
	}

	repo := repository.New(db)

	authSvc := service.NewAuthService(repo, cfg)
	deckSvc := service.NewDeckService(repo)
	communitySvc := service.NewCommunityService(repo)
	mergeSvc := service.NewMergeService(repo)
	gameSvc := service.NewGameService(repo)
	statsSvc := service.NewStatsService(repo)
	marketCache := service.NewMarketCache(cfg.RedisAddr)

	h := &router.Handlers{
		Auth:      handler.NewAuthHandler(authSvc),
		Deck:      handler.NewDeckHandler(deckSvc),
		Community: handler.NewCommunityHandler(communitySvc, marketCache),
		Version:   handler.NewVersionHandler(mergeSvc, deckSvc),
		Game:      handler.NewGameHandler(gameSvc),
		Stats:     handler.NewStatsHandler(statsSvc),
	}

	r := router.New(h, cfg)
	log.Printf("WordTile server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
