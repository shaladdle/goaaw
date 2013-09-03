// Package fs defines the basic filestore interface.
package fs

import (
	"io"
	"os"
)

// FileSystem defines the basic interface of a file store. This is meant to be
// a simple streaming interface, instead of a full blown file system.
type FileStore interface {
	// Open a file for reading. Errors if the file doesn't exist.
	//
	// Note: Currently this is limited compared to normal file systems. You
	// cannot Seek, ReadAt, Truncate, etc...
	Open(path string) (io.ReadCloser, error)

	// Open a file for writing. Creates a new file if it doesn't exist.
	// Truncates the file if there is already one at this path.
	//
	// Note: Currently this is limited compared to normal file systems. You
	// cannot Seek, ReadAt, Truncate, etc...
	Create(path string) (io.WriteCloser, error)

	// Creates a directory at path. Also creates all parent directories
	// required to make the final directory.
	Mkdir(path string) error

	// Get file info.
	Stat(path string) (os.FileInfo, error)

	// Delete file.
	Remove(path string) error

	// Get a list of all files in the directory at path.
	GetFiles(path string) ([]os.FileInfo, error)
}
