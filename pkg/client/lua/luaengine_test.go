package lua

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mmcdole/runes/pkg/client/events"
	"github.com/mmcdole/runes/pkg/client/types"
	"github.com/stretchr/testify/assert"
)

// TestCase represents a single test case for alias or trigger functionality
type TestCase struct {
	Name             string   `json:"name"`
	SetupLua         any      `json:"setup_lua"` // Can be string or []string
	Input            string   `json:"input,omitempty"`
	Output           string   `json:"output,omitempty"`
	ExpectedCommands []string `json:"expected_commands"`
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	Tests []TestCase `json:"tests"`
}

// TestEventSystem implements the events.EventSystem interface for testing
type TestEventSystem struct {
	commands []string
	mutex    sync.Mutex
	wg       sync.WaitGroup
	expected []string
}

func NewTestEventSystem(expected []string) *TestEventSystem {
	tes := &TestEventSystem{
		commands: []string{},
		expected: expected,
	}
	
	// Initialize the wait group counter if we have expected commands
	if len(expected) > 0 {
		tes.wg.Add(1)
	}
	
	return tes
}

func (t *TestEventSystem) Subscribe(eventType events.EventType, handler events.Handler) {
	// Not needed for testing
}

func (t *TestEventSystem) Emit(event events.Event) {
	if event.Type == events.EventCommand {
		t.mutex.Lock()
		defer t.mutex.Unlock()

		if cmd, ok := event.Data.(string); ok {
			fmt.Printf("DEBUG: Received command: %s\n", cmd)
			t.commands = append(t.commands, cmd)

			// Check if we've received all expected commands in the correct order
			// Only call Done() once when we have exactly the right number of commands
			if len(t.expected) > 0 && len(t.commands) == len(t.expected) && t.hasAllExpectedInOrder() {
				fmt.Printf("DEBUG: All expected commands received, done waiting\n")
				t.wg.Done()
			}
		}
	}
}

func (t *TestEventSystem) hasAllExpectedInOrder() bool {
	if len(t.expected) == 0 {
		return true
	}

	if len(t.commands) < len(t.expected) {
		fmt.Printf("DEBUG: Not enough commands yet: have %d, need %d\n", 
			len(t.commands), len(t.expected))
		return false
	}

	// For exact matching, we need the same number of commands
	if len(t.commands) != len(t.expected) {
		fmt.Printf("DEBUG: Command count mismatch: have %d, expected %d\n", 
			len(t.commands), len(t.expected))
		return false
	}

	// Check that all commands match exactly
	for i, cmd := range t.expected {
		if t.commands[i] != cmd {
			fmt.Printf("DEBUG: Command mismatch at position %d: expected '%s', got '%s'\n", 
				i, cmd, t.commands[i])
			return false
		}
	}

	fmt.Printf("DEBUG: All commands match in order\n")
	return true
}

func (t *TestEventSystem) WaitForCommands(timeout time.Duration) ([]string, bool) {
	if len(t.expected) > 0 {
		// Don't add to the wait group here, it's already initialized in the constructor

		// Start a goroutine to handle timeout
		done := make(chan struct{})
		go func() {
			t.wg.Wait()
			close(done)
		}()

		// Wait for either completion or timeout
		select {
		case <-done:
			return t.commands, true
		case <-time.After(timeout):
			return t.commands, false
		}
	}

	return t.commands, true
}

// loadTestSuite loads a test suite from a JSON file
func loadTestSuite(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return &suite, nil
}

// executeSetupLua executes the setup Lua code for a test case
func executeSetupLua(engine *LuaEngine, setupLua any) error {
	switch v := setupLua.(type) {
	case string:
		return engine.state.DoString(v)
	case []interface{}:
		for _, script := range v {
			if scriptStr, ok := script.(string); ok {
				if err := engine.state.DoString(scriptStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
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
			if test.Output != "" {
				outputLine := &types.Line{
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
		})
	}
}
