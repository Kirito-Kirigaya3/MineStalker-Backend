package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"teamacedia/minestalker/internal/api"
	"teamacedia/minestalker/internal/config"
	"teamacedia/minestalker/internal/db"
	"teamacedia/minestalker/internal/discord"
	"teamacedia/minestalker/internal/scraper"
)

func main() {
	// Load config file
	cfg, err := config.LoadConfig("config.ini")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DB
	err = db.InitDB("minestalker.db")
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	// Start scraping job
	go scraper.StartScheduler(cfg.UpdateInterval, cfg.SnapshotInterval, cfg.LoggerWebhookUrl, cfg.LoggerWebhookUsername)

	// Start the Discord bot

	go discord.Start(cfg.Token, cfg.AppID, cfg.GuildID)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/player/", api.PlayerHistoryHandler)
	mux.HandleFunc("/api/server/", api.ServerHistoryHandler)
	mux.HandleFunc("/api/snapshot", api.SnapshotHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run server in goroutine
	go func() {
		log.Println("MineStalker-Backend running on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until signal is received
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
