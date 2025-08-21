package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"teamacedia/minestalker/internal/db"
)

// PlayerHistoryHandler serves player history by name
func PlayerHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Extract player name from the URL path
	// Expecting: /api/player/<name>
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "Missing player name in URL", http.StatusBadRequest)
		return
	}
	playerName := parts[3]

	// Query DB for player history
	history, err := db.GetPlayerHistory(playerName)
	if err != nil {
		http.Error(w, "Error retrieving player history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize and write JSON response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(history)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// ServerHistoryHandler serves server connection history by server address and port
func ServerHistoryHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[3] == "" || parts[4] == "" {
		http.Error(w, "Missing server address or port in URL", http.StatusBadRequest)
		return
	}
	serverAddress := parts[3]
	portStr := parts[4]

	serverPort, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid server port", http.StatusBadRequest)
		return
	}

	// Query DB for snapshot history of the server
	snapshotHistory, err := db.GetSnapshotHistoryForServer(serverAddress, serverPort)
	if err != nil {
		http.Error(w, "Error retrieving snapshot history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize and write JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshotHistory); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func SnapshotHandler(w http.ResponseWriter, r *http.Request) {
	// Query DB for latest server snapshot
	snapshot, err := db.GetLatestSnapshot()
	if err != nil {
		http.Error(w, "Error retrieving latest snapshot: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize and write JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func SendWebhook(webhookURL, message string) error {
	payload := map[string]string{
		"content":  message,
		"username": "MineStalker", // show up as "MineStalker"
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %s", resp.Status)
	}
	return nil
}
