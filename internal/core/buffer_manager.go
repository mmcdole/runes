package core

import "sync"

const primaryBufferName = "primary"

type BufferManager struct {
	bufferMus         map[string]*sync.Mutex
	buffers           map[string][]string
	clientBufferMap   map[string]string
	bufferClientsMap  map[string][]string
	maxLinesPerBuffer int
}

func NewBufferManager(maxLinesPerBuffer int) *BufferManager {
	return &BufferManager{
		buffers:           map[string][]string{},
		bufferMus:         map[string]*sync.Mutex{},
		clientBufferMap:   map[string]string{},
		bufferClientsMap:  map[string][]string{},
		maxLinesPerBuffer: maxLinesPerBuffer,
	}
}

func (bm *BufferManager) AppendLine(bufferName string, line string) {
	bufferName = bm.primaryOrBufferName(bufferName)

	// Buffer doesn't exist, create it
	if _, ok := bm.buffers[bufferName]; !ok {
		bm.CreateBuffer(bufferName)
	}
	bm.bufferMus[bufferName].Lock()
	bm.buffers[bufferName] = append(bm.buffers[bufferName], line)
	if len(bm.buffers[bufferName]) > bm.maxLinesPerBuffer {
		bm.buffers[bufferName] = bm.buffers[bufferName][len(bm.buffers[bufferName])-bm.maxLinesPerBuffer:]
	}
	bm.bufferMus[bufferName].Unlock()
}

func (bm *BufferManager) AppendLines(bufferName string, lines []string) {
	bufferName = bm.primaryOrBufferName(bufferName)

	// Buffer doesn't exist, create it
	if _, ok := bm.buffers[bufferName]; !ok {
		bm.CreateBuffer(bufferName)
	}
	bm.bufferMus[bufferName].Lock()
	bm.buffers[bufferName] = append(bm.buffers[bufferName], lines...)
	if len(bm.buffers[bufferName]) > bm.maxLinesPerBuffer {
		bm.buffers[bufferName] = bm.buffers[bufferName][len(bm.buffers[bufferName])-bm.maxLinesPerBuffer:]
	}
	bm.bufferMus[bufferName].Unlock()
}

func (bm *BufferManager) CreateBuffer(bufferName string) {
	bufferName = bm.primaryOrBufferName(bufferName)
	bm.buffers[bufferName] = []string{}
	bm.bufferMus[bufferName] = &sync.Mutex{}
}

func (bm *BufferManager) GetBuffers() []string {
	buffers := []string{}
	for bufferName := range bm.buffers {
		buffers = append(buffers, bufferName)
	}
	return buffers
}

func (bm *BufferManager) GetBufferForClient(clientID string) string {
	return bm.clientBufferMap[clientID]
}

func (bm *BufferManager) GetClientsForBuffer(bufferName string) []string {
	bufferName = bm.primaryOrBufferName(bufferName)
	return bm.bufferClientsMap[bufferName]
}

func (bm *BufferManager) SwitchClientToBuffer(clientID string, bufferName string) {
	bufferName = bm.primaryOrBufferName(bufferName)
	bm.unassignClientFromBuffer(clientID)
	bm.assignClientToBuffer(clientID, bufferName)
}

func (bm *BufferManager) SetMaxLinesPerBuffer(maxLines int) {
	bm.maxLinesPerBuffer = maxLines
}

func (bm *BufferManager) GetLastLines(bufferName string, numLines int) []string {
	bufferName = bm.primaryOrBufferName(bufferName)

	// Buffer doesn't exist, return empty array
	if _, ok := bm.buffers[bufferName]; !ok {
		return []string{}
	}

	bm.bufferMus[bufferName].Lock()
	if numLines > len(bm.buffers[bufferName]) {
		numLines = len(bm.buffers[bufferName])
	}
	lines := bm.buffers[bufferName][len(bm.buffers[bufferName])-numLines:]
	bm.bufferMus[bufferName].Unlock()
	return lines
}

func (bm *BufferManager) assignClientToBuffer(clientID string, bufferName string) {
	bufferName = bm.primaryOrBufferName(bufferName)
	bm.clientBufferMap[clientID] = bufferName
	bm.bufferClientsMap[bufferName] = append(bm.bufferClientsMap[bufferName], clientID)
}

func (bm *BufferManager) unassignClientFromBuffer(clientID string) {
	bufferName := bm.clientBufferMap[clientID]
	delete(bm.clientBufferMap, clientID)
	for i, c := range bm.bufferClientsMap[bufferName] {
		if c == clientID {
			bm.bufferClientsMap[bufferName] = append(bm.bufferClientsMap[bufferName][:i], bm.bufferClientsMap[bufferName][i+1:]...)
			break
		}
	}
}

func (bm *BufferManager) primaryOrBufferName(bufferName string) string {
	if bufferName == "" {
		return primaryBufferName
	}
	return bufferName
}
