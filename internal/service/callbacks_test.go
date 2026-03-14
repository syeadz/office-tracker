package service

import (
	"sync"
	"testing"
	"time"
)

// Test that the registered attendance callback is invoked when triggered.
func TestTriggerAttendanceChangeCallback(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	SetAttendanceChangeCallback(func() {
		defer wg.Done()
	})

	TriggerAttendanceChangeCallback()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("attendance callback was not invoked")
	}

	// clear callback
	SetAttendanceChangeCallback(nil)
}

// Test that setting nil disables the callback.
func TestSetNilCallbackPreventsInvocation(t *testing.T) {
	called := false
	SetAttendanceChangeCallback(nil)

	// set a callback then clear it to ensure Trigger does nothing
	SetAttendanceChangeCallback(func() { called = true })
	SetAttendanceChangeCallback(nil)

	TriggerAttendanceChangeCallback()
	// wait briefly to let any goroutine run
	time.Sleep(50 * time.Millisecond)
	if called {
		t.Fatal("callback should not have been invoked after being cleared")
	}
}

func TestTriggerEnvironmentChangeCallback(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	SetEnvironmentChangeCallback(func() {
		defer wg.Done()
	})

	TriggerEnvironmentChangeCallback()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("environment callback was not invoked")
	}

	SetEnvironmentChangeCallback(nil)
}
