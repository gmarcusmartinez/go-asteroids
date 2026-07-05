package engine

import (
	"testing"
	"time"
)

func TestTimerReadyResetCycle(t *testing.T) {
	timer := NewTimer(100 * time.Millisecond)

	if timer.IsReady() {
		t.Fatal("timer should not be ready before any updates")
	}

	// Advance well past the target tick count.
	for range 100 {
		timer.Update()
	}

	if !timer.IsReady() {
		t.Fatal("timer should be ready after enough updates")
	}

	timer.Reset()
	if timer.IsReady() {
		t.Fatal("timer should not be ready immediately after reset")
	}
}
