package config

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

func LoadOrCreateConfig() *Config {
	// Use the XDG config directory & create subdirs as needed
	configFile, err := xdg.ConfigFile("runes/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	k := koanf.New(".")

	// Set the defaults
	k.Load(confmap.Provider(map[string]interface{}{
		"client.telnet.port":                 "2000",
		"client.telnet.host":                 "",
		"core.commandPrefix":                 "!",
		"core.commandSeparator":              ";",
		"core.idleTime":                      300,
		"core.disablePluginsOnIdle":          true,
		"core.disablePluginsOnIdleWhitelist": []string{},
	}, "."), nil)

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config file if it does not exist
		defaultConfig, err := k.Marshal(yaml.Parser())
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(configFile, defaultConfig, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Merge config file settings with the defaults, if it exists
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			log.Fatalf("Failed to load config file: %v", err)
		}
	}

	conf := &Config{}
	if err := k.Unmarshal("", &conf); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	return conf
}
