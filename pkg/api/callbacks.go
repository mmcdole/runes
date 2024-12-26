package api

// LuaCallbacks defines the interface for callbacks that Lua scripts can invoke
type LuaCallbacks interface {
	Output(text string, buffer string)
	Connect(host string, port int) error
	Disconnect()
	ListBuffers() []string
	SwitchBuffer(name string)
}
