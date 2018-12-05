package watch_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus999/go-testpredicate"
	"github.com/marcus999/go-testpredicate/pred"
)

type fsTestEnv struct {
	t        *testing.T
	basePath string
}

func newFsTestEnv(t *testing.T) *fsTestEnv {
	t.Helper()
	basePath, err := ioutil.TempDir("", "go-test-")
	if err != nil {
		t.Errorf("failed to create base directory, %v", err)
	}

	e := &fsTestEnv{
		t:        t,
		basePath: basePath,
	}

	return e
}

func (e *fsTestEnv) teardown() {
	e.t.Helper()
	err := os.RemoveAll(e.basePath)
	if err != nil {
		e.t.Errorf("failed to teardown base directory, %v", err)
	}
}

func (e *fsTestEnv) getBasePath() string {
	return e.basePath
}

func (e *fsTestEnv) expandFilename(filename string) string {
	if strings.HasPrefix(filename, e.basePath) {
		return filename
	}
	return filepath.Join(e.basePath, filename)
}

func (e *fsTestEnv) createFile(filename string) {
	e.t.Helper()
	filename = e.expandFilename(filename)
	path := filepath.Dir(filename)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		e.t.Errorf("failed to create folder for '%v', %v", filename, err)
	}

	err = ioutil.WriteFile(filename, []byte{}, 0666)
	if err != nil {
		e.t.Errorf("failed to create file '%v', %v", filename, err)
	}
}

func (e *fsTestEnv) appendToFile(filename string, content []byte) {
	e.t.Helper()
	filename = e.expandFilename(filename)
	path := filepath.Dir(filename)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		e.t.Errorf("failed to create folder for '%v', %v", filename, err)
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		e.t.Errorf("failed to open file, %v", err)
	}

	if _, err := f.Write(content); err != nil {
		e.t.Errorf("failed to write to file, %v", err)
	}
	if err := f.Close(); err != nil {
		e.t.Errorf("failed to close file, %v", err)
	}
}

func (e *fsTestEnv) mkDir(path string) {
	e.t.Helper()

	path = e.expandFilename(path)
	err := os.MkdirAll(path, 0777)
	if err != nil {
		e.t.Errorf("failed to create folder '%v', %v", path, err)
	}
}

func (e *fsTestEnv) move(from, to string) {
	e.t.Helper()
	from = e.expandFilename(from)
	to = e.expandFilename(to)
	err := os.Rename(from, to)
	if err != nil {
		e.t.Errorf("failed to remove '%v' to '%v', %v", from, to, err)
	}
}

func (e *fsTestEnv) delete(path string) {
	e.t.Helper()
	path = e.expandFilename(path)
	err := os.RemoveAll(path)
	if err != nil {
		e.t.Errorf("failed to remove '%v', %v", path, err)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestFSTestEnv(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	e := newFsTestEnv(t)
	assert.That(e.getBasePath(), pred.Matches(`go-test-\d+`))

	n := e.expandFilename("/a/c.s")
	n2 := e.expandFilename(n)

	assert.That(n, pred.Matches(`go-test-\d+/a/c.s`))
	assert.That(n, pred.IsEqualTo(n2))

	e.mkDir("aaa/bbb")
	e.createFile("aaa/bbb/ccc.yaml")

	e.createFile("aaa/bbc/ccc.yaml")
	e.move("aaa/bbc", "aaa/bcc")
	e.move("aaa/bcc/ccc.yaml", "aaa/bcc/ddd.yaml")

	e.appendToFile("aaa/bcc/ddd.yaml", []byte("aaa\n"))
	e.appendToFile("aaa/bcc/ddd.yaml", []byte("bbb\n"))
	e.appendToFile("aaa/bcc/ddd.yaml", []byte("ccc\n"))

	e.teardown()
}
