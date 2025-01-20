package lua

// Registry table names
const (
	// Output/Input listener tables
	OutputListenerTable = "__output_listeners"
	InputListenerTable  = "__input_listeners"

	// Event handler tables  
	OnConnectTable    = "__connect_handlers"
	OnDisconnectTable = "__disconnect_handlers"

	// Script management
	ScriptResetTable = "__script_reset_handlers"
)
