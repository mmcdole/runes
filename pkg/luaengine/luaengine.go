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

// CoreScripts returns the embedded filesystem containing core Lua scripts
func CoreScripts() fs.FS {
	return coreLuaScripts
}

type LuaEngine struct {
	L             *lua.LState
	userScriptDir string
	eventSystem   *events.EventProcessor
	bindings      *luaBindings
	cachedEmitFn  lua.LValue
}

func New(userScriptDir string, eventSystem *events.EventProcessor) *LuaEngine {
	engine := &LuaEngine{
		L:             lua.NewState(),
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
	// First initialize the Lua state with all bindings
	if err := engine.setupLuaBindings(); err != nil {
		return err
	}

	// Load core modules
	if err := engine.loadCoreModules(); err != nil {
		return err
	}

	// Cache the emit function after all core modules are loaded
	if err := engine.initializeEventSystem(); err != nil {
		return err
	}

	return engine.loadUserLuaScripts()
}

func (engine *LuaEngine) setupLuaBindings() error {
	L := engine.L
	runesTable := L.NewTable()
	L.SetGlobal("runes", runesTable)

	// Register all bindings from the bindings map
	for name, fn := range engine.bindings.getBindingsMap() {
		L.SetField(runesTable, name, L.NewFunction(fn))
	}

	return nil
}

func (engine *LuaEngine) loadCoreModules() error {
	coreModules := []struct {
		name string
		path string
	}{
		{"defaults", "core/defaults.lua"}, // Most fundamental, others depend on it
		{"events", "core/events.lua"},     // Most fundamental, others depend on it
		{"alias", "core/alias.lua"},       // Input and commands depend on this
		{"input", "core/input.lua"},       // Core input handling
		{"trigger", "core/trigger.lua"},   // Output processing
		{"timer", "core/timer.lua"},       // Timer system
		{"commands", "core/commands.lua"}, // Default commands, depends on alias
		{"init", "core/init.lua"},         // Final initialization
	}

	for _, module := range coreModules {
		content, err := fs.ReadFile(coreLuaScripts, module.path)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", module.path, err)
		}
		if err := engine.L.DoString(string(content)); err != nil {
			return fmt.Errorf("error executing %s: %w", module.path, err)
		}
	}

	return nil
}

// Close cleans up the Lua state
func (engine *LuaEngine) Close() {
	engine.L.Close()
}

// loadUserLuaScripts loads all .lua files from the user script directory
func (engine *LuaEngine) loadUserLuaScripts() error {
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

func (engine *LuaEngine) initializeEventSystem() error {
	eventsTable := engine.L.GetGlobal("events")
	if eventsTable.Type() == lua.LTNil {
		return fmt.Errorf("events table is nil after initialization")
	}
	engine.cachedEmitFn = engine.L.GetField(eventsTable, "emit")
	if engine.cachedEmitFn.Type() == lua.LTNil {
		return fmt.Errorf("events.emit function is nil after initialization")
	}
	return nil
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

	L.Push(engine.cachedEmitFn)
	L.Push(lua.LString(eventName))
	L.Push(lua.LString(eventData))

	if err := L.PCall(2, 0, nil); err != nil {
		fmt.Printf("[ERROR] Failed to emit Lua event %s: %v\n", eventName, err)
	}
}
