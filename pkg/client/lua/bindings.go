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
func registerBindings(L *lua.LState, engine *LuaEngine) error {
	bindings := newLuaBindings(engine)
	
	// Register types
	RegisterLine(L)
	
	// Create runes table
	mt := L.NewTable()
	L.SetGlobal("runes", mt)

	// Register functions
	L.SetFuncs(mt, map[string]lua.LGFunction{
		// Core system bindings
		"_debug":         bindings.debug,
		"_version":       bindings.version,
		"_log":           bindings.log,

		// Connection bindings
		"_connect":       bindings.connect,
		"_disconnect":    bindings.disconnect,

		// Output/Input bindings
		"_output":        bindings.output,
		"_prompt":        bindings.prompt,
		"_send":          bindings.send,
		"_send_raw":      bindings.sendCommand,

		// Processor bindings
		"_add_output_processor": bindings.addOutputProcessor,
		"_add_input_processor":  bindings.addInputProcessor,

		// Buffer management
		"_list_buffers":  bindings.listBuffers,
		"_switch_buffer": bindings.switchBuffer,

		// Script management
		"_load_script":   bindings.loadScript,
		"_quit":          bindings.quit,
	})
	return nil
}

// Core system bindings

// debug prints a debug message to the console
// Lua syntax: runes._debug(text)
func (b *luaBindings) debug(L *lua.LState) int {
	text := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventDebug,
		Data: text,
	})
	return 0
}

// version returns the version of Runes
// Lua syntax: runes._version() -> string
func (b *luaBindings) version(L *lua.LState) int {
	L.Push(lua.LString("1.0.0"))
	return 1
}

// log prints a log message to the console
// Lua syntax: runes._log(text)
func (b *luaBindings) log(L *lua.LState) int {
	text := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventLog,
		Data: text,
	})
	return 0
}

// Connection bindings

// connect connects to a server
// Lua syntax: runes._connect(host, port)
func (b *luaBindings) connect(L *lua.LState) int {
	host := L.ToString(1)
	port := L.ToInt(2)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventConnect,
		Data: struct {
			Host string
			Port int
		}{host, port},
	})
	return 0
}

// disconnect disconnects from the server
// Lua syntax: runes._disconnect()
func (b *luaBindings) disconnect(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventDisconnect,
	})
	return 0
}

// Output/Input bindings

// output sends text to the output window
// Lua syntax: runes._output(text)
func (b *luaBindings) output(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewLine(text)
	
	// Process the line through output processors
	if processed := b.engine.ProcessOutput(line); processed != nil {
		// Add to the output buffer instead of emitting directly
		b.engine.addOutputLine(processed)
	}
	return 0
}

// prompt sends a prompt to the output window
// Lua syntax: runes._prompt(text)
func (b *luaBindings) prompt(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewPrompt(text)
	
	// Process the line through output processors
	if processed := b.engine.ProcessOutput(line); processed != nil {
		// Add to the output buffer instead of emitting directly
		b.engine.addOutputLine(processed)
	}
	return 0
}

// send sends text to the server
// Lua syntax: runes._send(text)
func (b *luaBindings) send(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewClientLine(text)
	if processed := b.engine.ProcessInput(line); processed != nil {
		b.engine.eventSystem.Emit(events.Event{
			Type: events.EventRedraw,
			Data: processed,
		})
	}
	return 0
}

// sendCommand sends a raw command to the server
// Lua syntax: runes._send_raw(text)
func (b *luaBindings) sendCommand(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewClientLine(text)
	if processed := b.engine.ProcessInput(line); processed != nil {
		b.engine.eventSystem.Emit(events.Event{
			Type: events.EventRedraw,
			Data: processed,
		})
	}
	return 0
}

// Buffer management

// listBuffers lists all buffers
// Lua syntax: runes._list_buffers()
func (b *luaBindings) listBuffers(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventListBuffers,
	})
	return 0
}

// switchBuffer switches to a buffer
// Lua syntax: runes._switch_buffer(name)
func (b *luaBindings) switchBuffer(L *lua.LState) int {
	name := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventSwitchBuffer,
		Data: name,
	})
	return 0
}

// Processor bindings

// addOutputProcessor adds an output processor
// Lua syntax: runes._add_output_processor(function(line) ... end)
func (b *luaBindings) addOutputProcessor(L *lua.LState) int {
	fn := L.CheckFunction(1)
	b.engine.addOutputProcessor(fn)
	return 0
}

// addInputProcessor adds an input processor
// Lua syntax: runes._add_input_processor(function(line) ... end)
func (b *luaBindings) addInputProcessor(L *lua.LState) int {
	fn := L.CheckFunction(1)
	b.engine.addInputProcessor(fn)
	return 0
}

// Script management

// loadScript loads a script
// Lua syntax: runes._load_script(path) -> error or nil
func (b *luaBindings) loadScript(L *lua.LState) int {
	path := L.ToString(1)
	if err := b.engine.loadUserScript(path); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

// quit quits the application
// Lua syntax: runes._quit()
func (b *luaBindings) quit(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventQuit,
	})
	return 0
}
