package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func usage() {
	fmt.Println("usage: tf <filename>")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}

	watcher, _ := NewTailer(5 * time.Second)
	defer watcher.close()
	watcher.addFiles(args)
	watcher.run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func() {
		<-sigs
		watcher.close()
		os.Exit(0)
	}()

	// wait for main goroutine
	<-make(chan struct{})
}
