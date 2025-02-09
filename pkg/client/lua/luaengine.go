package lua

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine handles Lua script execution and line processing
type LuaEngine struct {
	state           *lua.LState
	scriptDir       string
	eventSystem     events.EventSystem
	inputProcessors  []*lua.LFunction
	outputProcessors []*lua.LFunction
}

// NewLuaEngine creates a new Lua scripting engine
func NewLuaEngine(eventSystem events.EventSystem, scriptDir string) (*LuaEngine, error) {
	state := lua.NewState()
	engine := &LuaEngine{
		state:           state,
		scriptDir:       scriptDir,
		eventSystem:     eventSystem,
		inputProcessors: make([]*lua.LFunction, 0),
		outputProcessors: make([]*lua.LFunction, 0),
	}

	// Register types and functions
	RegisterLine(engine.state)
	if err := registerBindings(engine.state, engine); err != nil {
		return nil, fmt.Errorf("failed to register bindings: %w", err)
	}

	// Load core scripts
	if err := engine.loadCoreScripts(); err != nil {
		return nil, fmt.Errorf("failed to load core scripts: %w", err)
	}

	return engine, nil
}

// ProcessInput processes an input line through all registered input processors
func (e *LuaEngine) ProcessInput(line *types.Line) *types.Line {
	if line == nil {
		return nil
	}

	// Convert to LuaLine
	luaLine := &LuaLine{line: line}
	ud := e.state.NewUserData()
	ud.Value = luaLine
	e.state.SetMetatable(ud, e.state.GetTypeMetatable("line"))

	// Process through each processor
	for _, fn := range e.inputProcessors {
		e.state.Push(fn)
		e.state.Push(ud)

		if err := e.state.PCall(1, 1, nil); err != nil {
			// Log error but continue processing
			e.eventSystem.Emit(events.Event{
				Type: events.EventDebug,
				Data: fmt.Sprintf("Error in input processor: %v", err),
			})
			continue
		}

		// Get result
		if e.state.Get(-1) != lua.LNil {
			if resultUD, ok := e.state.Get(-1).(*lua.LUserData); ok {
				if resultLine, ok := resultUD.Value.(*LuaLine); ok {
					luaLine = resultLine
				}
			}
		}
		e.state.Pop(1)

		// Check if line was gagged
		if luaLine.line.Flags.Gag {
			return nil
		}
	}

	return luaLine.line
}

// ProcessOutput processes an output line through all registered output processors
func (e *LuaEngine) ProcessOutput(line *types.Line) *types.Line {
	if line == nil {
		return nil
	}

	// Convert to LuaLine
	luaLine := &LuaLine{line: line}
	ud := e.state.NewUserData()
	ud.Value = luaLine
	e.state.SetMetatable(ud, e.state.GetTypeMetatable("line"))

	// Process through each processor
	for _, fn := range e.outputProcessors {
		e.state.Push(fn)
		e.state.Push(ud)

		if err := e.state.PCall(1, 1, nil); err != nil {
			// Log error but continue processing
			e.eventSystem.Emit(events.Event{
				Type: events.EventDebug,
				Data: fmt.Sprintf("Error in output processor: %v", err),
			})
			continue
		}

		// Get result
		if e.state.Get(-1) != lua.LNil {
			if resultUD, ok := e.state.Get(-1).(*lua.LUserData); ok {
				if resultLine, ok := resultUD.Value.(*LuaLine); ok {
					luaLine = resultLine
				}
			}
		}
		e.state.Pop(1)

		// Check if line was gagged
		if luaLine.line.Flags.Gag {
			return nil
		}
	}

	return luaLine.line
}

// AddInputProcessor adds a function to process input lines
func (e *LuaEngine) AddInputProcessor(fn *lua.LFunction) {
	e.inputProcessors = append(e.inputProcessors, fn)
}

// AddOutputProcessor adds a function to process output lines
func (e *LuaEngine) AddOutputProcessor(fn *lua.LFunction) {
	e.outputProcessors = append(e.outputProcessors, fn)
}

// loadCoreScripts loads the core Lua scripts
func (e *LuaEngine) loadCoreScripts() error {
	coreDir := filepath.Join(e.scriptDir, "core")
	entries, err := os.ReadDir(coreDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Core directory doesn't exist, that's ok
		}
		return fmt.Errorf("error reading core directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lua" {
			continue
		}

		path := filepath.Join(coreDir, entry.Name())
		if err := e.loadUserScript(path); err != nil {
			return fmt.Errorf("error loading core script %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// loadUserScript loads a user script
func (e *LuaEngine) loadUserScript(path string) error {
	if err := e.state.DoFile(path); err != nil {
		return fmt.Errorf("error loading script %s: %w", path, err)
	}
	return nil
}

// Close closes the Lua state
func (e *LuaEngine) Close() {
	if e.state != nil {
		e.state.Close()
		e.state = nil
	}
}
