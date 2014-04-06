package grafic2d

import (
	"syscall"
)

type Timer struct {
	start      int64
	lastCall   int64
	fpsBase    int64
	fpsCounter int
	fps        int
	isStarted  bool
}

func (t *Timer) Start() {
	t.isStarted = true
	now := syscall.Timeval{0.0, 0.0}
	syscall.Gettimeofday(&now)
	t.start = (int64(now.Sec)*1e3 + int64(now.Usec)/1e3)
	t.lastCall = t.start
	t.fpsBase = t.start
	t.fpsCounter = 0
}

func (t *Timer) Reset() {
	t.isStarted = false
	t.start = 0
	t.lastCall = 0
	t.fpsBase = 0
	t.fpsCounter = 0
}


func (t *Timer) TimeSinceLastCall() int {
	if !t.isStarted {
		return 0
	}

	// calc timing stuff
	n := syscall.Timeval{0.0, 0.0}
	syscall.Gettimeofday(&n)
	now := (int64(n.Sec)*1e3 + int64(n.Usec)/1e3)
	diff := int(now - t.lastCall)
	t.lastCall = now

	return diff
}

func (t *Timer) TimeSinceStart() int {
	if !t.isStarted {
		return 0
	}

	n := syscall.Timeval{0.0, 0.0}
	syscall.Gettimeofday(&n)
	now := (int64(n.Sec)*1e3 + int64(n.Usec)/1e3)
	diff := int(now - t.start)
	return diff
}

func (t *Timer) CallsPerSec() int {
	if !t.isStarted {
		return 0
	}

	// calc timing stuff
	n := syscall.Timeval{0.0, 0.0}
	syscall.Gettimeofday(&n)
	now := (int64(n.Sec)*1e3 + int64(n.Usec)/1e3)

	// calc fps stuff
	t.fpsCounter++
	if now-t.fpsBase > 1000 {
		t.fps = (t.fpsCounter * 1000) / int(now-t.fpsBase)
		t.fpsBase = now
	}
	
	return t.fps
}
