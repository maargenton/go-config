package watch_test

import (
	"testing"
	"time"

	"github.com/marcus999/go-config/pkg/watch"

	"github.com/marcus999/go-testpredicate"
	"github.com/marcus999/go-testpredicate/pred"
)

const defaultTimeout = 100 * time.Millisecond

func readChannel(
	ch <-chan watch.EventType, timeout time.Duration) (
	watch.EventType, bool, bool) {

	select {
	case e, ok := <-ch:
		return e, ok, false
	case <-time.After(timeout):
		return watch.EventType(0), false, true
	}
}

func TestWatchModifyingExistingFile(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.createFile(target)

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.appendToFile(target, []byte("aaa\n"))

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Updated), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchDeletingExistingFile(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.createFile(target)

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.delete("path/to/file.yaml")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Deleted), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchDeletingParentOfExistingFile(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.createFile(target)

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.delete("path/to")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Deleted), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchCreateInExistingFolder(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.mkDir("path/to")

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.createFile("path/to/other_file.yaml")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.createFile("path/to/file.yaml")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Created), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchMovingParentFolderIntoPlace(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.createFile("path/not_to/file.yaml")

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.move("path/not_to", "path/to")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Created), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchMovingParentFolderOutOfPlace(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/file.yaml")
	fs.createFile("path/to/file.yaml")

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.move("path/to", "path/not_to")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Deleted), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}

func TestWatchMovingParentFolderOutOfPlace2(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	fs := newFsTestEnv(t)

	target := fs.expandFilename("path/to/intermediate/file.yaml")
	fs.createFile("path/to/intermediate/file.yaml")

	w, err := watch.NewFileWatcher(target)
	assert.That(err, pred.IsNil(), "failed create watcher, %v", err)

	e, ok, timeout := readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(timeout, pred.IsEqualTo(true), "expected timeout, e: %v, ok: %v", e, ok)

	fs.move("path/to", "path/not_to")

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(e, pred.IsEqualTo(watch.Deleted), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	w.Close()

	e, ok, timeout = readChannel(w.UpdateChannel(), defaultTimeout)
	assert.That(ok, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)
	assert.That(timeout, pred.IsEqualTo(false), "e: %v, ok: %v, timeout: %v", e, ok, timeout)

	fs.teardown()
}
