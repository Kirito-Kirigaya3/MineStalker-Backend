package config

import (
	"teamacedia/minestalker/internal/models"

	"gopkg.in/ini.v1"
)

// LoadConfig loads BotConfig from an INI file
func LoadConfig(path string) (*models.BotConfig, error) {
	cfgFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	cfg := &models.BotConfig{
		Token:   cfgFile.Section("").Key("Token").String(),
		AppID:   cfgFile.Section("").Key("AppID").String(),
		GuildID: cfgFile.Section("").Key("GuildID").String(),
	}

	return cfg, nil
}
