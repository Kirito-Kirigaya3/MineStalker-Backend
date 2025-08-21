package config

import (
	"teamacedia/minestalker/internal/models"

	"gopkg.in/ini.v1"
)

// LoadConfig loads Config from an INI file
func LoadConfig(path string) (*models.Config, error) {
	cfgFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	cfg := &models.Config{
		Token:            cfgFile.Section("").Key("Token").String(),
		AppID:            cfgFile.Section("").Key("AppID").String(),
		GuildID:          cfgFile.Section("").Key("GuildID").String(),
		UpdateInterval:   cfgFile.Section("").Key("UpdateInterval").MustInt(5),
		SnapshotInterval: cfgFile.Section("").Key("SnapshotInterval").MustInt(300),
		LoggerWebhookUrl: cfgFile.Section("").Key("LoggerWebhookUrl").String(),
	}

	return cfg, nil
}
