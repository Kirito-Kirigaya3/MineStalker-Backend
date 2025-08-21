package tracker

import (
	"fmt"
	"log"
	"teamacedia/minestalker/internal/db"
	"teamacedia/minestalker/internal/models"
	"time"
)

var previousState = map[string]map[int]map[string]bool{} // map[serverAddr][serverPort][playerName]bool
var lastSnapshotSave time.Time

func RefreshTracker(current models.ServerListResponse, snapshot_interval_seconds int) []models.TrackingEvent {
	now := time.Now()
	var events []models.TrackingEvent

	// Only save snapshot if specified interval has passed
	if now.Sub(lastSnapshotSave) > time.Duration(snapshot_interval_seconds)*time.Second {
		log.Println("Saving serverlist snapshot to DB...")
		snapshot := models.Snapshot{
			Servers: current.List,
			Time:    now,
		}

		err := db.SaveSnapshot(snapshot)
		if err != nil {
			fmt.Printf("Error saving snapshot: %v\n", err)
			return nil
		}

		lastSnapshotSave = now
	}

	currentState := make(map[string]map[int]map[string]bool)

	// Detect offline servers (in previousState but NOT in current)
	for addr, prevPorts := range previousState {
		if !serverInList(addr, current.List) {
			for port, prevPlayers := range prevPorts {
				server, _ := db.GetServerInfo(addr, port)
				// Send serverOffline event
				events = append(events, models.TrackingEvent{
					Type:      "serverOffline",
					Server:    addr,
					Port:      port,
					Timestamp: now,
					Name:      server.Name,
				})

				// Send playerLeave events for all players on that server+port
				for player := range prevPlayers {
					events = append(events, models.TrackingEvent{
						Type:      "playerLeave",
						Player:    player,
						Server:    addr,
						Port:      port,
						Timestamp: now,
						Name:      server.Name,
					})
				}
			}
		}
	}

	// Detect online servers (in current but NOT in previousState)
	for _, server := range current.List {
		if server.Address == "" {
			continue
		}

		if _, ok := previousState[server.Address]; !ok {
			// Entire server is new (not in previousState)
			events = append(events, models.TrackingEvent{
				Type:      "serverOnline",
				Server:    server.Address,
				Port:      server.Port,
				Timestamp: now,
				Name:      server.Name,
			})
		} else if _, ok := previousState[server.Address][server.Port]; !ok {
			// This port is new on existing server address
			events = append(events, models.TrackingEvent{
				Type:      "serverOnline",
				Server:    server.Address,
				Port:      server.Port,
				Timestamp: now,
				Name:      server.Name,
			})
		}

		// Initialize currentState nested maps
		if currentState[server.Address] == nil {
			currentState[server.Address] = make(map[int]map[string]bool)
		}
		currPlayers := make(map[string]bool)
		for _, p := range server.PlayerList {
			currPlayers[p] = true
		}
		currentState[server.Address][server.Port] = currPlayers

		prevPlayers := getPrevPlayers(previousState, server.Address, server.Port)

		// Detect player joins
		for player := range currPlayers {
			if !prevPlayers[player] {
				events = append(events, models.TrackingEvent{
					Type:      "playerJoin",
					Player:    player,
					Server:    server.Address,
					Port:      server.Port,
					Timestamp: now,
					Name:      server.Name,
				})
			}
		}

		// Detect player leaves
		for player := range prevPlayers {
			if !currPlayers[player] {
				events = append(events, models.TrackingEvent{
					Type:      "playerLeave",
					Player:    player,
					Server:    server.Address,
					Port:      server.Port,
					Timestamp: now,
					Name:      server.Name,
				})
			}
		}
	}
	previousState = map[string]map[int]map[string]bool{} // reset previousState
	for addr, ports := range currentState {
		if previousState[addr] == nil {
			previousState[addr] = make(map[int]map[string]bool)
		}
		for port, players := range ports {
			previousState[addr][port] = players
		}
	}

	return events
}

func serverInList(addr string, list []models.Server) bool {
	for _, s := range list {
		if s.Address == addr {
			return true
		}
	}
	return false
}

func getPrevPlayers(state map[string]map[int]map[string]bool, addr string, port int) map[string]bool {
	if ports, ok := state[addr]; ok {
		if players, ok := ports[port]; ok {
			return players
		}
	}
	return map[string]bool{}
}
