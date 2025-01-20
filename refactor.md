# Line Type Refactor

## Overview
Consolidating multiple Line types into a single `client.Line` type with consistent raw/display fields and standardizing its usage across the codebase.

## Core Changes Made
1. Created unified `client.Line` type with:
   - `Raw` - Original bytes from network
   - `Display` - ANSI processed for terminal display
   - Various flags (IsPrompt, Gag, etc.)

2. Updated Lua bindings to expose:
   - `line:raw()` - Get original network bytes
   - `line:display()` - Get display-ready ANSI text
   - Flag methods (gag, prompt, etc.)

## Required Changes

### 1. Buffer Package
- [ ] Update Buffer to use client.Line for storage
- [ ] Use ANSI processor to convert Raw to Display:
  ```go
  func (b *Buffer) Write(line *client.Line) {
      // Process ANSI on write
      line.Display = b.processor.Process(line.Raw)
      // Store line...
  }
  ```

### 2. Viewport Package
- [ ] Update Viewport to render line.Display field, when it fetches lines from buffer


### 3. Event System
- [ ] Update event types to use client.Line:
  ```go
  const (
      EventRawInput  = "raw_input"   // string
      EventRawOutput = "raw_output"  // *client.Line
      EventPrompt    = "prompt"      // *client.Line
      // ...
  )
  ```
- [ ] Update event handlers to properly create/handle client.Line objects

### 4. Data Flow
The complete flow should be:
1. Network receives bytes
2. Connection creates client.Line with Raw set
3. Events emit client.Line to LuaEngine
4. LuaEngine wraps as LuaLine for Lua scripts
5. Lua scripts process line (can modify flags, etc)
6. Lua bindings convert back to client.Line
7. client.Line flows to Buffer
8. Buffer processes Raw → Display using ANSI processor
9. Viewport fetches and renders line.Display from Buffer

Key points:
- Lua gets first chance to process raw lines
- Buffer handles ANSI processing after Lua
- Viewport only renders final Display content

## Migration Steps
1. Update Buffer package first (central point)
2. Update Event system to use new types
3. Update Viewport to use Display field
4. Update all tests
6. Update any lua files or bindings