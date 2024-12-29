package luaengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/mmcdole/runes/pkg/events"
)

// Define the test case type at package level
type testCase struct {
	Name             string         `json:"name"`
	SetupLua         any            `json:"setup_lua"`
	Input            string         `json:"input,omitempty"`
	Output           string         `json:"output,omitempty"`
	ExpectedCommands []string       `json:"expected_commands,omitempty"`
	ExpectedEvents   []events.Event `json:"expected_events,omitempty"`
}

type testDataFile struct {
	Tests []testCase `json:"tests"`
}

type mockEventCollector struct {
	sync.Mutex
	events []events.Event
}

func newMockEventCollector() *mockEventCollector {
	return &mockEventCollector{
		events: make([]events.Event, 0),
	}
}

func (m *mockEventCollector) collect(event events.Event) {
	m.Lock()
	defer m.Unlock()
	m.events = append(m.events, event)
}

// setupTest creates a test environment and returns a cleanup function
func setupTest(t *testing.T) (*LuaEngine, *mockEventCollector, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "luaengine_test")
	if err != nil {
		t.Fatal("Failed to create temp directory:", err)
	}

	eventSystem := events.New()
	collector := newMockEventCollector()

	// Subscribe to processed events
	eventSystem.Subscribe(events.EventCommand, collector.collect)
	eventSystem.Subscribe(events.EventOutput, collector.collect)
	eventSystem.Subscribe(events.EventConnect, collector.collect)
	eventSystem.Subscribe(events.EventDisconnect, collector.collect)

	engine := New(coreLuaScripts, tempDir, eventSystem)
	if err := engine.Initialize(); err != nil {
		t.Fatal("Failed to initialize engine:", err)
	}

	cleanup := func() {
		engine.Close()
		os.RemoveAll(tempDir)
	}

	return engine, collector, cleanup
}

func loadTestData(t *testing.T, filename string) testDataFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("Failed to read test data %s: %v", filename, err)
	}

	var testData testDataFile
	if err := json.Unmarshal(data, &testData); err != nil {
		t.Fatalf("Failed to parse test data %s: %v", filename, err)
	}
	return testData
}

// executeSetupLua handles both string and []string Lua setup code
func executeSetupLua(t *testing.T, engine *LuaEngine, setup any) {
	t.Helper()
	switch lua := setup.(type) {
	case string:
		if err := engine.L.DoString(lua); err != nil {
			t.Fatalf("Failed to execute setup Lua code: %v", err)
		}
	case []interface{}:
		for _, cmd := range lua {
			if err := engine.L.DoString(cmd.(string)); err != nil {
				t.Fatalf("Failed to execute setup Lua code: %v", err)
			}
		}
	}
}

// executeTest runs a single test case and returns pass/fail status
func executeTest(t *testing.T, feature string, tt testCase) {
	t.Helper()
	testName := fmt.Sprintf("%s/%s", feature, tt.Name)
	t.Run(testName, func(t *testing.T) {
		engine, collector, cleanup := setupTest(t)
		defer cleanup()

		if tt.SetupLua != nil {
			executeSetupLua(t, engine, tt.SetupLua)
		}

		if tt.Input != "" {
			engine.eventSystem.Emit(events.Event{
				Type: events.EventRawInput,
				Data: tt.Input,
			})
		}
		if tt.Output != "" {
			engine.eventSystem.Emit(events.Event{
				Type: events.EventRawOutput,
				Data: tt.Output,
			})
		}

		if tt.ExpectedEvents != nil {
			assertEvents(t, collector, tt.ExpectedEvents)
		}
		if tt.ExpectedCommands != nil {
			assertCommands(t, collector, tt.ExpectedCommands)
		}
	})
}

func assertEvents(t *testing.T, collector *mockEventCollector, expected []events.Event) {
	t.Helper()
	collector.Lock()
	defer collector.Unlock()

	if len(collector.events) != len(expected) {
		// Only show debug output if there's a mismatch
		fmt.Printf("\nExpected Events (%d):\n", len(expected))
		for i, exp := range expected {
			fmt.Printf("  %d: Type=%q Data=%q\n", i, exp.Type, exp.Data)
		}

		fmt.Printf("\nActual Events (%d):\n", len(collector.events))
		for i, got := range collector.events {
			fmt.Printf("  %d: Type=%q Data=%q\n", i, got.Type, got.Data)
		}

		t.Errorf("expected %d events, got %d", len(expected), len(collector.events))
		return
	}

	for i, exp := range expected {
		got := collector.events[i]
		if got.Type != exp.Type || !compareEventData(t, i, exp, got) {
			// Show debug output for mismatched events
			fmt.Printf("\nMismatch at event %d:\n", i)
			fmt.Printf("Expected: Type=%q Data=%q\n", exp.Type, exp.Data)
			fmt.Printf("Got:      Type=%q Data=%q\n", got.Type, got.Data)
			t.Errorf("event %d: mismatch", i)
		}
	}
}

// compareEventData compares the data field of two events based on their type
func compareEventData(t *testing.T, index int, expected, got events.Event) bool {
	t.Helper()

	switch expected.Type {
	case events.EventCommand:
		if expected.Data != got.Data {
			t.Errorf("event %d: expected command %q, got %q", index, expected.Data, got.Data)
			return false
		}
	case events.EventOutput:
		if expected.Data != got.Data {
			t.Errorf("event %d: expected output %q, got %q", index, expected.Data, got.Data)
			return false
		}
	case events.EventConnect, events.EventDisconnect:
		// These events typically don't carry data, so no comparison needed
		return true
	default:
		t.Errorf("event %d: unexpected event type %q", index, expected.Type)
		return false
	}
	return true
}

// assertCommands verifies commands are received in order with timeout
func assertCommands(t *testing.T, collector *mockEventCollector, expected []string) {
	t.Helper()

	collector.Lock()
	actualCommands := make([]string, 0)
	for _, event := range collector.events {
		if event.Type == events.EventCommand {
			actualCommands = append(actualCommands, event.Data.(string))
		}
	}
	collector.Unlock()

	if len(actualCommands) != len(expected) {
		// Only show debug output if there's a mismatch
		fmt.Printf("\nExpected Commands (%d):\n", len(expected))
		for i, cmd := range expected {
			fmt.Printf("  %d: %q\n", i, cmd)
		}

		fmt.Printf("\nActual Commands (%d):\n", len(actualCommands))
		for i, cmd := range actualCommands {
			fmt.Printf("  %d: %q\n", i, cmd)
		}

		t.Errorf("expected %d commands, got %d", len(expected), len(actualCommands))
		return
	}

	for i, exp := range expected {
		if actualCommands[i] != exp {
			// Show debug output for mismatched commands
			fmt.Printf("\nMismatch at command %d:\n", i)
			fmt.Printf("Expected: %q\n", exp)
			fmt.Printf("Got:      %q\n", actualCommands[i])
			t.Errorf("command %d: expected %q, got %q", i, exp, actualCommands[i])
		}
	}
}

// TestFeatures runs all feature tests from JSON files
func TestFeatures(t *testing.T) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "_tests.json") {
			feature := strings.TrimSuffix(file.Name(), "_tests.json")
			t.Run(feature, func(t *testing.T) {
				testData := loadTestData(t, file.Name())

				for _, tt := range testData.Tests {
					t.Run(tt.Name, func(t *testing.T) {
						executeTest(t, feature, tt)
					})
				}
			})
		}
	}
}
