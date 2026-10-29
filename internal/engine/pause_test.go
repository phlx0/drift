package engine

import (
	"testing"
	"time"
)

func TestPauseResumeFrameClock(t *testing.T) {
	e := &Engine{interval: time.Millisecond}

	e.resume()
	if e.paused || e.tick == nil || e.ticker == nil {
		t.Fatal("resume should start the frame clock")
	}
	select {
	case <-e.tick:
	case <-time.After(time.Second):
		t.Fatal("no tick arrived after resume")
	}

	e.pause()
	if !e.paused || e.tick != nil || e.ticker != nil {
		t.Fatal("pause should stop the frame clock and nil the tick channel")
	}
	e.pause() // idempotent

	e.resume()
	if e.paused || e.tick == nil {
		t.Fatal("resume after pause should restart the frame clock")
	}
	e.pause()
}
