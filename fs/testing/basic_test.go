package testing

import (
	"os"
	"path"
	"testing"

	"github.com/shaladdle/goaaw/fs"
	"github.com/shaladdle/goaaw/fs/std"
)

var tests = []struct {
	name string
	fs   func() (fs.FileSystem, func())
}{
	{"std", func() (fs.FileSystem, func()) {
		dir := path.Join(os.TempDir(), "std-fs-test")
		cleanup := func() {
			os.RemoveAll(dir)
		}

		return std.New(dir), cleanup
	}},
}

// TestCreate creates a new file, closes it, and then reads it back to make
// sure it was properly persisted.
func TestCreate(t *testing.T) {
}
