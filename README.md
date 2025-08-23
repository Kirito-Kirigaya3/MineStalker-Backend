# MineStalker

MineStalker is a backend service for tracking players and servers on Minetest public servers.  
It provides a REST API for querying player history, server history, and live snapshots of the server list.  
Additionally, it integrates with a Discord bot for real-time notifications about player activity and server status.

---

## Features

- **Player Tracking:** Track which servers players have connected to and when.  
- **Server Tracking:** Monitor online/offline status and connected players for public servers.  
- **Live Snapshot:** Fetch a snapshot of all public servers with their current players.  
- **Discord Integration:** Optional bot for sending join/leave and server status notifications.  
- **Configurable Scraping:** Scheduler scrapes the server list at configurable intervals.  

---

## Requirements

- Go 1.20+  
- SQLite3  
- Discord bot token (optional, for Discord notifications)  

---

## Setup

1. **Clone the repository**
```bash
git clone https://github.com/TeamAcedia/MineStalker-Backend.git
cd MineStalker-Backend
````

This creates `minestalker.db` with all necessary tables.

2. **Configure the bot**

Create a `config.ini` file in the root directory:

```ini
Token = YOUR_BOT_TOKEN
AppID = YOUR_APP_ID
GuildID = YOUR_GUILD_ID
UpdateInterval = 5
SnapshotInterval = 300
LoggerWebhookURL = LOGGER_WEBHOOK_URL
LoggerWebhookUsername = USERNAME_TO_SHOW_AS_WHEN_LOGGING_VIA_WEBHOOK
```

3. **Run the server**

```bash
./build_and_run.bat
```

The backend will start on `:8080`. The scraper and Discord bot run in the background.

---

## API Endpoints

| Endpoint                  | Description                                |
| ------------------------- | ------------------------------------------ |
| `/api/player/{name}`      | Get the history of a player across servers |
| `/api/server/{ip}/{port}` | Get history of a server including players  |
| `/api/snapshot`           | Get a snapshot of current public servers   |

---

## Discord Bot Commands

* `/playertracker add <playername>` – Track a player
* `/playertracker remove <playername>` – Stop tracking a player
* `/playertracker list` – List all tracked players

* `/servertracker add <playername>` – Track a server
* `/servertracker remove <playername>` – Stop tracking a server
* `/servertracker list` – List all tracked servers

* `/playerhistory <playername> <page>` – List all tracked activity of a specific player

The bot sends notifications when players join/leave servers or when servers go online/offline.
The notifications are sent to the dms of the user who runs those commands, and they can be sent to multiple users if each one adds it to their tracker.

---

## Development

* The scraper runs in the background every 5 seconds by default.
* It saves snapshots of the server list every 5 minutes.
* Database is SQLite for simplicity and portability.
* Go modules are used for dependency management.

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature-name`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature-name`)
5. Open a Pull Request

---

## License

MIT License © 2025 Team Acedia

---

## Contact

For issues, questions, or suggestions, contact the project maintainers via GitHub issues or our Discord, https://discord.gg/invite/z2bW5h6PXQ