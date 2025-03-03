package lua

import (
	"testing"
	"time"

	"github.com/mmcdole/runes/pkg/client/types"
	"github.com/stretchr/testify/assert"
)

// TestOutputBufferAndEmission tests the output buffer and event emission
func TestOutputBufferAndEmission(t *testing.T) {
	// Create a custom event system that tracks output events
	eventSystem := NewTestEventSystem([]string{})

	// Create a new LuaEngine
	engine, err := NewLuaEngine(eventSystem, "")
	assert.NoError(t, err)
	defer engine.Close()

	// Clear any initial output lines from engine initialization
	_ = engine.getOutputLines()

	// Add an output line
	outputLine := &types.Line{
		Raw:     "Test output",
		Content: "Test output",
	}

	// Test addOutputLine and emitOutputLines
	engine.addOutputLine(outputLine)

	// Get the lines without emitting
	lines := engine.getOutputLines()
	assert.Equal(t, 1, len(lines))
	assert.Equal(t, "Test output", lines[0].Content)

	// Verify buffer is cleared
	lines = engine.getOutputLines()
	assert.Equal(t, 0, len(lines))
}

// TestLineModificationInProcessors tests that processors can modify lines
func TestLineModificationInProcessors(t *testing.T) {
	// Create a test event system
	eventSystem := NewTestEventSystem([]string{})

	// Create a new LuaEngine
	engine, err := NewLuaEngine(eventSystem, "")
	assert.NoError(t, err)
	defer engine.Close()

	// Add a processor that modifies the line
	modifyLua := `
	runes._add_output_processor(function(line)
		line:gag(true)
		line:skiplog(true)
		return line
	end)
	`

	// Execute the Lua code
	err = engine.state.DoString(modifyLua)
	assert.NoError(t, err)

	// Process a line
	line := &types.Line{
		Raw:     "Test output",
		Content: "Test output",
	}

	// Process the line
	result := engine.ProcessOutput(line)

	// Verify the modifications
	assert.True(t, result.Flags.Gag)
	assert.True(t, result.Flags.SkipLog)
}

// TestMultipleProcessorChaining tests that multiple processors can be chained
func TestMultipleProcessorChaining(t *testing.T) {
	// Create a test event system
	eventSystem := NewTestEventSystem([]string{"processor1", "processor2"})

	// Create a new LuaEngine
	engine, err := NewLuaEngine(eventSystem, "")
	assert.NoError(t, err)
	defer engine.Close()

	// Add multiple processors
	chainLua := `
	runes._add_input_processor(function(line)
		runes.send("processor1")
		return line
	end)
	
	runes._add_input_processor(function(line)
		runes.send("processor2")
		return line
	end)
	`

	// Execute the Lua code
	err = engine.state.DoString(chainLua)
	assert.NoError(t, err)

	// Process a line
	line := &types.Line{
		Raw:     "test",
		Content: "test",
	}

	engine.ProcessInput(line)

	// Verify both processors were called
	commands, ok := eventSystem.WaitForCommands(1 * time.Second)
	assert.True(t, ok)
	assert.Contains(t, commands, "processor1")
	assert.Contains(t, commands, "processor2")
}

// TestNilLineHandling tests that the engine properly handles nil lines
func TestNilLineHandling(t *testing.T) {
	// Create a test event system
	eventSystem := NewTestEventSystem([]string{})

	// Create a new LuaEngine
	engine, err := NewLuaEngine(eventSystem, "")
	assert.NoError(t, err)
	defer engine.Close()

	// These should not panic
	engine.ProcessInput(nil)
	result := engine.ProcessOutput(nil)
	assert.Nil(t, result)

	engine.directOutput(nil)
}
