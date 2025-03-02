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

// luaLineNew creates a new Line object
// Lua syntax: line.new(text) -> line
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

// luaLineRaw gets the raw text of a line
// Lua syntax: line:raw() -> string
func luaLineRaw(L *lua.LState) int {
    line := checkLine(L)
    L.Push(lua.LString(line.line.Raw))
    return 1
}

// luaLineDisplay gets the display text of a line
// Lua syntax: line:display() -> string
func luaLineDisplay(L *lua.LState) int {
    line := checkLine(L)
    L.Push(lua.LString(line.line.Display))
    return 1
}

// luaLineGag gets or sets the gag flag of a line
// Lua syntax: line:gag() -> boolean
// Lua syntax: line:gag(boolean) -> boolean
func luaLineGag(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Flags.Gag = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Flags.Gag))
    return 1
}

// luaLinePrompt gets or sets the prompt flag of a line
// Lua syntax: line:prompt() -> boolean
// Lua syntax: line:prompt(boolean) -> boolean
func luaLinePrompt(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Flags.IsPrompt = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Flags.IsPrompt))
    return 1
}

// luaLineComplete gets or sets the complete flag of a line
// Lua syntax: line:complete() -> boolean
// Lua syntax: line:complete(boolean) -> boolean
func luaLineComplete(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Flags.Complete = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Flags.Complete))
    return 1
}

// luaLineMatched gets or sets the matched flag of a line
// Lua syntax: line:matched() -> boolean
// Lua syntax: line:matched(boolean) -> boolean
func luaLineMatched(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Flags.Matched = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Flags.Matched))
    return 1
}

// luaLineSkipLog gets or sets the skiplog flag of a line
// Lua syntax: line:skiplog() -> boolean
// Lua syntax: line:skiplog(boolean) -> boolean
func luaLineSkipLog(L *lua.LState) int {
    line := checkLine(L)
    if L.GetTop() > 1 {
        line.line.Flags.SkipLog = L.ToBool(2)
    }
    L.Push(lua.LBool(line.line.Flags.SkipLog))
    return 1
}
