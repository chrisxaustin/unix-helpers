package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type FileWatcher struct {
	watcher     *fsnotify.Watcher
	fileUpdated chan string
}

func NewFileWatcher(fileUpdated chan string) *FileWatcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}

	fw := FileWatcher{
		watcher:     w,
		fileUpdated: fileUpdated,
	}
	fw.run()
	return &fw
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
				case event.Op == fsnotify.Chmod:
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
