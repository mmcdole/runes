package luaengine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Define the test case type at package level
type testCase struct {
	Name             string   `json:"name"`
	SetupLua         any      `json:"setup_lua"`
	Input            string   `json:"input,omitempty"`
	Output           string   `json:"output,omitempty"`
	ExpectedCommands []string `json:"expected_commands,omitempty"`
}

type testDataFile struct {
	Tests []testCase `json:"tests"`
}

type mockCallbacks struct {
	outputs []string
	buffers []string
}

func (m *mockCallbacks) Output(text string, buffer string) {
	m.outputs = append(m.outputs, text)
	m.buffers = append(m.buffers, buffer)
}

func (m *mockCallbacks) Connect(host string, port int) error { return nil }
func (m *mockCallbacks) Disconnect()                         {}
func (m *mockCallbacks) ListBuffers() []string               { return []string{} }
func (m *mockCallbacks) SwitchBuffer(name string)            {}

// setupTest creates a test environment and returns a cleanup function
func setupTest(t *testing.T) (*LuaEngine, <-chan string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "luaengine_test")
	if err != nil {
		t.Fatal("Failed to create temp directory:", err)
	}

	callbacks := &mockCallbacks{}
	engine := New(coreLuaScripts, tempDir, callbacks)
	fmt.Println("Creating new engine...")

	if err := engine.Initialize(); err != nil {
		t.Fatal("Failed to initialize engine:", err)
	}
	fmt.Println("Engine initialized")

	cleanup := func() {
		engine.Close()
		os.RemoveAll(tempDir)
	}

	return engine, engine.GetSentCommands(), cleanup
}

// assertCommands verifies commands are received in order with timeout
func assertCommands(t *testing.T, ch <-chan string, expected []string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for _, exp := range expected {
		select {
		case cmd := <-ch:
			if cmd != exp {
				t.Errorf("expected command %q, got %q", exp, cmd)
				return
			}
		case <-ctx.Done():
			t.Errorf("timeout waiting for command %q", exp)
			return
		}
	}

	// Verify no additional commands were sent
	select {
	case cmd := <-ch:
		t.Errorf("unexpected extra command: %q", cmd)
	case <-time.After(50 * time.Millisecond):
		// Expected - no more commands
	}
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
	fmt.Printf("=== RUN   %s\n", testName)

	engine, sentCommands, cleanup := setupTest(t)
	defer cleanup()

	// Add debug output for setup
	if tt.SetupLua != nil {
		fmt.Printf("Executing setup for %s\n", testName)
		executeSetupLua(t, engine, tt.SetupLua)
		fmt.Printf("Setup complete for %s\n", testName)
	}

	if tt.Input != "" {
		fmt.Printf("Emitting input: %s\n", tt.Input)
		engine.EmitEvent("input", tt.Input)
	}
	if tt.Output != "" {
		fmt.Printf("Emitting output: %s\n", tt.Output)
		engine.EmitEvent("output", tt.Output)
	}

	if tt.ExpectedCommands != nil {
		assertCommands(t, sentCommands, tt.ExpectedCommands)
	}

	if !t.Failed() {
		fmt.Printf("--- PASS: %s\n", testName)
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
