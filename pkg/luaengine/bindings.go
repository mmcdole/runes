package luaengine

import (
	"github.com/mmcdole/runes/pkg/events"
	lua "github.com/yuin/gopher-lua"
)

// luaBindings provides the Go implementations for functions exposed to Lua
type luaBindings struct {
	engine *LuaEngine
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
	if buffer == "" {
		buffer = "main"
	}
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
