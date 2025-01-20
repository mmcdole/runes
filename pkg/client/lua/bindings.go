package lua

import (
	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	lua "github.com/yuin/gopher-lua"
)

// luaBindings provides the Lua API for Runes
type luaBindings struct {
	engine *LuaEngine
}

func newLuaBindings(engine *LuaEngine) *luaBindings {
	return &luaBindings{engine: engine}
}

// register registers all bindings with Lua
func (b *luaBindings) register(L *lua.LState) {
	// Create runes table
	mt := L.NewTable()
	L.SetGlobal("runes", mt)

	// Register functions
	L.SetFuncs(mt, map[string]lua.LGFunction{
		"output":             b.output,
		"prompt":             b.prompt,
		"add_output_listener": b.addOutputListener,
		"add_input_listener":  b.addInputListener,
		"on_connect":         b.onConnect,
		"on_disconnect":      b.onDisconnect,
		"on_reset":          b.onReset,
		"quit":              b.quit,
	})
}

// output emits an output event
func (b *luaBindings) output(L *lua.LState) int {
	text := L.CheckString(1)
	line := types.NewLine(text)
	
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventOutput,
		Data: line,
	})
	return 0
}

// prompt emits a prompt event
func (b *luaBindings) prompt(L *lua.LState) int {
	text := L.CheckString(1)
	line := types.NewLine(text)
	line.IsPrompt = true
	
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventPrompt,
		Data: line,
	})
	return 0
}

// addOutputListener adds a function to process output lines
func (b *luaBindings) addOutputListener(L *lua.LState) int {
	fn := L.CheckFunction(1)
	
	// Get output listener table from registry
	table := L.GetField(L.Get(lua.RegistryIndex).(*lua.LTable), OutputListenerTable).(*lua.LTable)
	
	// Add function to table with next available index
	table.Append(fn)
	
	return 0
}

// addInputListener adds a function to process input lines
func (b *luaBindings) addInputListener(L *lua.LState) int {
	fn := L.CheckFunction(1)
	
	// Get input listener table from registry
	table := L.GetField(L.Get(lua.RegistryIndex).(*lua.LTable), InputListenerTable).(*lua.LTable)
	
	// Add function to table with next available index
	table.Append(fn)
	
	return 0
}

// onConnect adds a connect handler
func (b *luaBindings) onConnect(L *lua.LState) int {
	fn := L.CheckFunction(1)
	
	// Get connect handler table from registry
	table := L.GetField(L.Get(lua.RegistryIndex).(*lua.LTable), OnConnectTable).(*lua.LTable)
	
	// Add function to table with next available index
	table.Append(fn)
	
	return 0
}

// onDisconnect adds a disconnect handler
func (b *luaBindings) onDisconnect(L *lua.LState) int {
	fn := L.CheckFunction(1)
	
	// Get disconnect handler table from registry
	table := L.GetField(L.Get(lua.RegistryIndex).(*lua.LTable), OnDisconnectTable).(*lua.LTable)
	
	// Add function to table with next available index
	table.Append(fn)
	
	return 0
}

// onReset adds a reset handler
func (b *luaBindings) onReset(L *lua.LState) int {
	fn := L.CheckFunction(1)
	
	// Get reset handler table from registry
	table := L.GetField(L.Get(lua.RegistryIndex).(*lua.LTable), ScriptResetTable).(*lua.LTable)
	
	// Add function to table with next available index
	table.Append(fn)
	
	return 0
}

// quit exits the application
func (b *luaBindings) quit(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventQuit,
	})
	return 0
}

const (
	OutputListenerTable = "outputListeners"
	InputListenerTable  = "inputListeners"
	OnConnectTable      = "connectHandlers"
	OnDisconnectTable   = "disconnectHandlers"
	ScriptResetTable    = "resetHandlers"
)
