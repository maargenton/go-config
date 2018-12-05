/*
Package watch provides watch events for a filesystem location.

Internally, it is using fsnotify, but provides a very different functionality.
fsnotify is a platform independent abstraction of filesystem event, and watches
filesystem items themselves (i.e. inode). Watch monitors a filesystem location
(i.e path or filename), wether a filesystem item exists at this location or not.
Key differences include:

fsnotify will return an error if the target does not exist. watch doesn't and
will send a notification when a file is first created at this location.

An fsnotify object watching 'path/to/file.txt' will not fire if 'path/to' is
renamed to 'path/not_to'. watch will detect that change and signal that the
file has been deleted, as it is no longer present at the watched location

FileWatcher objects should be created with etiher watch.New() or watch.NewCtx().

*/
package watch

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// EventType represent the type of file watch event
type EventType int

const (
	// Created is the event type sent when the watched location is created
	Created EventType = iota + 1

	// Updated is the event type sent when the watched location is changed
	Updated

	// Deleted is the event type sent when the watched location is removed
	Deleted
)

var eventTypes = []string{
	"Invalid",
	"Created",
	"Updated",
	"Deleted",
}

func (e EventType) String() string {
	return eventTypes[int(e)]
}

// FileWatcher watches a single filesystem location and notifies xxx when
// a file at that location is created, updated or deleted
type FileWatcher struct {
	filename string
	fileInfo os.FileInfo
	watcher  *fsnotify.Watcher

	updateCh chan EventType
	ctx      context.Context
	cancel   func()
}

// NewFileWatcher creates a new FileWatcher
func NewFileWatcher(filename string) (*FileWatcher, error) {
	return NewFileWatcherWithContext(context.Background(), filename)
}

// NewFileWatcherWithContext creates a new FileWatcher with an explicit
// cancelation context
func NewFileWatcherWithContext(ctx context.Context, filename string) (*FileWatcher, error) {
	target, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)

	var w = &FileWatcher{
		filename: target,
		updateCh: make(chan EventType, 1),
		ctx:      ctx,
		cancel:   cancel,
	}

	n, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w.watcher = n

	info, _ := os.Stat(filename)
	if info != nil && !info.IsDir() {
		w.fileInfo = info
	}

	go w.run()

	return w, nil
}

// Info retuens the FileInfo of the watched file, or nil if there is not file
// at the watched location
func (w *FileWatcher) Info() os.FileInfo {
	return w.fileInfo
}

// UpdateChannel returns the readabl channel on which updates are sent
func (w *FileWatcher) UpdateChannel() <-chan EventType {
	return w.updateCh
}

// Close closes the watcher and releases associated resources
func (w *FileWatcher) Close() {
	w.cancel()
}

func (w *FileWatcher) run() {
	for {
		path, target := watchLocation(w.filename)
		targetStat, _ := os.Stat(target)

		err := w.watcher.Add(path)
		if err != nil {
			continue
		}
		w.watchParents(path)

	watchloop:
		for {
			select {
			case ev := <-w.watcher.Events:
				if (ev.Op & fsnotify.Remove) != 0 {
					w.handleDeleteEvent(&ev)
					break watchloop
				} else if (ev.Op & fsnotify.Create) != 0 {
					w.handleCreateEvent(&ev)
					if target != w.filename {
						break watchloop
					}
				} else {
					evTargetStat, _ := os.Stat(ev.Name)
					if os.SameFile(targetStat, evTargetStat) {
						if target != w.filename {
							break watchloop
						} else {
							w.handleEvent(&ev)
						}
					}
				}

			case <-w.watcher.Errors:
				break watchloop

			case <-w.ctx.Done():
				close(w.updateCh)
				w.watcher.Close()
				return
			}
		}

		w.watcher.Remove(path)
	}
}

func (w *FileWatcher) watchParents(path string) {
	for {
		next := filepath.Dir(path)
		if next == path {
			break
		}
		path = next
		w.watcher.Add(path)
	}
}

func (w *FileWatcher) handleEvent(ev *fsnotify.Event) {
	log.Printf("watch: %v", ev)
	w.fileInfo, _ = os.Stat(w.filename)
	w.updateCh <- Updated
}

func (w *FileWatcher) handleCreateEvent(ev *fsnotify.Event) {
	log.Printf("watch: %v", ev)
	newFileInfo, _ := os.Stat(w.filename)
	if newFileInfo != nil && w.fileInfo == nil {
		w.fileInfo = newFileInfo
		w.updateCh <- Created
	}
}

func (w *FileWatcher) handleDeleteEvent(ev *fsnotify.Event) {
	log.Printf("watch: %v", ev)
	newFileInfo, _ := os.Stat(w.filename)
	if newFileInfo == nil && w.fileInfo != nil {
		w.fileInfo = nil
		w.updateCh <- Deleted
	}
}

func watchLocation(path string) (watchPath, watchTarget string) {
	watchPath = path
	watchTarget = path
	for {
		if info, err := os.Stat(watchPath); err == nil && info.IsDir() {
			return
		}
		watchTarget = watchPath
		watchPath = filepath.Dir(watchPath)
	}
}
