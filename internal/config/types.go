package config

type Config struct {
	ConfigDir string
	Server    ServerConfig `koanf:"server"`
	Core      CoreConfig   `koanf:"core"`
	Proxy     ProxyConfig  `koanf:"proxy"`
}

type CoreConfig struct {
	EnableColors                  bool     `koanf:"enableColors"`
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
	SSL       *SSLServerConfig       `koanf:"ssl"`
}
type SSLServerConfig struct {
	Host         string `koanf:"host"`
	Port         int    `koanf:"port"`
	CertFile     string `koanf:"certFile"`
	KeyFile      string `koanf:"keyFile"`
	GeneratePair bool   `koanf:"generatePair"`
}

type TelnetServerConfig struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type WebsocketServerConfig struct {
	Port int `koanf:"port"`
}

type ProxyConfig struct{}
