package scripting

import (
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Constants for timer registry tables
const (
	TimerCallbackTable     = "timer_callbacks"
	TimerTickCallbackTable = "timer_tick_callbacks"
	TimerNextID            = "timer_next_id"
)

// TimerEntry represents a registered timer
type TimerEntry struct {
	Duration time.Duration
	Count    int  // Number of times to run (0 = infinite)
	LastRun  time.Time
	CoreMode bool // Whether this timer is a core timer
}

// TimerManager handles scheduling and executing timers
type TimerManager struct {
	engine     *LuaEngine
	timers     map[uint32]TimerEntry
	mutex      sync.Mutex
	nextTimerID uint32
}

// NewTimerManager creates a new timer manager
func NewTimerManager(engine *LuaEngine) *TimerManager {
	return &TimerManager{
		engine:    engine,
		timers:    make(map[uint32]TimerEntry),
		nextTimerID: 1, // Start from 1 like Blightmud does
	}
}

// RegisterTimer registers the timer module with Lua
func RegisterTimer(L *lua.LState, tm *TimerManager) {
	// Create timer callback tables
	callbacksTable := L.NewTable()
	L.SetGlobal(TimerCallbackTable, callbacksTable)
	
	tickCallbacksTable := L.NewTable()
	L.SetGlobal(TimerTickCallbackTable, tickCallbacksTable)
	
	// Set next ID
	L.SetGlobal(TimerNextID, lua.LNumber(1))
	
	// Create timer module
	timerMod := L.NewTable()
	L.SetGlobal("timer", timerMod)
	
	// Register timer functions
	L.SetFuncs(timerMod, map[string]lua.LGFunction{
		"add":     tm.luaAddTimer,
		"remove":  tm.luaRemoveTimer,
		"clear":   tm.luaClearTimers,
		"get_ids": tm.luaGetTimerIDs,
		"on_tick": tm.luaOnTick,
		
		// Internal functions (not to be called directly by users)
		"_execute": tm.luaExecuteTimer,
		"_tick":    tm.luaTickTimers,
	})
}

// RegisterTimer registers a new timer
func (tm *TimerManager) RegisterTimer(duration time.Duration, count int, coreMode bool) uint32 {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	id := tm.nextTimerID
	tm.nextTimerID++

	tm.timers[id] = TimerEntry{
		Duration: duration,
		Count:    count,
		LastRun:  time.Now(),
		CoreMode: coreMode,
	}

	return id
}

// RemoveTimer removes a timer
func (tm *TimerManager) RemoveTimer(id uint32) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	delete(tm.timers, id)
}

// ClearTimers removes all timers
func (tm *TimerManager) ClearTimers(coreMode bool) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// If coreMode is true, only clear core timers
	// If coreMode is false, only clear user timers
	for id, timer := range tm.timers {
		if timer.CoreMode == coreMode {
			delete(tm.timers, id)
		}
	}
}

// Tick processes all due timers
func (tm *TimerManager) Tick(elapsed time.Duration) {
	now := time.Now()
	
	// Get a list of timers to process
	var timersToProcess []uint32
	
	tm.mutex.Lock()
	// First, collect IDs of timers that need processing
	for id, timer := range tm.timers {
		if now.Sub(timer.LastRun) >= timer.Duration {
			timersToProcess = append(timersToProcess, id)
		}
	}
	tm.mutex.Unlock()
	
	// Process tick callbacks - these receive the elapsed time in milliseconds
	L := tm.engine.state
	if L != nil {
		// Call all tick callbacks with the elapsed time
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("timer"),
			NRet:    0,
			Protect: true,
		}, lua.LString("_tick"), lua.LNumber(elapsed.Milliseconds())); err != nil {
			// Ignore errors in timer tick callbacks
		}
	}
	
	// Process individual timers
	for _, id := range timersToProcess {
		tm.processTimer(id)
	}
}

// processTimer executes a specific timer
func (tm *TimerManager) processTimer(id uint32) {
	tm.mutex.Lock()
	timer, exists := tm.timers[id]
	if !exists {
		tm.mutex.Unlock()
		return
	}
	
	// Update the last run time
	timer.LastRun = time.Now()
	
	// Decrement count if not infinite
	if timer.Count > 0 {
		timer.Count--
	}
	
	// Check if timer should be removed
	removeTimer := timer.Count == 0 && timer.Count != 0 // Remove if count reached 0 and wasn't infinite
	
	// Update the timer in the map
	if !removeTimer {
		tm.timers[id] = timer
	} else {
		delete(tm.timers, id)
	}
	tm.mutex.Unlock()
	
	// Execute the timer callback
	L := tm.engine.state
	if L != nil {
		// Call the timer callback
		if err := L.CallByParam(lua.P{
			Fn:      L.GetGlobal("timer"),
			NRet:    0,
			Protect: true,
		}, lua.LString("_execute"), lua.LNumber(id)); err != nil {
			// Ignore errors in timer execution
		}
	}
}

// Lua binding functions

// luaAddTimer adds a timer
// Lua syntax: timer.add(seconds, count, callback) -> id
func (tm *TimerManager) luaAddTimer(L *lua.LState) int {
	seconds := float64(L.CheckNumber(1))
	count := L.CheckInt(2)
	callback := L.CheckFunction(3)
	
	// Convert seconds to duration
	duration := time.Duration(seconds * float64(time.Second))
	
	// Store callback in registry
	callbacksTable := L.GetGlobal(TimerCallbackTable)
	if callbacksTable.Type() != lua.LTTable {
		callbacksTable = L.NewTable()
		L.SetGlobal(TimerCallbackTable, callbacksTable)
	}
	
	// Register the timer
	id := tm.RegisterTimer(duration, count, false)
	
	// Store the callback
	L.SetTable(callbacksTable, lua.LNumber(id), callback)
	
	// Return the timer ID
	L.Push(lua.LNumber(id))
	return 1
}

// luaRemoveTimer removes a timer
// Lua syntax: timer.remove(id)
func (tm *TimerManager) luaRemoveTimer(L *lua.LState) int {
	id := uint32(L.CheckNumber(1))
	
	// Remove from Go side
	tm.RemoveTimer(id)
	
	// Remove from Lua side
	callbacksTable := L.GetGlobal(TimerCallbackTable)
	if callbacksTable.Type() == lua.LTTable {
		L.SetTable(callbacksTable, lua.LNumber(id), lua.LNil)
	}
	
	return 0
}

// luaClearTimers clears all timers
// Lua syntax: timer.clear()
func (tm *TimerManager) luaClearTimers(L *lua.LState) int {
	// Clear from Go side (non-core timers only)
	tm.ClearTimers(false)
	
	// Clear from Lua side
	L.SetGlobal(TimerCallbackTable, L.NewTable())
	
	return 0
}

// luaGetTimerIDs gets all timer IDs
// Lua syntax: timer.get_ids() -> {id1, id2, ...}
func (tm *TimerManager) luaGetTimerIDs(L *lua.LState) int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	// Create result table
	result := L.NewTable()
	
	// Add all non-core timer IDs
	i := 1
	for id, timer := range tm.timers {
		if !timer.CoreMode {
			L.RawSetInt(result, i, lua.LNumber(id))
			i++
		}
	}
	
	L.Push(result)
	return 1
}

// luaOnTick registers a function to be called on every tick
// Lua syntax: timer.on_tick(callback)
func (tm *TimerManager) luaOnTick(L *lua.LState) int {
	callback := L.CheckFunction(1)
	
	// Store callback in registry
	tickCallbacksTable := L.GetGlobal(TimerTickCallbackTable)
	if tickCallbacksTable.Type() != lua.LTTable {
		tickCallbacksTable = L.NewTable()
		L.SetGlobal(TimerTickCallbackTable, tickCallbacksTable)
	}
	
	// Get next ID
	nextID := L.GetGlobal(TimerNextID)
	id := int(lua.LVAsNumber(nextID))
	
	// Store the callback
	L.SetTable(tickCallbacksTable, lua.LNumber(id), callback)
	
	// Increment next ID
	L.SetGlobal(TimerNextID, lua.LNumber(id+1))
	
	return 0
}

// luaExecuteTimer executes a timer callback (internal use)
// Lua syntax: timer._execute(id)
func (tm *TimerManager) luaExecuteTimer(L *lua.LState) int {
	id := uint32(L.CheckNumber(1))
	
	// Get the callback
	callbacksTable := L.GetGlobal(TimerCallbackTable)
	if callbacksTable.Type() != lua.LTTable {
		return 0
	}
	
	// Get and call the callback
	callback := L.GetTable(callbacksTable, lua.LNumber(id))
	if callback.Type() == lua.LTFunction {
		if err := L.CallByParam(lua.P{
			Fn:      callback,
			NRet:    0,
			Protect: true,
		}); err != nil {
			// Ignore errors in timer callback
		}
	}
	
	return 0
}

// luaTickTimers calls all tick callbacks (internal use)
// Lua syntax: timer._tick(milliseconds)
func (tm *TimerManager) luaTickTimers(L *lua.LState) int {
	millis := L.CheckNumber(1)
	
	// Get the tick callbacks table
	tickCallbacksTable := L.GetGlobal(TimerTickCallbackTable)
	if tickCallbacksTable.Type() != lua.LTTable {
		return 0
	}
	
	// Call all callbacks
	tickCallbacksTable.(*lua.LTable).ForEach(func(_, value lua.LValue) {
		if value.Type() == lua.LTFunction {
			if err := L.CallByParam(lua.P{
				Fn:      value,
				NRet:    0,
				Protect: true,
			}, millis); err != nil {
				// Ignore errors in tick callback
			}
		}
	})
	
	return 0
}
