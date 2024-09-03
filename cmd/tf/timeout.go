package main

import (
	"fmt"
	"time"
)

type Timeout struct {
	timeout time.Duration
	timer   *time.Timer
	reset   chan bool
}

func NewTimeout(delay time.Duration) *Timeout {
	timeout := Timeout{
		timeout: delay,
		timer:   time.NewTimer(delay),
		reset:   make(chan bool),
	}
	timeout.run()
	return &timeout
}

func (self *Timeout) run() {
	activityStarted := false
	go func() {
		for {
			select {
			case <-self.timer.C:
				if activityStarted {
					fmt.Println("----------------------------------------")
				}

			case <-self.reset:
				activityStarted = true
				self.timer.Reset(self.timeout)
			}
		}
	}()
}
