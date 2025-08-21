package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"teamacedia/minestalker/internal/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(path string) error {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT NOT NULL,
		port INTEGER NOT NULL,
		name TEXT,
		game TEXT,
		first_seen DATETIME,
		last_seen DATETIME,
		UNIQUE(address, port)
	);

	CREATE TABLE IF NOT EXISTS server_sightings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL,
		seen_at DATETIME NOT NULL,
		disconnected_at DATETIME,
		FOREIGN KEY(server_id) REFERENCES servers(id)
	);

	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE
	);

	CREATE TABLE IF NOT EXISTS player_sightings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_sighting_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		seen_at DATETIME NOT NULL,
		disconnected_at DATETIME,
		FOREIGN KEY(server_sighting_id) REFERENCES server_sightings(id),
		FOREIGN KEY(player_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS snapshot_servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id INTEGER NOT NULL,
		address TEXT NOT NULL,
		port INTEGER NOT NULL,
		name TEXT,
		game TEXT,
		mods TEXT,
		clients INTEGER,
		player_list TEXT, -- store as JSON array
		FOREIGN KEY(snapshot_id) REFERENCES snapshots(id)
	);

	CREATE TABLE IF NOT EXISTS tracking_alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		player_name TEXT NOT NULL,
		discord_id TEXT NOT NULL,
		UNIQUE(player_name, discord_id)
	);

	CREATE TABLE IF NOT EXISTS server_tracking_alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_address TEXT NOT NULL,
		server_port INTEGER NOT NULL,
		discord_id TEXT NOT NULL,
		UNIQUE(server_address, server_port, discord_id)
	);
	`
	_, err = DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// CheckIfServerTrackingAlertExists checks if a server tracking alert already exists.
func CheckIfServerTrackingAlertExists(alert models.ServerTrackingAlert) bool {
	var exists bool
	query := `
	SELECT EXISTS(
		SELECT 1 FROM server_tracking_alerts
		WHERE server_address = ? AND server_port = ? AND discord_id = ?
	)
	`
	err := DB.QueryRow(query, alert.ServerAddress, alert.ServerPort, alert.DiscordID).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking server tracking alert: %v\n", err)
		return false
	}
	return exists
}

// AddServerTrackingAlert inserts a new server tracking alert.
func AddServerTrackingAlert(alert models.ServerTrackingAlert) error {
	_, err := DB.Exec(`
		INSERT INTO server_tracking_alerts (server_address, server_port, discord_id)
		VALUES (?, ?, ?)
		ON CONFLICT(server_address, server_port, discord_id) DO NOTHING
	`, alert.ServerAddress, alert.ServerPort, alert.DiscordID)
	return err
}

// RemoveServerTrackingAlert removes a server tracking alert.
func RemoveServerTrackingAlert(alert models.ServerTrackingAlert) error {
	_, err := DB.Exec(`
		DELETE FROM server_tracking_alerts
		WHERE server_address = ? AND server_port = ? AND discord_id = ?
	`, alert.ServerAddress, alert.ServerPort, alert.DiscordID)
	return err
}

// GetServerTrackingAlerts retrieves all server tracking alerts for a specific Discord ID.
func GetServerTrackingAlerts(discordId string) ([]models.ServerTrackingAlert, error) {
	query := `
	SELECT id, server_address, server_port, discord_id
	FROM server_tracking_alerts
	WHERE discord_id = ?
	`
	rows, err := DB.Query(query, discordId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []models.ServerTrackingAlert
	for rows.Next() {
		var alert models.ServerTrackingAlert
		if err := rows.Scan(&alert.ID, &alert.ServerAddress, &alert.ServerPort, &alert.DiscordID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetAllServerTrackingAlerts retrieves all server tracking alerts.
func GetAllServerTrackingAlerts() ([]models.ServerTrackingAlert, error) {
	query := `
	SELECT id, server_address, server_port, discord_id
	FROM server_tracking_alerts
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []models.ServerTrackingAlert
	for rows.Next() {
		var alert models.ServerTrackingAlert
		if err := rows.Scan(&alert.ID, &alert.ServerAddress, &alert.ServerPort, &alert.DiscordID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CheckIfTrackingAlertExists checks if a tracking alert already exists.
func CheckIfTrackingAlertExists(alert models.TrackingAlert) bool {
	var exists bool
	query := `
	SELECT EXISTS(
		SELECT 1 FROM tracking_alerts
		WHERE player_name = ? AND discord_id = ?
	)
	`
	err := DB.QueryRow(query, alert.PlayerName, alert.DiscordID).Scan(&exists)
	if err != nil {
		fmt.Printf("Error checking tracking alert: %v\n", err)
		return false
	}
	return exists
}

// AddTrackingAlert inserts a new tracking alert.
func AddTrackingAlert(alert models.TrackingAlert) error {
	_, err := DB.Exec(`
		INSERT INTO tracking_alerts (player_name, discord_id)
		VALUES (?, ?)
		ON CONFLICT(player_name, discord_id) DO NOTHING
	`, alert.PlayerName, alert.DiscordID)
	return err
}

// RemoveTrackingAlert removes a tracking alert.
func RemoveTrackingAlert(alert models.TrackingAlert) error {
	_, err := DB.Exec(`
		DELETE FROM tracking_alerts
		WHERE player_name = ? AND discord_id = ?
	`, alert.PlayerName, alert.DiscordID)
	return err
}

// GetTrackingAlerts retrieves all tracking alerts.
func GetTrackingAlerts(discordId string) ([]models.TrackingAlert, error) {
	query := `
	SELECT id, player_name, discord_id
	FROM tracking_alerts
	WHERE discord_id = ?
	`
	rows, err := DB.Query(query, discordId)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []models.TrackingAlert
	for rows.Next() {
		var alert models.TrackingAlert
		if err := rows.Scan(&alert.ID, &alert.PlayerName, &alert.DiscordID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetAllTrackingAlerts retrieves all tracking alerts.
func GetAllTrackingAlerts() ([]models.TrackingAlert, error) {
	query := `
	SELECT id, player_name, discord_id
	FROM tracking_alerts
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var alerts []models.TrackingAlert
	for rows.Next() {
		var alert models.TrackingAlert
		if err := rows.Scan(&alert.ID, &alert.PlayerName, &alert.DiscordID); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// getOrCreateServer inserts the server if missing and returns its id.
func getOrCreateServer(tx *sql.Tx, address string, port int, name, game string) (int64, error) {
	query := `
	INSERT INTO servers (address, port, name, game, first_seen, last_seen)
	VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
	ON CONFLICT(address, port) DO UPDATE SET
		name = excluded.name,
		game = excluded.game,
		last_seen = datetime('now');
	`
	_, err := tx.Exec(query, address, port, name, game)
	if err != nil {
		return 0, fmt.Errorf("failed to insert/update server: %w", err)
	}

	var id int64
	err = tx.QueryRow("SELECT id FROM servers WHERE address = ? AND port = ?", address, port).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get server id: %w", err)
	}

	return id, nil
}

// getOrCreatePlayer inserts the player if missing and returns its id.
func getOrCreatePlayer(tx *sql.Tx, name string) (int64, error) {
	query := `
	INSERT INTO players (name)
	VALUES (?)
	ON CONFLICT(name) DO NOTHING;
	`
	_, err := tx.Exec(query, name)
	if err != nil {
		return 0, fmt.Errorf("failed to insert player: %w", err)
	}

	var id int64
	err = tx.QueryRow("SELECT id FROM players WHERE name = ?", name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get player id: %w", err)
	}
	return id, nil
}

func getActiveServerSighting(address string, port int) (int64, error) {
	var sightingID int64
	query := `
	SELECT ss.id
	FROM server_sightings ss
	JOIN servers s ON ss.server_id = s.id
	WHERE s.address = ? AND s.port = ? AND ss.disconnected_at IS NULL
	LIMIT 1
	`
	err := DB.QueryRow(query, address, port).Scan(&sightingID)
	if err == sql.ErrNoRows {
		return 0, nil // no active sighting
	}
	if err != nil {
		return 0, err
	}
	return sightingID, nil
}

// startServerSightingIfNeeded starts a new sighting only if no active sighting exists and returns sighting ID.
func startServerSightingIfNeeded(address string, port int, name, game string) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	serverID, err := getOrCreateServer(tx, address, port, name, game)
	if err != nil {
		return 0, err
	}

	var sightingID int64
	err = tx.QueryRow(`
		SELECT id FROM server_sightings
		WHERE server_id = ? AND disconnected_at IS NULL
		LIMIT 1
	`, serverID).Scan(&sightingID)

	if err == sql.ErrNoRows {
		// no active sighting, create one
		res, err := tx.Exec(`
			INSERT INTO server_sightings (server_id, seen_at) VALUES (?, datetime('now'))
		`, serverID)
		if err != nil {
			return 0, err
		}
		sightingID, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return sightingID, nil
}

// stopServerSighting closes sighting and optionally closes all open player sightings for that server sighting.
func stopServerSightingAndPlayers(serverSightingID int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE server_sightings
		SET disconnected_at = datetime('now')
		WHERE id = ? AND disconnected_at IS NULL
	`, serverSightingID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE player_sightings
		SET disconnected_at = datetime('now')
		WHERE server_sighting_id = ? AND disconnected_at IS NULL
	`, serverSightingID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// startPlayerSighting creates a player sighting for given player and active server sighting.
func startPlayerSighting(address string, port int, playerName string) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	sightingID, err := getActiveServerSighting(address, port)
	if err != nil {
		return 0, err
	}
	if sightingID == 0 {
		return 0, fmt.Errorf("no active server sighting found for %s:%d", address, port)
	}

	playerID, err := getOrCreatePlayer(tx, playerName)
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO player_sightings (server_sighting_id, player_id, seen_at)
		VALUES (?, ?, datetime('now'))
	`, sightingID, playerID)
	if err != nil {
		return 0, err
	}

	playerSightingID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return playerSightingID, tx.Commit()
}

// stopPlayerSighting closes the player's sighting for current server sighting.
func stopPlayerSighting(address string, port int, playerName string) error {
	sightingID, err := getActiveServerSighting(address, port)
	if err != nil {
		return err
	}
	if sightingID == 0 {
		return fmt.Errorf("no active server sighting found for %s:%d", address, port)
	}

	var playerID int64
	err = DB.QueryRow("SELECT id FROM players WHERE name = ?", playerName).Scan(&playerID)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`
		UPDATE player_sightings
		SET disconnected_at = datetime('now')
		WHERE server_sighting_id = ? AND player_id = ? AND disconnected_at IS NULL
	`, sightingID, playerID)
	return err
}

func HandleEvent(event models.TrackingEvent) error {
	switch event.Type {
	case "serverOnline":
		_, err := startServerSightingIfNeeded(event.Server, event.Port, event.Name, event.Game)
		return err

	case "serverOffline":
		sightingID, err := getActiveServerSighting(event.Server, event.Port)
		if err != nil || sightingID == 0 {
			return err
		}
		return stopServerSightingAndPlayers(sightingID)

	case "playerJoin":
		_, err := startPlayerSighting(event.Server, event.Port, event.Player)
		return err

	case "playerLeave":
		return stopPlayerSighting(event.Server, event.Port, event.Player)
	}

	return nil
}

func GetPlayerHistory(name string) ([]models.PlayerSighting, error) {
	query := `
	SELECT ps.seen_at, ps.disconnected_at, s.address, s.port
	FROM player_sightings ps
	JOIN players p ON ps.player_id = p.id
	JOIN server_sightings ss ON ps.server_sighting_id = ss.id
	JOIN servers s ON ss.server_id = s.id
	WHERE LOWER(p.name) = LOWER(?)
	ORDER BY ps.seen_at DESC
	`
	rows, err := DB.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.PlayerSighting
	for rows.Next() {
		var event models.PlayerSighting
		event.Player = name
		var disconnectedAt sql.NullTime
		err := rows.Scan(&event.ConnectedAt, &disconnectedAt, &event.Address, &event.Port)
		if err != nil {
			return nil, err
		}
		if disconnectedAt.Valid {
			event.DisconnectedAt = &disconnectedAt.Time
		} else {
			event.DisconnectedAt = nil
		}
		history = append(history, event)
	}

	return history, nil
}

func GetServerHistory(address string, port int) ([]models.ServerSighting, error) {
	query := `
	SELECT ss.seen_at, ss.disconnected_at
	FROM server_sightings ss
	JOIN servers s ON ss.server_id = s.id
	WHERE s.address = ? AND s.port = ?
	ORDER BY ss.seen_at DESC
	`
	rows, err := DB.Query(query, address, port)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ServerSighting
	for rows.Next() {
		var sighting models.ServerSighting
		var disconnectedAt sql.NullTime
		err := rows.Scan(&sighting.SeenAt, &disconnectedAt)
		if err != nil {
			return nil, err
		}
		if disconnectedAt.Valid {
			sighting.DisconnectedAt = &disconnectedAt.Time
		} else {
			sighting.DisconnectedAt = nil
		}
		history = append(history, sighting)
	}

	return history, nil
}

func SaveSnapshot(snapshot models.Snapshot) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the snapshot timestamp
	res, err := tx.Exec(`INSERT INTO snapshots (timestamp) VALUES (?)`, snapshot.Time)
	if err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}
	snapshotID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get snapshot ID: %w", err)
	}

	// Insert each server in the snapshot
	for _, server := range snapshot.Servers {
		playerListJSON, err := json.Marshal(server.PlayerList)
		if err != nil {
			return fmt.Errorf("failed to marshal player list: %w", err)
		}

		_, err = tx.Exec(`
			INSERT INTO snapshot_servers 
			(snapshot_id, address, port, name, game, clients, player_list)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			snapshotID,
			server.Address,
			server.Port,
			server.Name,
			server.Game,
			server.Clients,
			string(playerListJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert snapshot server: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit snapshot: %w", err)
	}

	return nil
}

func GetSnapshotHistoryForServer(address string, port int) ([]models.Snapshot, error) {
	query := `
	SELECT snap.timestamp, s.name, s.game, s.mods, s.clients, s.player_list
	FROM snapshot_servers s
	JOIN snapshots snap ON s.snapshot_id = snap.id
	WHERE s.address = ? AND s.port = ?
	ORDER BY snap.timestamp DESC
	`

	rows, err := DB.Query(query, address, port)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var snapshots []models.Snapshot

	for rows.Next() {
		var snapshot models.Snapshot
		var server models.Server
		var playerListJSON string

		err := rows.Scan(
			&snapshot.Time,
			&server.Name,
			&server.Game,
			&server.Clients,
			&playerListJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		server.Address = address
		server.Port = port

		err = json.Unmarshal([]byte(playerListJSON), &server.PlayerList)
		if err != nil {
			return nil, fmt.Errorf("failed to parse player list JSON: %w", err)
		}

		snapshot.Servers = []models.Server{server}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func GetSnapshotByTime(t time.Time) (models.Snapshot, error) {
	query := `
	SELECT s.address, s.port, s.name, s.game, s.clients, s.player_list
	FROM snapshot_servers s
	JOIN snapshots snap ON s.snapshot_id = snap.id
	WHERE snap.timestamp = ?
	`

	rows, err := DB.Query(query, t)
	if err != nil {
		return models.Snapshot{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var servers []models.Server

	for rows.Next() {
		var server models.Server
		var playerListJSON string

		err := rows.Scan(
			&server.Address,
			&server.Port,
			&server.Name,
			&server.Game,
			&server.Clients,
			&playerListJSON,
		)
		if err != nil {
			return models.Snapshot{}, fmt.Errorf("row scan failed: %w", err)
		}

		err = json.Unmarshal([]byte(playerListJSON), &server.PlayerList)
		if err != nil {
			return models.Snapshot{}, fmt.Errorf("failed to parse player list JSON: %w", err)
		}

		servers = append(servers, server)
	}

	return models.Snapshot{
		Time:    t,
		Servers: servers,
	}, nil
}

func GetLatestSnapshot() (models.Snapshot, error) {
	var snapshotID int64
	var snapshotTime time.Time
	err := DB.QueryRow(`SELECT id, timestamp FROM snapshots ORDER BY timestamp DESC LIMIT 1`).Scan(&snapshotID, &snapshotTime)
	if err != nil {
		return models.Snapshot{}, fmt.Errorf("query failed: %w", err)
	}

	query := `
	SELECT address, port, name, game, clients, player_list
	FROM snapshot_servers
	WHERE snapshot_id = ?
	`

	rows, err := DB.Query(query, snapshotID)
	if err != nil {
		return models.Snapshot{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var servers []models.Server

	for rows.Next() {
		var server models.Server
		var playerListJSON string

		err := rows.Scan(
			&server.Address,
			&server.Port,
			&server.Name,
			&server.Game,
			&server.Clients,
			&playerListJSON,
		)
		if err != nil {
			return models.Snapshot{}, fmt.Errorf("row scan failed: %w", err)
		}

		err = json.Unmarshal([]byte(playerListJSON), &server.PlayerList)
		if err != nil {
			return models.Snapshot{}, fmt.Errorf("failed to parse player list JSON: %w", err)
		}

		servers = append(servers, server)
	}

	return models.Snapshot{
		Time:    snapshotTime,
		Servers: servers,
	}, nil
}

func GetServerInfo(address string, port int) (models.Server, error) {
	var server models.Server
	query := `
	SELECT name, game
	FROM servers
	WHERE address = ? AND port = ?
	`
	err := DB.QueryRow(query, address, port).Scan(&server.Name, &server.Game)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Server{}, fmt.Errorf("server not found: %s:%d", address, port)
		}
		return models.Server{}, fmt.Errorf("query failed: %w", err)
	}
	server.Address = address
	server.Port = port
	return server, nil
}
