package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func (watcher *Tailer) tail(fh *os.File) {
	watcher.idleTimeout.reset <- true
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		watcher.idleTimeout.reset <- true
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

type Tailer struct {
	watchedFiles map[string]*os.File
	watchingDir  map[string]bool
	watchingFile map[string]bool
	fileWatcher  *fsnotify.Watcher
	dirWatcher   *fsnotify.Watcher
	idleTimeout  *IdleTimer
}

func NewWatcher(timeout time.Duration) (*Tailer, bool) {
	watcher := Tailer{
		watchingDir:  make(map[string]bool),
		watchingFile: make(map[string]bool),
		watchedFiles: make(map[string]*os.File),
		idleTimeout: NewIdleTimer(timeout, func() {
			fmt.Println("----------------------------------------")
		}),
	}
	var err error
	watcher.fileWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, false
	}
	watcher.dirWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, false
	}

	return &watcher, true
}

func (watcher *Tailer) openFile(name string, seek bool) bool {
	watcher.watchingFile[name] = true
	fh, err := os.Open(name)
	if err != nil {
		return false
	}
	// defer fh.Close()

	if seek {
		_, err = fh.Seek(0, io.SeekEnd)
		if err != nil {
			return false
		}
	}
	watcher.watchedFiles[name] = fh
	if !seek {
		watcher.tail(fh)
	}
	return true
}

func (watcher *Tailer) addFile(name string, seek bool) {
	watcher.watchingFile[name] = true
	watcher.openFile(name, seek)
	watcher.fileWatcher.Add(name)
	parent := filepath.Dir(name)
	if !watcher.watchingDir[parent] {
		watcher.watchingDir[parent] = true
		watcher.dirWatcher.Add(parent)
	}
}

func (watcher *Tailer) addFiles(filenames []string) {
	for _, name := range filenames {
		watcher.addFile(name, true)
	}
}

func (watcher *Tailer) close() {
	for _, watched := range watcher.watchedFiles {
		watched.Close()
	}
	if watcher.fileWatcher != nil {
		watcher.fileWatcher.Close()
	}
	if watcher.dirWatcher != nil {
		watcher.dirWatcher.Close()
	}
}

func (watcher *Tailer) run() {
	go func() {
		for {
			select {
			case event, ok := <-watcher.fileWatcher.Events:
				if !ok {
					return
				}

				watched := watcher.watchedFiles[event.Name]
				switch {
				case event.Op.Has(fsnotify.Write):
					watcher.tail(watched)
				case event.Op == fsnotify.Chmod:
				}

			case err, ok := <-watcher.fileWatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.dirWatcher.Events:
				if !ok {
					return
				}
				switch {
				case event.Op.Has(fsnotify.Rename):
					watcher.addFile(event.Name, false)
				case event.Op.Has(fsnotify.Create) && watcher.watchingFile[event.Name]:
					watcher.addFile(event.Name, false)
				}

			case err, ok := <-watcher.dirWatcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

}
