package client

import (
	"sync"
)

const (
	MainBuffer = "main"
)

type BufferManager struct {
	buffers     map[string]*Buffer
	currentName string
	mu          sync.RWMutex
}

type Buffer struct {
	Name    string
	Lines   []string
	Visible bool
}

func NewBufferManager() *BufferManager {
	bm := &BufferManager{
		buffers:     make(map[string]*Buffer),
		currentName: MainBuffer,
	}
	bm.buffers[MainBuffer] = &Buffer{Name: MainBuffer, Visible: true}
	return bm
}

func (bm *BufferManager) AddLine(text string, bufferName string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bufferName == "" {
		bufferName = bm.currentName
	}

	if _, exists := bm.buffers[bufferName]; !exists {
		bm.buffers[bufferName] = &Buffer{Name: bufferName, Visible: true}
	}

	bm.buffers[bufferName].Lines = append(bm.buffers[bufferName].Lines, text)
}

func (bm *BufferManager) SwitchBuffer(name string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.buffers[name]; !exists {
		bm.buffers[name] = &Buffer{Name: name, Visible: true}
	}
	bm.currentName = name
}

func (bm *BufferManager) GetCurrentBuffer() *Buffer {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.buffers[bm.currentName]
}

func (bm *BufferManager) ListBuffers() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	buffers := make([]string, 0, len(bm.buffers))
	for name := range bm.buffers {
		buffers = append(buffers, name)
	}
	return buffers
}
