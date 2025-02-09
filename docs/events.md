# Event Handling Systems Comparison

## Current Async Event System (Runes)

The current event system in Runes implements a simple pub/sub (publisher/subscriber) pattern with the following characteristics:

- **Architecture**: Fully asynchronous event emission
- **Implementation**: Uses a table of event handlers where multiple handlers can subscribe to a single event
- **Flow**: 
  1. Events are emitted (`events.emit`)
  2. All registered handlers are called asynchronously
  3. Each handler runs independently without affecting other handlers
  4. No return values are processed from handlers
- **Error Handling**: Uses pcall to safely handle errors in individual handlers without affecting others
- **Advantages**:
  - Simple and clean implementation
  - Good for general purpose event handling
  - Handlers can't block each other
  - Excellent for UI updates, logging, and non-critical data processing
- **Disadvantages**:
  - No built-in way to modify data in transit
  - Can't easily implement features like "gags" that need to prevent output
  - No guaranteed order of execution

## Blightmud's Synchronous Chain Approach

Blightmud uses a synchronous chain of responsibility pattern for input/output handling, implemented in Rust with Lua bindings:

- **Architecture**: Synchronous chain of listeners stored in Lua registry tables
- **Implementation**: 
  - Uses two separate registry tables: `__output_listeners` and `__input_listeners`
  - Each listener is a Lua function that receives and returns a Line object
  - Lines can be modified or gagged (blocked) by any listener
  - Processing is fully synchronous - each listener runs in sequence
- **Flow**:
  1. Data enters the system (input/output)
  2. Line is converted to a Lua Line object
  3. Each registered listener processes the line in sequence:
     - Can modify line content via `replacement`
     - Can modify line flags (e.g., gag)
     - Must return the modified line object
  4. Final line state determines what happens:
     - If gagged: line is dropped
     - If modified: new content is used
     - Otherwise: original line continues
- **Error Handling**:
  - Uses pcall for safe execution
  - For input listeners: error = line is considered matched
  - For output listeners: error is logged but processing continues
- **Advantages**:
  - Predictable processing order
  - Simple modification chain
  - Error handling built-in
  - Clean separation between input and output processing
  - Efficient - no async overhead for core operations
- **Disadvantages**:
  - More complex implementation than pure event system
  - Blocking - slow handlers affect entire chain
  - Less flexible than async for complex operations
  - Must be careful with error handling

## Recommendation: Hybrid Approach

I recommend implementing a hybrid system that combines both patterns, taking specific implementation cues from Blightmud:

1. **Keep the Current Async System**:
   - Perfect for general-purpose events (UI updates, timers, etc.)
   - Maintain current simple API for most use cases
   - Use for any non-critical event processing

2. **Add Synchronous Line Processors**:
   ```lua
   -- New line processor system
   local input_processors = {}
   local output_processors = {}
   
   -- Line object to pass through chain
   local Line = {
       content = "",
       flags = {
           gag = false,
           matched = false,
           bypass_script = false,
           source = nil
       },
       replacement = nil
   }
   
   function Line:new(content)
       local line = setmetatable({}, { __index = Line })
       line.content = content
       line.flags = { gag = false, matched = false, bypass_script = false }
       return line
   end
   
   -- Core processing functions
   function process_output(line_obj)
       if line_obj.flags.bypass_script then return line_obj end
       
       for _, processor in ipairs(output_processors) do
           local status, result = pcall(processor, line_obj)
           if not status then
               -- Log error but continue chain
               log_error("Output processor error: " .. result)
               goto continue
           end
           if result == nil then
               -- Line was gagged
               line_obj.flags.gag = true
               break
           end
           line_obj = result
           ::continue::
       end
       return line_obj
   end
   
   function process_input(line_obj)
       if line_obj.flags.bypass_script then return line_obj end
       
       for _, processor in ipairs(input_processors) do
           local status, result = pcall(processor, line_obj)
           if not status then
               -- On error, consider line matched
               line_obj.flags.matched = true
               break
           end
           if result == nil then
               -- Line was gagged
               line_obj.flags.gag = true
               break
           end
           line_obj = result
       end
       return line_obj
   end
   
   -- Registration functions
   function add_output_processor(fn)
       table.insert(output_processors, fn)
   end
   
   function add_input_processor(fn)
       table.insert(input_processors, fn)
   end
   ```

3. **Usage Example**:
   ```lua
   -- Example output processor that gags health lines
   add_output_processor(function(line)
       if line.content:match("^Health:") then
           return nil  -- Gag the line
       end
       return line
   end)
   
   -- Example input processor that adds shortcuts
   add_input_processor(function(line)
       if line.content == "gg" then
           line.replacement = "say good game!"
       end
       return line
   end)
   ```

4. **Integration with Existing System**:
   - Use the sync chain for core input/output
   - Use async events for everything else
   - Bridge between systems when needed:
   ```lua
   -- Example of bridging
   add_output_processor(function(line)
       -- Process sync first
       if line.content:match("^Combat:") then
           line.replacement = "[COMBAT] " .. line.content
       end
       
       -- Then fire async event for UI updates etc
       events.emit("combat_line", line)
       
       return line
   end)
   ```

5. **Migration Path**:
   1. Implement the new Line processor system
   2. Add integration points in core input/output handling
   3. Move existing triggers/gags to new system
   4. Keep async events for non-line processing

This hybrid approach gives you:
- Efficient, predictable line processing
- Simple modification chain
- Clear error handling
- Compatibility with existing code
- Best of both worlds for different use cases

## Migration Path

1. Keep existing event system unchanged
2. Add new synchronous processors for input/output
3. Gradually move input/output handling to the new system
4. Document both patterns clearly for plugin developers

This approach allows you to maintain the simplicity of your current system while adding the power of Blightmud's chain processing where it matters most.
