package lua

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine handles Lua script execution and event processing
type LuaEngine struct {
	state             *lua.LState
	bindings         *luaBindings
	eventSystem      events.EventSystem
	scriptDir        string
}

// NewLuaEngine creates a new Lua scripting engine
func NewLuaEngine(eventSystem events.EventSystem, scriptDir string) *LuaEngine {
	state := lua.NewState()
	engine := &LuaEngine{
		state:        state,
		eventSystem:  eventSystem,
		scriptDir:   scriptDir,
	}
	engine.bindings = newLuaBindings(engine)
	engine.initRegistryTables()
	engine.bindings.register(state)
	return engine
}

// initRegistryTables initializes all registry tables
func (e *LuaEngine) initRegistryTables() {
	tables := []string{
		OutputListenerTable,
		InputListenerTable,
		OnConnectTable,
		OnDisconnectTable,
		ScriptResetTable,
	}

	for _, name := range tables {
		table := e.state.NewTable()
		e.state.SetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), name, table)
	}
}

// Reset resets the Lua state while preserving registered handlers
func (e *LuaEngine) Reset() error {
	// Call reset handlers
	if err := e.callResetHandlers(); err != nil {
		return fmt.Errorf("error calling reset handlers: %w", err)
	}

	// Create new state
	newState := lua.NewState()
	oldState := e.state
	e.state = newState

	// Re-initialize registry tables
	e.initRegistryTables()

	// Copy handlers from old state
	if err := e.copyRegistryTables(oldState); err != nil {
		return fmt.Errorf("error copying registry tables: %w", err)
	}

	// Close old state
	oldState.Close()

	// Re-register bindings
	e.bindings.register(newState)

	return nil
}

// copyRegistryTables copies all registry tables from old state to new state
func (e *LuaEngine) copyRegistryTables(oldState *lua.LState) error {
	tables := []string{
		OutputListenerTable,
		InputListenerTable,
		OnConnectTable,
		OnDisconnectTable,
		ScriptResetTable,
	}

	for _, name := range tables {
		oldTable := oldState.GetField(oldState.Get(lua.RegistryIndex).(*lua.LTable), name).(*lua.LTable)
		newTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), name).(*lua.LTable)

		oldTable.ForEach(func(k, v lua.LValue) {
			e.state.SetTable(newTable, k, v)
		})
	}

	return nil
}

// callResetHandlers calls all registered reset handlers
func (e *LuaEngine) callResetHandlers() error {
	resetTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), ScriptResetTable).(*lua.LTable)
	
	var err error
	resetTable.ForEach(func(k, v lua.LValue) {
		if fn, ok := v.(*lua.LFunction); ok {
			if err = e.state.CallByParam(lua.P{
				Fn:      fn,
				NRet:    0,
				Protect: true,
			}); err != nil {
				return
			}
		}
	})

	return err
}

// ProcessOutput processes an output line through all registered output listeners
func (e *LuaEngine) ProcessOutput(line *types.Line) error {
	if line.BypassScript {
		return nil
	}

	luaLine := newLuaLine(e.state, line)
	outputTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), OutputListenerTable).(*lua.LTable)

	var err error
	outputTable.ForEach(func(k, v lua.LValue) {
		if fn, ok := v.(*lua.LFunction); ok {
			if err = e.state.CallByParam(lua.P{
				Fn:      fn,
				NRet:    1,
				Protect: true,
			}, luaLine); err != nil {
				return
			}

			// Get returned line
			ret := e.state.Get(-1)
			if retLine, ok := ret.(*lua.LUserData); ok {
				if ll, ok := retLine.Value.(*LuaLine); ok {
					luaLine = ll
				}
			}
			e.state.Pop(1)
		}
	})

	if err != nil {
		return err
	}

	// Update original line
	line.Replace(luaLine.inner)
	if luaLine.replacement != nil {
		line.Content = *luaLine.replacement
	}

	return nil
}

// ProcessInput processes an input line through all registered input listeners
func (e *LuaEngine) ProcessInput(line *types.Line) error {
	if line.BypassScript {
		return nil
	}

	luaLine := newLuaLine(e.state, line)
	inputTable := e.state.GetField(e.state.Get(lua.RegistryIndex).(*lua.LTable), InputListenerTable).(*lua.LTable)

	var err error
	inputTable.ForEach(func(k, v lua.LValue) {
		if fn, ok := v.(*lua.LFunction); ok {
			if err = e.state.CallByParam(lua.P{
				Fn:      fn,
				NRet:    1,
				Protect: true,
			}, luaLine); err != nil {
				return
			}

			// Get returned line
			ret := e.state.Get(-1)
			if retLine, ok := ret.(*lua.LUserData); ok {
				if ll, ok := retLine.Value.(*LuaLine); ok {
					luaLine = ll
				}
			}
			e.state.Pop(1)
		}
	})

	if err != nil {
		return err
	}

	// Update original line
	line.Replace(luaLine.inner)
	if luaLine.replacement != nil {
		line.Content = *luaLine.replacement
	}

	return nil
}

// LoadScript loads and executes a Lua script
func (e *LuaEngine) LoadScript(path string) error {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(e.scriptDir, path)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", absPath)
	}

	return e.state.DoFile(absPath)
}

// Close closes the Lua state
func (e *LuaEngine) Close() {
	if e.state != nil {
		e.state.Close()
		e.state = nil
	}
}
