package server

// DefaultServer is a mock server intended to respond to commands
// while clients are attached to the default session.
type DefaultServer struct {
	inputChan  chan string
	outputChan chan string
}

func NewDefaultServer() *DefaultServer {
	return &DefaultServer{
		inputChan:  make(chan string),
		outputChan: make(chan string),
	}
}

func (ds *DefaultServer) Connect() error {
	go func() {
		for {
			select {
			case input := <-ds.inputChan:
				ds.handleInput(input)
			}
		}
	}()

	return nil
}

func (ds *DefaultServer) Input() chan string {
	return ds.inputChan
}

func (ds *DefaultServer) Output() chan string {
	return ds.outputChan
}

func (ds *DefaultServer) Close() error {
	return nil
}

func (ds *DefaultServer) handleInput(input string) {
}
