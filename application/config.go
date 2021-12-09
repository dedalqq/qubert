package application

type Config struct {
	Debug          bool
	Host           string `json:"host"`
	Port           uint16 `json:"port"`
	SettingsFile   string `json:"settings-file"`
	HostBadgeColor string `json:"host-badge-color,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Debug:        false,
		Host:         "",
		Port:         8080,
		SettingsFile: "./settings.json",
	}
}
