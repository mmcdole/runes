package config

type Config struct {
	Client ClientConfig `koanf:"client"`
	Core   CoreConfig   `koanf:"core"`
	Server ServerConfig `koanf:"server"`
}

type CoreConfig struct {
	CommandPrefix                 string   `koanf:"commandPrefix"`
	CommandSeparator              string   `koanf:"commandSeparator"`
	IdleTime                      int      `koanf:"idleTime"`
	DisablePluginsOnIdle          bool     `koanf:"disablePluginsOnIdle"`
	DisablePluginsOnIdleWhitelist []string `koanf:"disablePluginsOnIdleWhitelist"`
}

type ClientConfig struct {
	Telnet    *TelnetClientConfig    `koanf:"telnet"`
	Websocket *WebsocketClientConfig `koanf:"websocket"`
}

type TelnetClientConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type WebsocketClientConfig struct {
	Port int `koanf:"port"`
}

type ServerConfig struct{}
