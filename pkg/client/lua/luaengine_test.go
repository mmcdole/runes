package lua

// // TestOutputBufferAndEmission tests the output buffer and event emission
// func TestOutputBufferAndEmission(t *testing.T) {
// 	// Create a custom event system that tracks output events
// 	eventSystem := NewTestEventSystem([]string{})

// 	// Create a new LuaEngine
// 	engine, err := NewLuaEngine(eventSystem, "")
// 	assert.NoError(t, err)
// 	defer engine.Close()

// 	// Clear any initial output lines from engine initialization
// 	_ = engine.flushOutputLines()

// 	// Add an output line
// 	outputLine := types.NewLine("Test output")

// 	// Test addOutputLine and emitOutputLines
// 	engine.addOutputLine(outputLine)

// 	// Get the lines without emitting
// 	lines := engine.flushOutputLines()
// 	assert.Equal(t, 1, len(lines))
// 	assert.Equal(t, "Test output", lines[0].Display)

// 	// Verify buffer is cleared
// 	lines = engine.flushOutputLines()
// 	assert.Equal(t, 0, len(lines))
// }

// // TestLineModificationInProcessors tests that processors can modify lines
// func TestLineModificationInProcessors(t *testing.T) {
// 	// Create a test event system
// 	eventSystem := NewTestEventSystem([]string{})

// 	// Create a new LuaEngine
// 	engine, err := NewLuaEngine(eventSystem, "")
// 	assert.NoError(t, err)
// 	defer engine.Close()

// 	// Add a processor that modifies the line
// 	modifyLua := `
// 	runes._add_output_processor(function(line)
// 		line:gag(true)
// 		line:skiplog(true)
// 		return line
// 	end)
// 	`

// 	// Execute the Lua code
// 	err = engine.state.DoString(modifyLua)
// 	assert.NoError(t, err)

// 	// Process a line
// 	line := types.NewLine("Test output")

// 	// Process the line
// 	result := engine.ProcessOutput(line)

// 	// Verify the modifications
// 	assert.True(t, result.Flags.Gag)
// 	assert.True(t, result.Flags.SkipLog)
// }

// // TestMultipleProcessorChaining tests that multiple processors can be chained
// func TestMultipleProcessorChaining(t *testing.T) {
// 	fmt.Println("Test starting")
// 	// Create a simple mock event system that just records commands
// 	mockEvents := &struct {
// 		commands []string
// 		mu       sync.Mutex
// 	}{}

// 	eventSystem := &testEventSystem{
// 		emitFunc: func(event events.Event) {
// 			fmt.Println("Event emitted:", event.Type)
// 			if event.Type == events.EventCommand {
// 				if cmd, ok := event.Data.(string); ok {
// 					fmt.Println("Command received:", cmd)
// 					mockEvents.mu.Lock()
// 					mockEvents.commands = append(mockEvents.commands, cmd)
// 					mockEvents.mu.Unlock()
// 				}
// 			}
// 		},
// 	}

// 	fmt.Println("Creating LuaEngine")
// 	// Create a new LuaEngine
// 	engine, err := NewLuaEngine(eventSystem, "")
// 	assert.NoError(t, err)
// 	defer engine.Close()

// 	fmt.Println("Adding processors")
// 	// Add multiple processors
// 	chainLua := `
// 	runes._add_output_processor(function(line)
// 		print("Processor 1 called")
// 		line:set_display(line:display() .. " - modified by processor1")
// 		runes.send("processor1")
// 		return line
// 	end)

// 	runes._add_output_processor(function(line)
// 		print("Processor 2 called")
// 		line:set_display(line:display() .. " - modified by processor2")
// 		runes.send("processor2")
// 		return line
// 	end)
// 	`

// 	fmt.Println("Executing Lua code")
// 	// Execute the Lua code
// 	err = engine.state.DoString(chainLua)
// 	assert.NoError(t, err)

// 	fmt.Println("Creating line")
// 	// Process a line
// 	line := types.NewLine("test")

// 	fmt.Println("Processing output")
// 	// Process the line through output processors
// 	result := engine.ProcessOutput(line)
// 	fmt.Println("Output processed")

// 	// Wait a short time for commands to be processed
// 	fmt.Println("Waiting for commands")
// 	time.Sleep(100 * time.Millisecond)

// 	fmt.Println("Checking commands")
// 	// Verify the commands were sent
// 	mockEvents.mu.Lock()
// 	commands := mockEvents.commands
// 	mockEvents.mu.Unlock()

// 	fmt.Println("Commands:", commands)
// 	assert.Contains(t, commands, "processor1")
// 	assert.Contains(t, commands, "processor2")

// 	// Verify the line was modified by both processors
// 	fmt.Println("Checking line modifications")
// 	assert.Equal(t, "test - modified by processor1 - modified by processor2", result.Display)
// 	fmt.Println("Test completed")
// }

// // Simple test event system implementation
// type testEventSystem struct {
// 	emitFunc func(event events.Event)
// }

// func (t *testEventSystem) Subscribe(eventType events.EventType, handler events.Handler) {
// 	// No-op for testing
// }

// func (t *testEventSystem) Emit(event events.Event) {
// 	if t.emitFunc != nil {
// 		t.emitFunc(event)
// 	}
// }
