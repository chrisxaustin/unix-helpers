package main

import (
	"fmt"
	"time"
)

type IdleTimer struct {
	timeout time.Duration
	timer   *time.Timer
	reset   chan bool
}

func NewIdleTimer(delay time.Duration) *IdleTimer {
	timeout := IdleTimer{
		timeout: delay,
		timer:   time.NewTimer(delay),
		reset:   make(chan bool),
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
					fmt.Println("----------------------------------------")
				}

			case <-idleTimer.reset:
				activityStarted = true
				idleTimer.timer.Reset(idleTimer.timeout)
			}
		}
	}()
}
