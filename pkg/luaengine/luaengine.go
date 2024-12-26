package luaengine

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mmcdole/runes/pkg/api"
	lua "github.com/yuin/gopher-lua"
)

//go:embed core/*.lua
var coreLuaScripts embed.FS

type LuaEngine struct {
	L             *lua.LState
	coreScriptsFS fs.FS
	userScriptDir string
	callbacks     api.LuaCallbacks
	sentCommands  chan string
}

// New creates a new LuaEngine but does not initialize it
func New(coreScriptsFS fs.FS, userScriptDir string, callbacks api.LuaCallbacks) *LuaEngine {
	return &LuaEngine{
		L:             lua.NewState(),
		coreScriptsFS: coreScriptsFS,
		userScriptDir: userScriptDir,
		callbacks:     callbacks,
		sentCommands:  make(chan string, 100),
	}
}

// Initialize sets up the Lua environment and loads all scripts
func (engine *LuaEngine) Initialize() error {
	// 1. Set up Go bindings and global tables
	L := engine.L
	L.SetGlobal("runes", L.NewTable())

	// 2. Register Go functions
	runesNamespace := map[string]lua.LGFunction{
		"connect":       engine.connect,
		"disconnect":    engine.disconnect,
		"output":        engine.output,
		"log":           engine.log,
		"debug":         engine.debug,
		"version":       engine.version,
		"list_buffers":  engine.listBuffers,
		"switch_buffer": engine.switchBuffer,
		"sendRaw":       engine.sendCommand,
	}

	engine.registerGoFunctions(runesNamespace, "runes")

	// 3. Load core modules in specific order
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
		fmt.Printf("[DEBUG] Loading module %s from %s\n", module.name, module.path)
		content, err := fs.ReadFile(engine.coreScriptsFS, module.path)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", module.path, err)
		}
		if err := engine.L.DoString(string(content)); err != nil {
			return fmt.Errorf("error executing %s: %w\n\nContent:\n%s", module.path, err, string(content))
		}
		fmt.Printf("[DEBUG] Successfully loaded module: %s\n", module.name)
	}

	// 4. Load user scripts last
	if err := engine.loadUserScripts(); err != nil {
		return fmt.Errorf("error loading user scripts: %w", err)
	}

	return nil
}

// Close closes the Lua state.
func (engine *LuaEngine) Close() {
	engine.L.Close()
}

// registerGoFunctions is a helper function to register Go functions to a Lua namespace.
func (engine *LuaEngine) registerGoFunctions(functions map[string]lua.LGFunction, namespace string) {
	table := engine.L.GetGlobal(namespace)
	for name, fn := range functions {
		engine.L.SetField(table, name, engine.L.NewFunction(fn))
	}
}

// loadUserScripts loads user scripts from the engine's userScriptDir.
func (engine *LuaEngine) loadUserScripts() error {
	if engine.userScriptDir != "" {
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
	return nil
}

// GetSentCommands returns the channel for receiving sent commands.
func (engine *LuaEngine) GetSentCommands() <-chan string {
	return engine.sentCommands
}

func (engine *LuaEngine) sendCommand(L *lua.LState) int {
	command := L.ToString(1)
	fmt.Println("[SEND]", command)
	engine.sentCommands <- command
	return 0
}

// Lua function implementations
func (engine *LuaEngine) connect(L *lua.LState) int {
	host := L.ToString(1)
	port := L.ToInt(2)
	err := engine.callbacks.Connect(host, port)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func (engine *LuaEngine) disconnect(L *lua.LState) int {
	engine.callbacks.Disconnect()
	return 0
}

func (engine *LuaEngine) output(L *lua.LState) int {
	text := L.ToString(1)
	buffer := L.ToString(2)
	engine.callbacks.Output(text, buffer)
	return 0
}

func (engine *LuaEngine) log(L *lua.LState) int {
	// Implementation for log
	return 0
}

func (engine *LuaEngine) debug(L *lua.LState) int {
	// Implementation for debug
	text := L.ToString(1)
	fmt.Println("[LUA] [DEBUG]", text)
	return 0
}

func (engine *LuaEngine) version(L *lua.LState) int {
	// Implementation for version
	L.Push(lua.LString("1.0.0"))
	return 1
}

func (engine *LuaEngine) listBuffers(L *lua.LState) int {
	buffers := engine.callbacks.ListBuffers()
	// Convert to Lua table and push to stack
	t := L.CreateTable(len(buffers), 0)
	for i, buffer := range buffers {
		L.SetTable(t, lua.LNumber(i+1), lua.LString(buffer))
	}
	L.Push(t)
	return 1
}

func (engine *LuaEngine) switchBuffer(L *lua.LState) int {
	name := L.ToString(1)
	engine.callbacks.SwitchBuffer(name)
	return 0
}

// EmitEvent emits an event to the Lua environment
func (engine *LuaEngine) EmitEvent(eventName string, eventData string) {
	L := engine.L

	// Get the events table and emit function
	eventsTable := L.GetGlobal("events")
	emitFn := L.GetField(eventsTable, "emit")

	// Push function and arguments
	L.Push(emitFn)
	L.Push(lua.LString(eventName))
	L.Push(lua.LString(eventData))

	// Call events.emit(eventName, eventData)
	if err := L.PCall(2, 0, nil); err != nil {
		fmt.Printf("[ERROR] Failed to emit event %s: %v\n", eventName, err)
	}
}
