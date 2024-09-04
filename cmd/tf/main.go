package main

import (
	"fmt"
	"os"
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

	watcher, _ := NewWatcher(5 * time.Second)
	defer watcher.close()
	watcher.addFiles(args)
	watcher.run()

	// wait for main goroutine
	<-make(chan struct{})
}
