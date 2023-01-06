package core

const primaryBufferName = "primary"

type BufferManager struct {
	bufferLines       map[string][]string
	clientBufferMap   map[string]string
	bufferClientsMap  map[string][]string
	maxLinesPerBuffer int
}

func NewBufferManager(maxLinesPerBuffer int) *BufferManager {
	return &BufferManager{
		bufferLines:       map[string][]string{},
		clientBufferMap:   map[string]string{},
		bufferClientsMap:  map[string][]string{},
		maxLinesPerBuffer: maxLinesPerBuffer,
	}
}

func (bm *BufferManager) AppendLine(bufferName string, line string) {
	bufferName = bm.primaryOrBufferName(bufferName)

	if _, ok := bm.bufferLines[bufferName]; !ok {
		bm.CreateBuffer(bufferName)
	}
	bm.bufferLines[bufferName] = append(bm.bufferLines[bufferName], line)
	if len(bm.bufferLines[bufferName]) > bm.maxLinesPerBuffer {
		bm.bufferLines[bufferName] = bm.bufferLines[bufferName][len(bm.bufferLines[bufferName])-bm.maxLinesPerBuffer:]
	}
}

func (bm *BufferManager) AppendLines(bufferName string, lines []string) {
	bufferName = bm.primaryOrBufferName(bufferName)

	if _, ok := bm.bufferLines[bufferName]; !ok {
		bm.CreateBuffer(bufferName)
	}
	bm.bufferLines[bufferName] = append(bm.bufferLines[bufferName], lines...)
	if len(bm.bufferLines[bufferName]) > bm.maxLinesPerBuffer {
		bm.bufferLines[bufferName] = bm.bufferLines[bufferName][len(bm.bufferLines[bufferName])-bm.maxLinesPerBuffer:]
	}
}

func (bm *BufferManager) CreateBuffer(bufferName string) {
	bufferName = bm.primaryOrBufferName(bufferName)
	bm.bufferLines[bufferName] = []string{}
}

func (bm *BufferManager) GetBuffers() []string {
	buffers := []string{}
	for bufferName := range bm.bufferLines {
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
	if numLines > len(bm.bufferLines[bufferName]) {
		numLines = len(bm.bufferLines[bufferName])
	}
	return bm.bufferLines[bufferName][len(bm.bufferLines[bufferName])-numLines:]
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
