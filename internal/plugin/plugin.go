package plugin

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

const luaLibraryFile = `-- alias a command, or function so it can be called by another name 
function alias(cmd, arg2)
    if type(arg2) == "function" then
        -- argument is a function, register it as an alias
        registerAlias(cmd, arg2)
    else
        -- argument is a string, convert it to a function that calls send
        fn = function()
            send(arg2)
        end
        registerAlias(cmd, fn)
    end
end`

type Plugin struct {
	Name    string
	Path    string
	State   *lua.LState
	Engine  *PluginEngine
	Actions map[string]*lua.LFunction
	Aliases map[string]*lua.LFunction
	Events  map[PluginEvent]*lua.LFunction
}

func NewPlugin(name string, path string, engine *PluginEngine) *Plugin {
	return &Plugin{
		Name:   name,
		Path:   path,
		Engine: engine,
	}
}

func (p *Plugin) Load() error {
	// Setup alias/action/event handlers
	p.Actions = make(map[string]*lua.LFunction)
	p.Aliases = make(map[string]*lua.LFunction)
	p.Events = make(map[PluginEvent]*lua.LFunction) // TODO: Confirm we want single event handler per plugin?

	// Close out any previous Lua state before a reload
	if p.State != nil {
		p.State.Close()
	}

	// Create a new Lua state
	p.State = lua.NewState()

	// Register the Go functions as Lua functions
	p.State.SetGlobal("send", p.State.NewFunction(p.pluginSend))
	// p.State.SetGlobal("registerAction", p.registerAction)
	p.State.SetGlobal("registerAlias", p.State.NewFunction(p.pluginAlias))
	// p.State.SetGlobal("unregisterAction", p.unregisterAction)
	// p.State.SetGlobal("unregisterAlias", p.unregisterAlias)

	// Load the lua library file
	if err := p.State.DoString(luaLibraryFile); err != nil {
		p.State.Close()
		return err
	}

	// Load the Lua plugin file
	if err := p.State.DoFile(p.Path); err != nil {
		p.State.Close()
		return err
	}

	return nil
}

func (p *Plugin) CheckAndExecuteAlias(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if fn, ok := p.Aliases[cmd]; ok {
		p.State.Push(fn)
		p.State.PCall(0, 0, nil) // TODO: handle error?
		return true
	}
	return false
}

func (p *Plugin) CheckAndExecuteAction(text string) bool {
	return false
}

func (p *Plugin) pluginSend(state *lua.LState) int {
	str := p.State.CheckString(1)
	p.Engine.handlePluginSend(str + "\n")
	return 0
}

func (p *Plugin) pluginAlias(state *lua.LState) int {
	// get the alias name
	name := state.ToString(1)
	// get the function
	fn := state.ToFunction(2)

	// register the alias
	p.Aliases[name] = fn

	return 0
}
