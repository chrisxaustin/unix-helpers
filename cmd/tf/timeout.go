package main

import (
	"time"
)

type IdleTimer struct {
	timeout  time.Duration
	timer    *time.Timer
	reset    chan bool
	callback func()
}

func NewIdleTimer(delay time.Duration, callback func()) *IdleTimer {
	timeout := IdleTimer{
		timeout:  delay,
		timer:    time.NewTimer(delay),
		reset:    make(chan bool),
		callback: callback,
	}
	timeout.run()
	return &timeout
}

func (idleTimer *IdleTimer) run() {
	activityStarted := false
	go func() {
		for {
			select {
			case <-idleTimer.timer.C:
				if activityStarted {
					idleTimer.callback()
				}

			case <-idleTimer.reset:
				activityStarted = true
				idleTimer.timer.Reset(idleTimer.timeout)
			}
		}
	}()
}
