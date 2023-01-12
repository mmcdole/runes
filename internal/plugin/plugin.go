package plugin

import (
	lua "github.com/yuin/gopher-lua"
)

const luaLibraryFile = `-- alias a command, or function so it can be called by another name 
function alias(cmd, arg2)
    if type(arg2) == "function" then
        -- argument is a function, register it as an alias
        registerAlias(cmd, arg2)
    else
        -- argument is a string, convert it to a function that calls send
        simpleCmdText = arg2
        fn = function()
            send(simpleCmdText)
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
}

func NewPlugin(name string, path string, engine *PluginEngine) *Plugin {
	return &Plugin{
		Name:   name,
		Path:   path,
		Engine: engine,
	}
}

func (p *Plugin) Load() error {
	// Create a new Lua state
	p.State = lua.NewState()

	// load the lua library file
	if err := p.State.DoString(luaLibraryFile); err != nil {
		p.State.Close()
		return err
	}

	// Load the Lua plugin file
	if err := p.State.DoFile(p.Path); err != nil {
		p.State.Close()
		return err
	}

	p.Actions = make(map[string]*lua.LFunction)
	p.Aliases = make(map[string]*lua.LFunction)

	// register the Go functions as Lua functions
	// p.State.SetGlobal("send", p.send)
	// p.State.SetGlobal("registerAction", p.registerAction)
	p.State.SetGlobal("registerAlias", p.State.NewFunction(p.registerAlias))
	// p.State.SetGlobal("unregisterAction", p.unregisterAction)
	// p.State.SetGlobal("unregisterAlias", p.unregisterAlias)

	return nil
}

func (p *Plugin) send(state *lua.LState) int {
	return 1
}

func (p *Plugin) registerAction(state *lua.LState) int {
	return 1
}

func (p *Plugin) registerAlias(state *lua.LState) int {
	// get the alias name
	name := state.ToString(1)
	// get the function
	fn := state.ToFunction(2)

	// register the alias
	p.Aliases[name] = fn

	return 1
}
