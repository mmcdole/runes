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
	
	// Create runes table
	mt := L.NewTable()
	L.SetGlobal("runes", mt)

	// Register functions
	L.SetFuncs(mt, map[string]lua.LGFunction{
		// Core system bindings
		"debug":         bindings.debug,
		"version":       bindings.version,
		"log":           bindings.log,

		// Connection bindings
		"connect":       bindings.connect,
		"disconnect":    bindings.disconnect,

		// Output/Input bindings
		"output":        bindings.output,
		"prompt":        bindings.prompt,
		"send_raw":      bindings.sendCommand,

		// Line processing
		"add_output_listener": bindings.addOutputListener,
		"add_input_listener":  bindings.addInputListener,

		// Buffer management
		"list_buffers":  bindings.listBuffers,
		"switch_buffer": bindings.switchBuffer,

		// Script management
		"load_script":   bindings.loadScript,
		"quit":          bindings.quit,
	})
	return nil
}

// Core system bindings
func (b *luaBindings) debug(L *lua.LState) int {
	text := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventDebug,
		Data: text,
	})
	return 0
}

func (b *luaBindings) version(L *lua.LState) int {
	L.Push(lua.LString("1.0.0"))
	return 1
}

func (b *luaBindings) log(L *lua.LState) int {
	text := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventLog,
		Data: text,
	})
	return 0
}

// Connection bindings
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

func (b *luaBindings) disconnect(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventDisconnect,
	})
	return 0
}

// Output/Input bindings
func (b *luaBindings) output(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewLine(text)
	if processed := b.engine.ProcessOutput(line); processed != nil {
		b.engine.eventSystem.Emit(events.Event{
			Type: events.EventRedraw,
			Data: processed,
		})
	}
	return 0
}

func (b *luaBindings) prompt(L *lua.LState) int {
	text := L.ToString(1)
	line := types.NewPrompt(text)
	if processed := b.engine.ProcessOutput(line); processed != nil {
		b.engine.eventSystem.Emit(events.Event{
			Type: events.EventRedraw,
			Data: processed,
		})
	}
	return 0
}

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

// Line processing
func (b *luaBindings) addOutputListener(L *lua.LState) int {
	fn := L.CheckFunction(1)
	b.engine.AddOutputProcessor(fn)
	return 0
}

func (b *luaBindings) addInputListener(L *lua.LState) int {
	fn := L.CheckFunction(1)
	b.engine.AddInputProcessor(fn)
	return 0
}

// Buffer management
func (b *luaBindings) listBuffers(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventListBuffers,
	})
	return 0
}

func (b *luaBindings) switchBuffer(L *lua.LState) int {
	name := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventSwitchBuffer,
		Data: name,
	})
	return 0
}

// Script management
func (b *luaBindings) loadScript(L *lua.LState) int {
	path := L.ToString(1)
	if err := b.engine.loadUserScript(path); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func (b *luaBindings) quit(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventQuit,
	})
	return 0
}

const (
	OutputListenerTable = "outputListeners"
	InputListenerTable  = "inputListeners"
)
