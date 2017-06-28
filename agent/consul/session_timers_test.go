package consul

import (
	"testing"
	"time"
)

func TestSessionTimers(t *testing.T) {
	m := NewSessionTimers()
	ch := make(chan int)
	newTm := func(d time.Duration) *time.Timer {
		return time.AfterFunc(d, func() { ch <- 1 })
	}

	waitForTimer := func() {
		select {
		case <-ch:
			return
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timer did not fire")
		}
	}

	// check that non-existent id returns nil
	if _, ok := m.Get("foo"); ok {
		t.Fatalf("got %v want %v", ok, false)
	}

	// add a timer and look it up and delete it
	tm := newTm(time.Millisecond)
	m.Set("foo", tm)
	if got, want := m.Len(), 1; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	gottm, ok := m.Get("foo")
	if got, want := ok, true; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := gottm, tm; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	m.Del("foo")
	if _, ok := m.Get("foo"); ok {
		t.Fatalf("got %v want %v", ok, false)
	}
	waitForTimer()

	// create timer via ResetOrCreate
	m.ResetOrCreate("foo", time.Millisecond, func() { ch <- 1 })
	if got, want := m.Len(), 1; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	waitForTimer()

	// timer is still there
	if got, want := m.Len(), 1; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}

	// reset the timer and check that it fires again
	m.ResetOrCreate("foo", time.Millisecond, nil)
	waitForTimer()

	// reset the timer with a long ttl and then stop it
	m.ResetOrCreate("foo", 20*time.Millisecond, func() { ch <- 1 })
	m.Stop("foo")
	select {
	case <-ch:
		t.Fatal("timer fired although it shouldn't")
	case <-time.After(100 * time.Millisecond):
		// want
	}

	// stopping a stopped timer should not break
	m.Stop("foo")

	// stop should also remove the timer
	if got, want := m.Len(), 0; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}

	// create two timers and stop and then stop all
	m.ResetOrCreate("foo1", 20*time.Millisecond, func() { ch <- 1 })
	m.ResetOrCreate("foo2", 30*time.Millisecond, func() { ch <- 2 })
	m.StopAll()
	select {
	case x := <-ch:
		t.Fatalf("timer %d fired although it shouldn't", x)
	case <-time.After(100 * time.Millisecond):
		// want
	}

	// stopall should remove all timers
	if got, want := m.Len(), 0; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
}
