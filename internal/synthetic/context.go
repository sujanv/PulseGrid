package synthetic

import (
	"fmt"
	"strings"
	"sync"
)

type ExecutionContext struct {
	mu        sync.RWMutex
	variables map[string]string
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		variables: make(map[string]string),
	}
}

func (c *ExecutionContext) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.variables[key] = value
}

func (c *ExecutionContext) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.variables[key]
	return val, ok
}

func (c *ExecutionContext) Interpolate(input string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := input
	for k, v := range c.variables {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}
