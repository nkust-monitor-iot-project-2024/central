package utils_test

import (
	"errors"
	"testing"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
)

func TestCleanupStack_NoFunction(t *testing.T) {
	cs := utils.NewCleanupStack()

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}
}

func TestCleanupStack_OneFunction(t *testing.T) {
	cs := utils.NewCleanupStack()

	var called bool
	cs.Push(func() error {
		called = true
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}
	if !called {
		t.Error("cleanup function not called")
	}
}

func TestCleanupStack_MultipleFunctions(t *testing.T) {
	cs := utils.NewCleanupStack()

	var called1, called2 bool
	cs.Push(func() error {
		called1 = true
		return nil
	})
	cs.Push(func() error {
		called2 = true
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}
	if !called1 {
		t.Error("cleanup function 1 not called")
	}
	if !called2 {
		t.Error("cleanup function 2 not called")
	}
}

func TestCleanupStack_ContinueEvenError(t *testing.T) {
	cs := utils.NewCleanupStack()

	var called1, called2 bool
	cs.Push(func() error {
		called1 = true
		return nil
	})
	cs.Push(func() error {
		called2 = true
		return errors.New("error")
	})

	if err := cs.Cleanup(); err == nil {
		t.Error("CleanupStack.Cleanup() = nil; want error")
	}
	if !called1 {
		t.Error("cleanup function 1 not called")
	}
	if !called2 {
		t.Error("cleanup function 2 not called")
	}
}

func TestCleanupStack_NoDuplicateTrigger(t *testing.T) {
	cs := utils.NewCleanupStack()

	var called1, called2 int
	cs.Push(func() error {
		called1 += 1
		return nil
	})
	cs.Push(func() error {
		called2 += 1
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}
	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}

	if called1 != 1 {
		t.Error("cleanup function 1 more than once")
	}
	if called2 != 1 {
		t.Error("cleanup function 2 more than once")
	}
}

func TestCleanupStack_IgnoreNewFunctionAfterCleanup(t *testing.T) {
	cs := utils.NewCleanupStack()

	var called1, called2, called3 bool
	cs.Push(func() error {
		called1 = true
		return nil
	})

	cs.Push(func() error {
		called2 = true
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}

	cs.Push(func() error {
		t.Error("new cleanup function called")
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}
	if !called1 {
		t.Error("cleanup function 1 not called")
	}
	if !called2 {
		t.Error("cleanup function 2 not called")
	}

	cs.Push(func() error {
		called3 = true
		return nil
	})

	if err := cs.Cleanup(); err != nil {
		t.Errorf("CleanupStack.Cleanup() = %v; want nil", err)
	}

	if called3 {
		t.Error("cleanup function 3 is called")
	}
}
