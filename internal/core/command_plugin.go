package core

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type PluginCommand struct{}

func (c *PluginCommand) Execute(params *CommandParams) bool {
	if len(params.Args) == 0 {
		return c.handlePluginListCommand(params)
	}

	action := strings.ToLower(params.Args[0])
	switch action {
	case "list":
		return c.handlePluginListCommand(params)
	case "load":
		return c.handlePluginLoadCommand(params)
	case "unload":
		return c.handlePluginUnloadCommand(params)
	case "disable":
		return c.handlePluginDisableCommand(params)
	case "enable":
		return c.handlePluginEnableCommand(params)
	default:
		return false
	}
}

func (c *PluginCommand) handlePluginListCommand(params *CommandParams) bool {
	params.writeToExecutor("Plugins:\n")
	plugins := params.Session.pluginEngine.GetPlugins()
	for _, plugin := range plugins {
		if plugin.IsActive {
			params.writeToExecutor(fmt.Sprintf(" [%s] %s\n", color.HiGreenString("on"), plugin.Name))
		} else {
			params.writeToExecutor(fmt.Sprintf(" [%s] %s\n", color.HiRedString("off"), plugin.Name))
		}
	}
	return true
}

func (c *PluginCommand) handlePluginLoadCommand(params *CommandParams) bool {
	// TODO: implement subcommand
	return false
}

func (c *PluginCommand) handlePluginUnloadCommand(params *CommandParams) bool {
	// TODO: implement subcommand
	return false
}

func (c *PluginCommand) handlePluginDisableCommand(params *CommandParams) bool {
	// TODO: implement subcommand
	return false
}

func (c *PluginCommand) handlePluginEnableCommand(params *CommandParams) bool {
	// TODO: implement subcommand
	return false
}

func (c *PluginCommand) Usage() string {
	usage := `plugin list - Show info about the loaded plugins 
 plugin load <file> - Load / Reload a plugin with the provided path 
 plugin unload <file> - Unload a plugin with the provided path
 plugin disable <file> - Disable a plugin with the provided path
 plugin enable <file> - Enable a plugin with the provided path`
	return usage
}

func (c *PluginCommand) Help() string {
	// TODO: Plugin Help
	return "Plugin Help!"
}
