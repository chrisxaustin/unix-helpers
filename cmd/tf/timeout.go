package main

import (
	"time"
)

type IdleTimer struct {
	timeout  time.Duration
	timer    *time.Timer
	textSeen <-chan bool
	callback func()
}

func NewIdleTimer(delay time.Duration, textSeen <-chan bool, callback func()) *IdleTimer {
	timeout := IdleTimer{
		timeout:  delay,
		timer:    time.NewTimer(delay),
		textSeen: textSeen,
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

			case <-idleTimer.textSeen:
				activityStarted = true
				idleTimer.timer.Reset(idleTimer.timeout)
			}
		}
	}()
}
