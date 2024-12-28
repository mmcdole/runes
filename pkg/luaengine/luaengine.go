package luaengine

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mmcdole/runes/pkg/events"
	lua "github.com/yuin/gopher-lua"
)

//go:embed core/*.lua
var coreLuaScripts embed.FS

type LuaEngine struct {
	L             *lua.LState
	coreScriptsFS fs.FS
	userScriptDir string
	eventSystem   *events.EventProcessor
	bindings      *luaBindings
}

func New(coreScriptsFS fs.FS, userScriptDir string, eventSystem *events.EventProcessor) *LuaEngine {
	engine := &LuaEngine{
		L:             lua.NewState(),
		coreScriptsFS: coreScriptsFS,
		userScriptDir: userScriptDir,
		eventSystem:   eventSystem,
	}

	engine.bindings = &luaBindings{engine: engine}

	// Subscribe to raw events that need Lua processing
	eventSystem.Subscribe(events.EventRawInput, engine.handleRawInput)
	eventSystem.Subscribe(events.EventRawOutput, engine.handleRawOutput)

	return engine
}

func (engine *LuaEngine) Initialize() error {
	L := engine.L
	runesTable := L.NewTable()
	L.SetGlobal("runes", runesTable)

	// Register all bindings
	bindings := map[string]lua.LGFunction{
		"connect":       engine.bindings.connect,
		"disconnect":    engine.bindings.disconnect,
		"output":        engine.bindings.output,
		"log":           engine.bindings.log,
		"debug":         engine.bindings.debug,
		"version":       engine.bindings.version,
		"list_buffers":  engine.bindings.listBuffers,
		"switch_buffer": engine.bindings.switchBuffer,
		"sendRaw":       engine.bindings.sendCommand,
	}

	for name, fn := range bindings {
		L.SetField(runesTable, name, L.NewFunction(fn))
	}

	// Load core modules in dependency order
	coreModules := []struct {
		name string
		path string
	}{
		{"events", "core/events.lua"},
		{"command", "core/command.lua"},
		{"alias", "core/alias.lua"},
		{"trigger", "core/trigger.lua"},
		{"mud", "core/mud.lua"},
	}

	for _, module := range coreModules {
		content, err := fs.ReadFile(engine.coreScriptsFS, module.path)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", module.path, err)
		}
		if err := engine.L.DoString(string(content)); err != nil {
			return fmt.Errorf("error executing %s: %w", module.path, err)
		}
	}

	return engine.loadUserScripts()
}

// Close cleans up the Lua state
func (engine *LuaEngine) Close() {
	engine.L.Close()
}

// Raw event handlers that bridge between Go events and Lua events
func (engine *LuaEngine) handleRawInput(event events.Event) {
	engine.emitLuaEvent("input", event.Data.(string))
}

func (engine *LuaEngine) handleRawOutput(event events.Event) {
	engine.emitLuaEvent("output", event.Data.(string))
}

// emitLuaEvent sends an event to the Lua event system
func (engine *LuaEngine) emitLuaEvent(eventName string, eventData string) {
	L := engine.L
	eventsTable := L.GetGlobal("events")
	if eventsTable.Type() == lua.LTNil {
		fmt.Printf("[ERROR] events table is nil\n")
		return
	}

	emitFn := L.GetField(eventsTable, "emit")
	if emitFn.Type() == lua.LTNil {
		fmt.Printf("[ERROR] events.emit function is nil\n")
		return
	}

	fmt.Printf("[DEBUG] Calling Lua events.emit(%q, %q)\n", eventName, eventData)
	L.Push(emitFn)
	L.Push(lua.LString(eventName))
	L.Push(lua.LString(eventData))

	if err := L.PCall(2, 0, nil); err != nil {
		fmt.Printf("[ERROR] Failed to emit Lua event %s: %v\n", eventName, err)
	}
}

// loadUserScripts loads all .lua files from the user script directory
func (engine *LuaEngine) loadUserScripts() error {
	if engine.userScriptDir == "" {
		return nil
	}

	return filepath.Walk(engine.userScriptDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".lua" {
			return nil
		}
		if err := engine.L.DoFile(path); err != nil {
			return fmt.Errorf("error loading user Lua file %s: %w", path, err)
		}
		return nil
	})
}
