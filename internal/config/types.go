package config

type Config struct {
	Server ServerConfig `koanf:"server"`
	Core   CoreConfig   `koanf:"core"`
	Proxy  ProxyConfig  `koanf:"proxy"`
}

type CoreConfig struct {
	BufferSize                    int      `koanf:"bufferSize"`
	BufferReplaySize              int      `koanf:"bufferReplaySize"`
	CommandPrefix                 string   `koanf:"commandPrefix"`
	CommandSeparator              string   `koanf:"commandSeparator"`
	IdleTime                      int      `koanf:"idleTime"`
	DisablePluginsOnIdle          bool     `koanf:"disablePluginsOnIdle"`
	DisablePluginsOnIdleWhitelist []string `koanf:"disablePluginsOnIdleWhitelist"`
}

type ServerConfig struct {
	Telnet    *TelnetServerConfig    `koanf:"telnet"`
	Websocket *WebsocketServerConfig `koanf:"websocket"`
}

type TelnetServerConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type WebsocketServerConfig struct {
	Port int `koanf:"port"`
}

type ProxyConfig struct{}
