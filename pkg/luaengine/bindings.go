package luaengine

import (
	"github.com/mmcdole/runes/pkg/events"
	lua "github.com/yuin/gopher-lua"
)

// luaBindings provides the Go implementations for functions exposed to Lua
type luaBindings struct {
	engine *LuaEngine
}

// getBindingsMap returns a map of all Lua function bindings
func (b *luaBindings) getBindingsMap() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"connect":       b.connect,
		"disconnect":    b.disconnect,
		"output":        b.output,
		"log":           b.log,
		"debug":         b.debug,
		"version":       b.version,
		"list_buffers":  b.listBuffers,
		"switch_buffer": b.switchBuffer,
		"send_raw":      b.sendCommand,
		"quit":          b.quit,
		"load_script":   b.loadScript,
	}
}

// Core system bindings
func (b *luaBindings) debug(L *lua.LState) int {
	text := L.ToString(1)
	// b.engine.eventSystem.Emit(events.Event{
	// 	Type: events.EventDebug,
	// 	Data: text,
	// })
	println(text)
	return 0
}

func (b *luaBindings) version(L *lua.LState) int {
	L.Push(lua.LString("1.0.0"))
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

// Output bindings
func (b *luaBindings) output(L *lua.LState) int {
	text := L.ToString(1)
	buffer := L.ToString(2)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventOutput,
		Data: struct {
			Text   string
			Buffer string
		}{text, buffer},
	})
	return 0
}

func (b *luaBindings) log(L *lua.LState) int {
	text := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventLog,
		Data: text,
	})
	return 0
}

// Command bindings
func (b *luaBindings) sendCommand(L *lua.LState) int {
	command := L.ToString(1)
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventCommand,
		Data: command,
	})
	return 0
}

// Buffer management bindings
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

// Client lifecycle bindings
func (b *luaBindings) quit(L *lua.LState) int {
	b.engine.eventSystem.Emit(events.Event{
		Type: events.EventQuit,
	})
	return 0
}

// Script loading binding
func (b *luaBindings) loadScript(L *lua.LState) int {
	path := L.ToString(1)
	err := b.engine.loadUserScript(path)
	if err != nil {
		L.Push(lua.LBool(false))
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}
