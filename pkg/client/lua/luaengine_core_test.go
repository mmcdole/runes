package lua

import (
	"testing"
	"time"

	"github.com/mmcdole/runes/pkg/client/types"
	"github.com/stretchr/testify/assert"
)

// TestInputProcessing tests basic input processing functionality
func TestInputProcessing(t *testing.T) {
	// Load the test suite
	suite, err := loadTestSuite("testdata/input_tests.json")
	if err != nil {
		t.Fatalf("Failed to load input test suite: %v", err)
	}

	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			// Setup a test event system with expected commands
			eventSystem := NewTestEventSystem(test.ExpectedCommands)

			// Create a new LuaEngine
			engine, err := NewLuaEngine(eventSystem, "")
			if err != nil {
				t.Fatalf("Failed to create LuaEngine: %v", err)
			}
			defer engine.Close()

			// Execute setup Lua code if present
			if test.SetupLua != nil {
				if err := executeSetupLua(engine, test.SetupLua); err != nil {
					t.Fatalf("Failed to execute setup Lua: %v", err)
				}
			}

			// Process input
			line := &types.Line{
				Raw:     test.Input,
				Content: test.Input,
			}

			engine.ProcessInput(line)

			// Wait for expected commands with timeout
			commands, ok := eventSystem.WaitForCommands(1 * time.Second)
			if !ok {
				t.Fatalf("Timed out waiting for commands. Got: %v, Expected: %v", commands, test.ExpectedCommands)
			}

			// Verify commands match expected commands
			assert.Equal(t, test.ExpectedCommands, commands)
		})
	}
}

// TestAliases tests alias functionality
func TestAliases(t *testing.T) {
	// Load the test suite
	suite, err := loadTestSuite("testdata/alias_tests.json")
	if err != nil {
		t.Fatalf("Failed to load alias test suite: %v", err)
	}

	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			// Setup a test event system with expected commands
			eventSystem := NewTestEventSystem(test.ExpectedCommands)

			// Create a new LuaEngine
			engine, err := NewLuaEngine(eventSystem, "")
			if err != nil {
				t.Fatalf("Failed to create LuaEngine: %v", err)
			}
			defer engine.Close()

			// Execute setup Lua code
			if err := executeSetupLua(engine, test.SetupLua); err != nil {
				t.Fatalf("Failed to execute setup Lua: %v", err)
			}

			// Process input
			line := &types.Line{
				Raw:     test.Input,
				Content: test.Input,
			}

			// Reset the commands array before processing input
			eventSystem.commands = []string{}

			engine.ProcessInput(line)

			// Wait for expected commands with timeout
			commands, ok := eventSystem.WaitForCommands(1 * time.Second)
			if !ok {
				t.Fatalf("Timed out waiting for commands. Got: %v, Expected: %v", commands, test.ExpectedCommands)
			}

			// Verify commands match expected commands in order
			assert.Equal(t, test.ExpectedCommands, commands[len(commands)-len(test.ExpectedCommands):],
				"Commands should match expected commands in order")
		})
	}
}

// TestTriggers tests trigger functionality
func TestTriggers(t *testing.T) {
	// Load the test suite
	suite, err := loadTestSuite("testdata/trigger_tests.json")
	if err != nil {
		t.Fatalf("Failed to load trigger test suite: %v", err)
	}

	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			// Setup a test event system with expected commands
			eventSystem := NewTestEventSystem(test.ExpectedCommands)

			// Create a new LuaEngine
			engine, err := NewLuaEngine(eventSystem, "")
			if err != nil {
				t.Fatalf("Failed to create LuaEngine: %v", err)
			}
			defer engine.Close()

			// Execute setup Lua code
			if err := executeSetupLua(engine, test.SetupLua); err != nil {
				t.Fatalf("Failed to execute setup Lua: %v", err)
			}

			// Process input if present
			if test.Input != "" {
				inputLine := &types.Line{
					Raw:     test.Input,
					Content: test.Input,
				}

				// Reset the commands array before processing input
				eventSystem.commands = []string{}

				engine.ProcessInput(inputLine)
			}

			// Process output
			var outputLine *types.Line
			if test.Output != "" {
				outputLine = &types.Line{
					Raw:     test.Output,
					Content: test.Output,
				}
				engine.ProcessOutput(outputLine)
			}

			// Wait for expected commands with timeout
			commands, ok := eventSystem.WaitForCommands(1 * time.Second)
			if !ok {
				t.Fatalf("Timed out waiting for commands. Got: %v, Expected: %v", commands, test.ExpectedCommands)
			}

			// Verify commands match expected commands in order
			assert.Equal(t, test.ExpectedCommands, commands[len(commands)-len(test.ExpectedCommands):],
				"Commands should match expected commands in order")

			// Check if the line should be gagged
			if outputLine != nil && test.ShouldGag {
				assert.True(t, outputLine.Flags.Gag, "Line should be gagged")
			}
		})
	}
}
