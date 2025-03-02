package lua

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	lua "github.com/yuin/gopher-lua"
)

const (
	// Registry table names
	InputProcessorTable  = "__input_processors"
	OutputProcessorTable = "__output_processors"
)

// LuaEngine handles Lua script execution and line processing
type LuaEngine struct {
	state           *lua.LState
	scriptDir       string
	eventSystem     events.EventSystem
	outputBuffer    []*types.Line    // Buffer to collect output lines during script execution
	bufferMutex     sync.Mutex       // Mutex to protect the buffer during concurrent access
}

// NewLuaEngine creates a new Lua scripting engine
func NewLuaEngine(eventSystem events.EventSystem, scriptDir string) (*LuaEngine, error) {
	state := lua.NewState()
	engine := &LuaEngine{
		state:           state,
		scriptDir:       scriptDir,
		eventSystem:     eventSystem,
	}

	// Initialize registry tables for processors
	if err := engine.initRegistryTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize registry tables: %w", err)
	}

	// Register types and functions
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
	return e.processWithTable(line, InputProcessorTable, false)
}

// ProcessOutput processes an output line through all registered output processors
func (e *LuaEngine) ProcessOutput(line *types.Line) *types.Line {
	result := e.processWithTable(line, OutputProcessorTable, true)
	
	// After processing, emit any output lines that were collected during processing
	e.emitOutputLines()
	
	return result
}

// processWithTable processes a line using the specified processor table
func (e *LuaEngine) processWithTable(line *types.Line, tableKey string, isOutput bool) *types.Line {
	if line == nil {
		return nil
	}

	// Convert to LuaLine
	luaLine := &LuaLine{line: line}
	ud := e.state.NewUserData()
	ud.Value = luaLine
	e.state.SetMetatable(ud, e.state.GetTypeMetatable("line"))

	// Get the processor table from registry
	processorTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), tableKey).(*lua.LTable)
	
	// Process through each processor
	for i := 1; i <= processorTable.Len(); i++ {
		fn := processorTable.RawGetInt(i)
		if fn.Type() != lua.LTFunction {
			continue
		}
		
		// Call the processor function with the line
		e.state.Push(fn)
		e.state.Push(ud)
		
		if err := e.state.PCall(1, 1, nil); err != nil {
			// Log error but continue processing
			processorType := "input"
			if isOutput {
				processorType = "output"
			}
			fmt.Printf("Error in %s processor: %v\n", processorType, err)
			continue
		}
		
		// Update line if a valid result was returned
		if resultUD, ok := e.state.Get(-1).(*lua.LUserData); ok {
			if resultLine, ok := resultUD.Value.(*LuaLine); ok {
				luaLine = resultLine
				ud.Value = luaLine
			}
		}
		e.state.Pop(1)
	}
	
	return luaLine.line
}

// initRegistryTables initializes the Lua registry tables for processors
func (e *LuaEngine) initRegistryTables() error {
	// Create tables for input and output processors
	inputTable := e.state.NewTable()
	outputTable := e.state.NewTable()
	
	// Store in registry
	e.state.SetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), InputProcessorTable, inputTable)
	e.state.SetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), OutputProcessorTable, outputTable)
	
	return nil
}

// addInputProcessor adds a function to process input lines
func (e *LuaEngine) addInputProcessor(fn *lua.LFunction) {
	// Get the input processor table from registry
	inputTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), InputProcessorTable).(*lua.LTable)
	
	// Add the function to the table with the next available index
	inputTable.RawSetInt(inputTable.Len()+1, fn)
}

// addOutputProcessor adds a function to process output lines
func (e *LuaEngine) addOutputProcessor(fn *lua.LFunction) {
	// Get the output processor table from registry
	outputTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), OutputProcessorTable).(*lua.LTable)
	
	// Add the function to the table with the next available index
	outputTable.RawSetInt(outputTable.Len()+1, fn)
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

// addOutputLine adds a line to the output buffer
func (e *LuaEngine) addOutputLine(line *types.Line) {
	e.bufferMutex.Lock()
	defer e.bufferMutex.Unlock()
	e.outputBuffer = append(e.outputBuffer, line)
}

// getOutputLines retrieves and clears the output buffer
func (e *LuaEngine) getOutputLines() []*types.Line {
	e.bufferMutex.Lock()
	defer e.bufferMutex.Unlock()
	
	// Create a copy of the buffer
	lines := make([]*types.Line, len(e.outputBuffer))
	copy(lines, e.outputBuffer)
	
	// Clear the buffer
	e.outputBuffer = e.outputBuffer[:0]
	
	return lines
}

// emitOutputLines emits all collected output lines as events
func (e *LuaEngine) emitOutputLines() {
	lines := e.getOutputLines()
	for _, line := range lines {
		e.eventSystem.Emit(events.Event{
			Type: events.EventOutput, // We'll need to add this event type
			Data: line,
		})
	}
}

// directOutput emits a line directly without processing
func (e *LuaEngine) directOutput(line *types.Line) {
	if line == nil {
		return
	}
	
	e.eventSystem.Emit(events.Event{
		Type: events.EventOutput,
		Data: line,
	})
}

// Close closes the Lua state
func (e *LuaEngine) Close() {
	if e.state != nil {
		e.state.Close()
		e.state = nil
	}
}
