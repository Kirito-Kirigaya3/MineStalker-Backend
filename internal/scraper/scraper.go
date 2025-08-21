package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"teamacedia/minestalker/internal/api"
	"teamacedia/minestalker/internal/db"
	"teamacedia/minestalker/internal/discord"
	"teamacedia/minestalker/internal/models"
	"teamacedia/minestalker/internal/tracker"

	"time"
)

var isFirstScrape = true
var snapshot_interval_seconds = 300 // 5 minutes default ( configurable via config )

func StartScheduler(update_interval_seconds int, snapshot_interval_seconds_ int) {
	ticker := time.NewTicker(time.Duration(update_interval_seconds) * time.Second)
	defer ticker.Stop()

	snapshot_interval_seconds = snapshot_interval_seconds_
	log.Println("Scraper scheduler started, scraping every 5 minutes")
	Scrape()

	// Create a channel to listen for OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			Scrape()
		case <-sigs:
			log.Println("Received interrupt signal, shutting down scheduler...")
			return
		}
	}
}

type ServerListResponse struct {
	List []models.Server `json:"list"`
}

// Helper to send a formatted webhook
func sendEvent(message string) {
	api.SendWebhook("https://discord.com/api/webhooks/1407993761409142864/RrchikWMzGy-_JEFQ0t6pvLh0q_7EMBj0kWC1-ZQCR1415dSOeagfvtiDZ8HJvw6rZ4w", message)
}

func Scrape() {
	log.Println("Starting scrape of servers.minetest.net list...")

	resp, err := http.Get("https://servers.minetest.net/list")
	if err != nil {
		log.Printf("Failed to fetch server list: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-200 response: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return
	}

	var parsed models.ServerListResponse
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		log.Println(string(body[:min(len(body), 1000)]))
		return
	}

	log.Printf("Found %d servers", len(parsed.List))
	/*for i, s := range parsed.List {
		log.Printf("[%d] Server: %s — %s — %d players", i+1, s.Address, s.Name, s.Clients)
		if i >= 4 {
			break
		}
	}*/

	log.Println("Tracking player/server events...")

	events := tracker.RefreshTracker(parsed, snapshot_interval_seconds)
	sortEventsByType(events)

	log.Println("Committing changes to database...")

	for _, event := range events {
		db.HandleEvent(event)
	}

	log.Printf("Tracked %d events", len(events))
	if isFirstScrape {
		log.Println("First scrape complete, skipping webhook notifications and discord alerts")
		isFirstScrape = false
		return
	}
	for _, event := range events {
		switch event.Type {
		case "playerJoin":
			log.Printf("Player %s joined server %s at %s", event.Player, event.Server, event.Timestamp)
			sendEvent(fmt.Sprintf("👤 🟢 ➕ Player **%s** joined server **%s**", event.Player, event.Name))

		case "playerLeave":
			log.Printf("Player %s left server %s at %s", event.Player, event.Server, event.Timestamp)
			sendEvent(fmt.Sprintf("👤 🔴 ➖ Player **%s** left server **%s**", event.Player, event.Name))

		case "serverOnline":
			log.Printf("Server %s:%d is now online", event.Server, event.Port)
			sendEvent(fmt.Sprintf("📡 Server **%s** (%s:%d) is now **ONLINE** ✅", event.Name, event.Server, event.Port))

		case "serverOffline":
			log.Printf("Server %s:%d is now offline", event.Server, event.Port)
			sendEvent(fmt.Sprintf("📡 Server **%s** (%s:%d) is now **OFFLINE** ❌", event.Name, event.Server, event.Port))
		}
		time.Sleep(100 * time.Millisecond) // Throttle webhook sends
	}

	log.Println("Sending Discord alerts for tracked events...")
	discord.HandleEvents(events)
}

func sortEventsByType(events []models.TrackingEvent) {
	order := map[string]int{
		"serverOnline":  1,
		"serverOffline": 2,
		"playerJoin":    3,
		"playerLeave":   4,
	}

	sort.Slice(events, func(i, j int) bool {
		return order[events[i].Type] < order[events[j].Type]
	})
}
