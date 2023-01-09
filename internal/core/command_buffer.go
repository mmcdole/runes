package core

import (
	"fmt"

	"github.com/fatih/color"
)

type BufferCommand struct{}

func (c *BufferCommand) Execute(params *CommandParams) bool {
	return c.handleBufferListCommand(params)
}

func (c *BufferCommand) handleBufferListCommand(params *CommandParams) bool {
	activeBuffer := params.Session.bufferManager.GetBufferForClient(params.Executor.ID())
	buffers := params.Session.bufferManager.GetBuffers()

	params.writeToExecutor("Buffers: ")
	for _, buffer := range buffers {
		if buffer == activeBuffer {
			params.writeToExecutor(fmt.Sprintf(" [%s] %s", color.HiGreenString("*"), buffer))
		} else {
			params.writeToExecutor(fmt.Sprintf(" [ ] %s", buffer))
		}
	}
	return true
}

func (c *BufferCommand) handleBufferSwitchCommand(params *CommandParams) bool {
	return true
}

func (c *BufferCommand) handleBufferWriteCommand(params *CommandParams) bool {
	// params.Session.writeBufferText(bufferName string, text string)
	return true
}

func (c *BufferCommand) Usage() string {
	return "Buffer Usage!"
}

func (c *BufferCommand) Help() string {
	return "Buffer Help!"
}
