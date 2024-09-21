// DirWatcher uses fsnotify to listen for changes within any of the observed directories.
//
// When a Create or Rename event is observed it will send the name of the file to the fileCreated channel.

package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type DirWatcher struct {
	watcher     *fsnotify.Watcher
	fileCreated chan<- string
}

func NewDirWatcher(fileChanges chan<- string) (*DirWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	dw := DirWatcher{
		watcher:     w,
		fileCreated: fileChanges,
	}

	dw.run()
	return &dw, nil
}

func (watcher *DirWatcher) run() {
	go func() {
		for {
			select {
			case event, ok := <-watcher.watcher.Events:
				if !ok {
					return
				}
				switch {
				case event.Op.Has(fsnotify.Rename):
					watcher.fileCreated <- event.Name
				case event.Op.Has(fsnotify.Create):
					watcher.fileCreated <- event.Name
				}

			case err, ok := <-watcher.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
}

func (watcher *DirWatcher) Close() {
	watcher.watcher.Close()
}
