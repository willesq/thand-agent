package tasks

import (
	"reflect"
	"sync"

	"github.com/thand-io/agent/internal/config"
	"github.com/thand-io/agent/internal/models"
)

func NewTaskRegistry(config *config.Config) *TaskRegistry {
	return &TaskRegistry{}
}

// TaskRegistry manages custom task handlers
type TaskRegistry struct {
	handlers map[reflect.Type]models.TaskHandler
	mu       sync.RWMutex
}

// Global registry instance
var globalTaskRegistry = &TaskRegistry{
	handlers: make(map[reflect.Type]models.TaskHandler),
}

// RegisterTaskHandler registers a custom handler for a specific task type
func RegisterTaskHandler(taskType any, handler models.TaskHandler) {
	globalTaskRegistry.mu.Lock()
	defer globalTaskRegistry.mu.Unlock()

	t := reflect.TypeOf(taskType)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		t = reflect.PointerTo(t) // Ensure we store pointer types
	}
	globalTaskRegistry.handlers[t] = handler
}

// GetTaskHandler retrieves a handler for a specific task type
func (tr *TaskRegistry) GetTaskHandler(taskType reflect.Type) (models.TaskHandler, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	handler, exists := tr.handlers[taskType]
	return handler, exists
}
