package lua

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mmcdole/runes/pkg/client/events"
)

// TestCase represents a single test case for alias or trigger functionality
type TestCase struct {
	Name             string   `json:"name"`
	SetupLua         any      `json:"setup_lua"` // Can be string or []string
	Input            string   `json:"input,omitempty"`
	Output           string   `json:"output,omitempty"`
	ExpectedCommands []string `json:"expected_commands"`
	ShouldGag        bool     `json:"should_gag,omitempty"`
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

// NewTestEventSystem creates a new test event system with expected commands
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

// Subscribe implements the events.EventSystem interface
func (t *TestEventSystem) Subscribe(eventType events.EventType, handler events.Handler) {
	// No-op for testing
}

// Emit implements the events.EventSystem interface
func (t *TestEventSystem) Emit(event events.Event) {
	if event.Type != events.EventCommand {
		return
	}

	command, ok := event.Data.(string)
	if !ok {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.commands = append(t.commands, command)

	// Check if we have all expected commands in order
	if t.hasAllExpectedInOrder() {
		t.wg.Done()
	}
}

// hasAllExpectedInOrder checks if all expected commands are present in order
func (t *TestEventSystem) hasAllExpectedInOrder() bool {
	if len(t.expected) == 0 {
		return true
	}

	if len(t.commands) < len(t.expected) {
		return false
	}

	// Check if the last N commands match the expected commands
	startIdx := len(t.commands) - len(t.expected)
	for i, cmd := range t.expected {
		if t.commands[startIdx+i] != cmd {
			return false
		}
	}

	return true
}

// WaitForCommands waits for expected commands with a timeout
func (t *TestEventSystem) WaitForCommands(timeout time.Duration) ([]string, bool) {
	if len(t.expected) == 0 {
		return t.commands, true
	}

	// Create a channel that will be closed when the wait group is done
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	// Wait for either the wait group to be done or the timeout to expire
	select {
	case <-done:
		t.mutex.Lock()
		defer t.mutex.Unlock()
		return t.commands, true
	case <-time.After(timeout):
		t.mutex.Lock()
		defer t.mutex.Unlock()
		return t.commands, false
	}
}

// loadTestSuite loads a test suite from a JSON file
func loadTestSuite(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test suite file: %w", err)
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test suite: %w", err)
	}

	return &suite, nil
}

// executeSetupLua executes the setup Lua code for a test case
func executeSetupLua(engine *LuaEngine, setupLua any) error {
	switch v := setupLua.(type) {
	case string:
		if err := engine.state.DoString(v); err != nil {
			return fmt.Errorf("failed to execute Lua: %w", err)
		}
	case []any:
		for _, code := range v {
			if codeStr, ok := code.(string); ok {
				if err := engine.state.DoString(codeStr); err != nil {
					return fmt.Errorf("failed to execute Lua: %w", err)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported setup Lua type: %T", setupLua)
	}

	return nil
}
