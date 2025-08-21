package models

import "time"

type Server struct {
	Address    string   `json:"address"`
	Port       int      `json:"port"`
	Name       string   `json:"name"`
	Game       string   `json:"gameid"`
	Clients    int      `json:"clients"`
	PlayerList []string `json:"clients_list"`
}

type ServerListResponse struct {
	List []Server `json:"list"`
}

type TrackingEvent struct {
	Type      string // "playerJoin, playerLeave, serverOnline, serverOffline"
	Server    string
	Port      int
	Timestamp time.Time
	Player    string // Optional, for player events only
	Game      string // Optional, for server events only
	Name      string
}

type PlayerSighting struct {
	Address        string
	Port           int
	Player         string
	ConnectedAt    time.Time
	DisconnectedAt *time.Time // Optional, nil if still connected
}

type ServerSighting struct {
	SeenAt         time.Time
	DisconnectedAt *time.Time // Optional, nil if still connected
}

type Snapshot struct {
	Servers []Server
	Time    time.Time
}

type TrackingAlert struct {
	ID         int
	PlayerName string
	DiscordID  string
}

type ServerTrackingAlert struct {
	ID            int
	ServerAddress string
	ServerPort    int
	DiscordID     string
}

type Config struct {
	Token            string
	AppID            string
	GuildID          string
	UpdateInterval   int // Interval in seconds for periodic updates
	SnapshotInterval int // Interval in seconds for snapshot updates
}
