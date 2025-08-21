package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"teamacedia/minestalker/internal/db"
	"teamacedia/minestalker/internal/models"

	"github.com/bwmarrin/discordgo"
)

var (
	session  *discordgo.Session
	cmdIDs   []*discordgo.ApplicationCommand
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "playertracker",
			Description: "Manage player tracking",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add a player to the tracker",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "playername",
							Description: "Name of the player",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove a player from the tracker",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "playername",
							Description: "Name of the player",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List tracked players",
				},
			},
		},
		{
			Name:        "servertracker",
			Description: "Manage server tracking",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Add a server to the tracker",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "address",
							Description: "Server address (IP or hostname)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "port",
							Description: "Server port",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove a server from the tracker",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "address",
							Description: "Server address (IP or hostname)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "port",
							Description: "Server port",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List tracked servers",
				},
			},
		},
		{
			Name:        "playerhistory",
			Description: "Get connection history for a player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "playername",
					Description: "Name of the player",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "page",
					Description: "Page number for paginated results (default 1)",
					Required:    false,
				},
			},
		},
	}
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	if data.Name == "playertracker" {
		if len(data.Options) == 0 {
			return
		}

		switch data.Options[0].Name {
		case "add":
			playerName := data.Options[0].Options[0].StringValue()
			discordId := i.Member.User.ID
			alert := models.TrackingAlert{
				PlayerName: playerName,
				DiscordID:  discordId,
			}

			if db.CheckIfTrackingAlertExists(alert) {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "You are already tracking player **" + playerName + "**.",
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}

			err := db.AddTrackingAlert(alert)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error adding player to tracker: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Player Added",
				Description: "Player **" + playerName + "** has been added to your tracker.\n You will receive DMs when they join or leave servers.",
				Color:       0x00FF00, // Green
			}
			replyEmbed(s, i, embed)
		case "remove":
			playerName := data.Options[0].Options[0].StringValue()
			discordId := i.Member.User.ID
			alert := models.TrackingAlert{
				PlayerName: playerName,
				DiscordID:  discordId,
			}

			if !db.CheckIfTrackingAlertExists(alert) {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "You are not tracking player **" + playerName + "**.",
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}

			err := db.RemoveTrackingAlert(alert)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error removing player from tracker: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Player Removed",
				Description: "Player **" + playerName + "** has been removed from your tracker.",
				Color:       0x00FF00, // Green
			}
			replyEmbed(s, i, embed)
		case "list":
			alerts, err := db.GetTrackingAlerts(i.Member.User.ID)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error retrieving tracked players: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			if len(alerts) == 0 {
				embed := &discordgo.MessageEmbed{
					Title:       "No Tracked Players",
					Description: "You are not currently tracking any players.",
					Color:       0xFFFF00, // Yellow
				}
				replyEmbed(s, i, embed)
				return
			}
			var response string
			response = "Currently tracked players:\n"
			for _, alert := range alerts {
				response += "- **" + alert.PlayerName + "**\n"
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Tracked Players",
				Description: response,
				Color:       0x00FFFF, // Cyan
			}
			replyEmbed(s, i, embed)
		default:
			reply(s, i, "Unknown subcommand")
			return
		}
	}

	if data.Name == "servertracker" {
		if len(data.Options) == 0 {
			return
		}

		switch data.Options[0].Name {
		case "add":
			address := data.Options[0].Options[0].StringValue()
			port := int(data.Options[0].Options[1].IntValue())
			alert := models.ServerTrackingAlert{
				ServerAddress: address,
				ServerPort:    port,
				DiscordID:     i.Member.User.ID,
			}

			if db.CheckIfServerTrackingAlertExists(alert) {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "You are already tracking server **" + address + ":" + fmt.Sprint(port) + "**.",
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}

			err := db.AddServerTrackingAlert(alert)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error adding server to tracker: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Server Added",
				Description: "Server **" + address + ":" + fmt.Sprint(port) + "** has been added to your tracker.\n You will receive DMs when the server starts or stops.",
				Color:       0x00FF00, // Green
			}
			replyEmbed(s, i, embed)
		case "remove":
			address := data.Options[0].Options[0].StringValue()
			port := int(data.Options[0].Options[1].IntValue())
			alert := models.ServerTrackingAlert{
				ServerAddress: address,
				ServerPort:    port,
				DiscordID:     i.Member.User.ID,
			}

			if !db.CheckIfServerTrackingAlertExists(alert) {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "You are not tracking server **" + address + ":" + fmt.Sprint(port) + "**.",
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			err := db.RemoveServerTrackingAlert(alert)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error removing server from tracker: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Server Removed",
				Description: "Server **" + address + ":" + fmt.Sprint(port) + "** has been removed from your tracker.",
				Color:       0x00FF00, // Green
			}
			replyEmbed(s, i, embed)
		case "list":
			alerts, err := db.GetServerTrackingAlerts(i.Member.User.ID)
			if err != nil {
				embed := &discordgo.MessageEmbed{
					Title:       "Error",
					Description: "Error retrieving tracked servers: " + err.Error(),
					Color:       0xFF0000, // Red
				}
				replyEmbed(s, i, embed)
				return
			}
			if len(alerts) == 0 {
				embed := &discordgo.MessageEmbed{
					Title:       "No Tracked Servers",
					Description: "You are not currently tracking any servers.",
					Color:       0xFFFF00, // Yellow
				}
				replyEmbed(s, i, embed)
				return
			}
			var response string
			response = "Currently tracked servers:\n"
			for _, alert := range alerts {
				server, err := db.GetServerInfo(alert.ServerAddress, alert.ServerPort)
				serverName := "Unknown"
				if server.Name != "" && err == nil {
					serverName = server.Name
				}
				response += "- **" + serverName + "** (" + alert.ServerAddress + ":" + fmt.Sprint(alert.ServerPort) + ")\n"
			}
			embed := &discordgo.MessageEmbed{
				Title:       "Tracked Servers",
				Description: response,
				Color:       0x00FFFF, // Cyan
			}
			replyEmbed(s, i, embed)
		default:
			reply(s, i, "Unknown subcommand")
			return
		}
	}

	if data.Name == "playerhistory" {
		playerName := data.Options[0].StringValue()
		page := 1
		if len(data.Options) > 1 {
			page = int(data.Options[1].IntValue())
			if page < 1 {
				page = 1
			}
		}

		history, err := db.GetPlayerHistory(playerName)
		if err != nil {
			embed := &discordgo.MessageEmbed{
				Title:       "Error",
				Description: "Error retrieving player history: " + err.Error(),
				Color:       0xFF0000, // Red
			}
			replyEmbed(s, i, embed)
			return
		}

		if len(history) == 0 {
			embed := &discordgo.MessageEmbed{
				Title:       "No History",
				Description: "No connection history found for player **" + playerName + "**.",
				Color:       0xFFFF00, // Yellow
			}
			replyEmbed(s, i, embed)
			return
		}

		// Pagination
		const pageSize = 5
		totalPages := (len(history) + pageSize - 1) / pageSize
		if page > totalPages {
			page = totalPages
		}

		start := len(history) - page*pageSize
		if start < 0 {
			start = 0
		}
		end := len(history) - (page-1)*pageSize
		if end > len(history) {
			end = len(history)
		}

		response := fmt.Sprintf("Connection history for player **%s** (Page %d/%d):\n", playerName, page, totalPages)
		for _, sighting := range history[start:end] {
			connectedTS := sighting.ConnectedAt.Unix()
			disconnected := "Still connected"
			if sighting.DisconnectedAt != nil {
				disconnected = fmt.Sprintf("<t:%d:f>", sighting.DisconnectedAt.Unix())
			}
			server, err := db.GetServerInfo(sighting.Address, sighting.Port)
			serverName := "Unknown"
			if server.Name != "" && err == nil {
				serverName = server.Name
			}
			response += fmt.Sprintf(
				"- Server: **%s** ( %s:%d )\n  Connected at: <t:%d:f>\n  Disconnected at: %s\n",
				serverName,
				sighting.Address,
				sighting.Port,
				connectedTS,
				disconnected,
			)
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Player Connection History",
			Description: response,
			Color:       0x00FFFF, // Cyan
		}
		replyEmbed(s, i, embed)
	}
}

func reply(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

func replyEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction with embed: %v", err)
	}
}

func DmUser(userID string, content string) error {
	// Create or fetch DM channel
	channel, err := session.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	// Send message to the DM channel ID
	_, err = session.ChannelMessageSend(channel.ID, content)
	if err != nil {
		return err
	}

	return nil
}

func DmUserEmbed(userID string, embed *discordgo.MessageEmbed) error {
	// Create or fetch DM channel
	channel, err := session.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	// Send message to the DM channel ID
	_, err = session.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		return err
	}

	return nil
}

// HandleEvents processes tracking events and sends DMs to users based on alerts
func HandleEvents(events []models.TrackingEvent) {
	alerts, err := db.GetAllTrackingAlerts()
	if err != nil {
		log.Printf("Error retrieving tracking alerts: %v", err)
		return
	}
	for _, event := range events {
		switch event.Type {
		case "playerJoin":
			for _, alert := range alerts {
				if alert.PlayerName == event.Player {
					embed := &discordgo.MessageEmbed{
						Title:       "Player Joined",
						Description: fmt.Sprintf("Player **%s** joined server **%s**", event.Player, event.Name),
						Color:       0x00FF00, // Green
					}
					err := DmUserEmbed(alert.DiscordID, embed)
					if err != nil {
						log.Printf("Error sending DM to %s: %v", alert.DiscordID, err)
					}
				}
			}
		case "playerLeave":
			for _, alert := range alerts {
				if alert.PlayerName == event.Player {
					embed := &discordgo.MessageEmbed{
						Title:       "Player Left",
						Description: fmt.Sprintf("Player **%s** left server **%s**", event.Player, event.Name),
						Color:       0xFF0000, // Red
					}
					err := DmUserEmbed(alert.DiscordID, embed)
					if err != nil {
						log.Printf("Error sending DM to %s: %v", alert.DiscordID, err)
					}
				}
			}
		default:
			continue
		}
	}

	// Handle server tracking alerts
	serverAlerts, err := db.GetAllServerTrackingAlerts()
	if err != nil {
		log.Printf("Error retrieving server tracking alerts: %v", err)
		return
	}
	for _, event := range events {
		switch event.Type {
		case "serverOnline":
			for _, alert := range serverAlerts {
				if alert.ServerAddress == event.Server && alert.ServerPort == event.Port {
					embed := &discordgo.MessageEmbed{
						Title:       "Server Online",
						Description: fmt.Sprintf("Server **%s** (%s:%d) is now **ONLINE** ✅", event.Name, event.Server, event.Port),
						Color:       0x00FF00, // Green
					}
					err := DmUserEmbed(alert.DiscordID, embed)
					if err != nil {
						log.Printf("Error sending DM to %s: %v", alert.DiscordID, err)
					}
				}
			}
		case "serverOffline":
			for _, alert := range serverAlerts {
				if alert.ServerAddress == event.Server && alert.ServerPort == event.Port {
					embed := &discordgo.MessageEmbed{
						Title:       "Server Offline",
						Description: fmt.Sprintf("Server **%s** (%s:%d) is now **OFFLINE** ❌", event.Name, event.Server, event.Port),
						Color:       0xFF0000, // Red
					}
					err := DmUserEmbed(alert.DiscordID, embed)
					if err != nil {
						log.Printf("Error sending DM to %s: %v", alert.DiscordID, err)
					}
				}
			}
		default:
			continue
		}
	}
}

// Start the bot, register commands, and block until SIGINT/SIGTERM
func Start(botToken string, appID string, guildID string) {
	var err error
	session, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	session.AddHandler(interactionHandler)

	// Open the session
	err = session.Open()
	if err != nil {
		log.Fatalf("Cannot open Discord session: %v", err)
	}
	log.Println("Discord bot is running...")

	// Register commands
	for _, v := range commands {
		cmd, err := session.ApplicationCommandCreate(appID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create '%s' command: %v", v.Name, err)
		}
		cmdIDs = append(cmdIDs, cmd)
	}
	log.Println("Commands registered.")

	// Wait here until Ctrl+C or kill signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Cleanup before exit
	log.Println("Shutting down, removing commands...")
	for _, cmd := range cmdIDs {
		err := session.ApplicationCommandDelete(appID, guildID, cmd.ID)
		if err != nil {
			log.Printf("Cannot delete command %s: %v", cmd.Name, err)
		}
	}
	session.Close()
	log.Println("Bot stopped cleanly.")
}
