package config

import "os"

type Config struct {
	DiscordToken     string
	GuildID          string
	ControlChannelID string
	ModRoleID        string
	CategoryID       string
}

func LoadConfig() Config {
	return Config{
		DiscordToken:     os.Getenv("DISCORD_TOKEN"),
		GuildID:          os.Getenv("DISCORD_GUILD_ID"),
		ControlChannelID: os.Getenv("CONTROL_CHANNEL_ID"),
		ModRoleID:        os.Getenv("MOD_ROLE_ID"),
		CategoryID:       os.Getenv("TICKET_CATEGORY_ID"),
	}
}
