package utils

import (
	"errors"
	"fmt"
	"sync/atomic"
)

type CleanupStack struct {
	stack     []func() error
	triggered *atomic.Bool
}

// NewCleanupStack creates a new cleanup stack.
func NewCleanupStack() *CleanupStack {
	triggered := &atomic.Bool{}
	triggered.Store(false)

	return &CleanupStack{
		stack:     nil,
		triggered: triggered,
	}
}

// Push pushes the cleanup function to the stack.
func (cs *CleanupStack) Push(cleanup func() error) {
	if cs.triggered.Load() {
		return
	}

	cs.stack = append(cs.stack, cleanup)
}

// Cleanup cleans up the stack.
func (cs *CleanupStack) Cleanup() (err error) {
	if cs.stack == nil || cs.triggered.Load() {
		return
	}

	cs.triggered.Store(true)

	for i := len(cs.stack) - 1; i >= 0; i-- {
		cleanup := cs.stack[i]
		if cleanupErr := cleanup(); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("cleanup: %w", cleanupErr))
		}
	}

	return
}
