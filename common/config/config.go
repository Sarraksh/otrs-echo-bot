package config

// Combine all available options.
type Config struct {
	OTRS     OTRSConf
	Telegram TelegramConf
}

// Options for OTRS module.
type OTRSConf struct {
	Host            string // Host on which OTRS is located.
	TicketURLPrefix string // Prefix for ticket URL for browser (include port if not default).
	API             OTRSAPI
}

// OTRS API configuration.
type OTRSAPI struct {
	Login                   string // Login for API.
	Password                string // Password for API.
	Protocol                string // Protocol over which the API is available. http or https.
	Port                    string // Port over which the API is available.
	InsecureConnection      bool   // If true allow insecure connections to API.
	GetTicketDetailListPath string // Get ticket details.
}

// Options for Telegram module.
type TelegramConf struct {
	Token string // Token from @BotFather.
}
