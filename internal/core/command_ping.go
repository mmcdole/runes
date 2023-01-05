package core

type PingCommand struct{}

func (c *PingCommand) Execute(params *CommandParams) bool {
	params.Session.writeText("Pong!")
	return true
}

func (c *PingCommand) Usage() string {
	return "Ping Usage!"
}

func (c *PingCommand) Help() string {
	return "Ping Help!"
}
