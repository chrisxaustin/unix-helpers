// FileWatcher uses fsnotify to listen for changes to any of the monitored files.
//
// When a Write event is observed it will send the name of the file to the fileUpdated channel.

package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type FileWatcher struct {
	watcher     *fsnotify.Watcher
	fileUpdated chan string
}

func NewFileWatcher(fileUpdated chan string) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := FileWatcher{
		watcher:     w,
		fileUpdated: fileUpdated,
	}
	fw.run()
	return &fw, nil
}

func (watcher *FileWatcher) run() {
	go func() {
		for {
			select {
			case event, ok := <-watcher.watcher.Events:
				if !ok {
					return
				}
				switch {
				case event.Op.Has(fsnotify.Write):
					watcher.fileUpdated <- event.Name
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

func (watcher *FileWatcher) Close() {
	watcher.watcher.Close()
}
