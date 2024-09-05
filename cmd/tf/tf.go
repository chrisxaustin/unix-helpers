// Tailer behaves similarly to `tail -F`, but with an idle timeout that prints a row of dashes if there is no activity for 5s.
//
// This relies on FileWatcher to publish the names of changed files, and DirWatcher to publish the names of created files.
// File rotation is handled automatically, and the user can start following a file before it exists.

package main

import (
	"bufio"
	"fmt"
	"io"
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
	fileCreated  chan string
	fileUpdated  chan string
	fileWatcher  *FileWatcher
	dirWatcher   *DirWatcher
	idleTimeout  *IdleTimer
}

func NewTailer(timeout time.Duration) (*Tailer, error) {
	fileCreatedChannel := make(chan string)
	fileUpdatedChannel := make(chan string)

	watcher := Tailer{
		watchingDir:  make(map[string]bool),
		watchingFile: make(map[string]bool),
		watchedFiles: make(map[string]*os.File),
		fileCreated:  fileCreatedChannel,
		fileUpdated:  fileUpdatedChannel,
		idleTimeout: NewIdleTimer(timeout, func() {
			fmt.Println("----------------------------------------")
		}),
	}
	var err error
	watcher.fileWatcher, err = NewFileWatcher(fileUpdatedChannel)
	if err != nil {
		return nil, err
	}
	watcher.dirWatcher, err = NewDirWatcher(fileCreatedChannel)
	if err != nil {
		return nil, err
	}

	return &watcher, nil
}

func (watcher *Tailer) openFile(name string, seek bool) bool {
	watcher.watchingFile[name] = true
	fh, err := os.Open(name)
	if err != nil {
		return false
	}

	if seek {
		_, err = fh.Seek(0, io.SeekEnd)
		if err != nil {
			return false
		}
	}
	existing_fh := watcher.watchedFiles[name]
	if existing_fh != nil {
		existing_fh.Close()
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
	watcher.fileWatcher.watcher.Add(name)
	parent := filepath.Dir(name)
	if !watcher.watchingDir[parent] {
		watcher.watchingDir[parent] = true
		watcher.dirWatcher.watcher.Add(parent)
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
	watcher.fileWatcher.Close()
	watcher.dirWatcher.Close()
}

func (watcher *Tailer) run() {
	go func() {
		for {
			select {
			case filename, ok := <-watcher.fileUpdated:
				if !ok {
					return
				}
				watched := watcher.watchedFiles[filename]
				watcher.tail(watched)
			case filename, ok := <-watcher.fileCreated:
				if !ok {
					return
				}
				if watcher.watchingFile[filename] {
					watcher.addFile(filename, false)
				}
			}
		}
	}()

}
