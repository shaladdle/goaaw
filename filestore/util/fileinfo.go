package util

import (
	"encoding/gob"
	"os"
	"time"
)

func init() {
	gob.Register(FileInfo{})
	gob.Register([]FileInfo{})
}

type FileInfo struct {
	I_Name    string      // base name of the file
	I_Size    int64       // length in bytes for regular files; system-dependent for others
	I_Mode    os.FileMode // file mode bits
	I_ModTime time.Time   // modification time
	I_IsDir   bool        // abbreviation for Mode().IsDir()
}

func (info FileInfo) Name() string       { return info.I_Name }
func (info FileInfo) Size() int64        { return info.I_Size }
func (info FileInfo) Mode() os.FileMode  { return info.I_Mode }
func (info FileInfo) ModTime() time.Time { return info.I_ModTime }
func (info FileInfo) IsDir() bool        { return info.I_IsDir }
func (info FileInfo) Sys() interface{}   { return nil }

func FromOSInfo(info os.FileInfo) FileInfo {
	return FileInfo{
		I_Name:    info.Name(),
		I_Size:    info.Size(),
		I_Mode:    info.Mode(),
		I_ModTime: info.ModTime(),
		I_IsDir:   info.IsDir(),
	}
}
