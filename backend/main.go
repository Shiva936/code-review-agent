package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/router"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

func main() {
	// Load configuration
	cfg, err := config.NewConfig().LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := storage.InitDB(cfg); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	handlers := router.Init(cfg)
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server running on port %s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(addr, handlers))
}
