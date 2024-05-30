package utils

import (
	"errors"
	"fmt"
)

type CleanupStack struct {
	stack []func() error
}

// NewCleanupStack creates a new cleanup stack.
func NewCleanupStack() *CleanupStack {
	return &CleanupStack{
		stack: nil,
	}
}

// Push pushes the cleanup function to the stack.
func (cs *CleanupStack) Push(cleanup func() error) {
	cs.stack = append(cs.stack, cleanup)
}

// Cleanup cleans up the stack.
func (cs *CleanupStack) Cleanup() (err error) {
	if cs.stack == nil {
		return
	}

	for i := len(cs.stack) - 1; i >= 0; i-- {
		cleanup := cs.stack[i]
		if cleanupErr := cleanup(); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("cleanup: %w", cleanupErr))
		}
	}

	return
}
