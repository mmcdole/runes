package lua

import (
    "github.com/mmcdole/runes/pkg/client/types"
    lua "github.com/yuin/gopher-lua"
)

// LuaLine wraps a Line for Lua
type LuaLine struct {
    line *types.Line
}

// Register adds the Line type to Lua
func RegisterLine(L *lua.LState) {
    mt := L.NewTypeMetatable("line")
    L.SetGlobal("line", mt)
    
    // Static functions
    L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
        "new": luaLineNew,
    }))

    // Instance methods
    instanceMt := L.NewTable()
    L.SetFuncs(instanceMt, map[string]lua.LGFunction{
        "raw":      luaLineRaw,      // Get raw text
        "line":     luaLineRaw,      // Alias for raw (Blightmud compat)
        "display":  luaLineDisplay,   // Get display text
        "gag":      luaLineGag,       // Get/set gag flag
        "prompt":   luaLinePrompt,    // Get/set prompt flag
        "complete": luaLineComplete,  // Get/set complete flag
        "matched":  luaLineMatched,   // Get/set matched flag
        "skiplog":  luaLineSkipLog,   // Get/set skiplog flag
    })
    L.SetField(mt, "__index", instanceMt)
}

// Constructor
func luaLineNew(L *lua.LState) int {
    text := L.CheckString(1)
    line := &LuaLine{line: types.NewLine(text)}
    ud := L.NewUserData()
    ud.Value = line
    L.SetMetatable(ud, L.GetTypeMetatable("line"))
    L.Push(ud)
    return 1
}

func checkLine(L *lua.LState) *LuaLine {
    if ud := L.CheckUserData(1); ud != nil {
        if v, ok := ud.Value.(*LuaLine); ok {
            return v
        }
    }
    L.ArgError(1, "line expected")
    return nil
}

// Raw text getter
func luaLineRaw(L *lua.LState) int {
    line := checkLine(L)
    L.Push(lua.LString(line.line.Raw))
    return 1
}

// Display text getter
func luaLineDisplay(L *lua.LState) int {
    line := checkLine(L)
    L.Push(lua.LString(line.line.Display))
    return 1
}

// Gag flag getter/setter
func luaLineGag(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Gag = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Gag))
    return 1
}

// Prompt flag getter/setter
func luaLinePrompt(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.IsPrompt = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.IsPrompt))
    return 1
}

// Complete flag getter/setter
func luaLineComplete(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Complete = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Complete))
    return 1
}

// Matched flag getter/setter
func luaLineMatched(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Matched = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Matched))
    return 1
}

// SkipLog flag getter/setter
func luaLineSkipLog(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.SkipLog = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.SkipLog))
    return 1
}
