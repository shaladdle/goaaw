package util

import (
	"crypto/rand"
    "log"
	"io"
	"os"
	"path"
	"strconv"
)

var (
	testPath = "/tmp"
)

func GetTestPath() string {
    return testPath
}

func SetTestPath(tpath string) {
    testPath = tpath
}

func MkTestDir(dir string) error {
    return TryMkdir(path.Join(testPath, dir))
}

func TryMkdir(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return nil
}

const (
	KB = int64(1024)
	MB = KB * KB
	GB = MB * KB
)

type FileSpec struct {
	Num  int   // how many to generate
	Size int64 // size in bytes of the file
}

func GenRandFiles(specs []FileSpec) ([]string, error) {
    log.Println("UTIL", "generating test files")
    ret := []string{}

	for _, spec := range specs {
		for i := 0; i < spec.Num; i++ {
            fname := strconv.Itoa(i)+strconv.FormatInt(spec.Size, 10)
			fpath := path.Join(testPath, fname)
			if err := GenRandFile(fpath, spec.Size); err != nil {
				return nil, err
			}

			ret = append(ret, fname)
		}
	}
    log.Println("UTIL", "done")

	return ret, nil
}

func GenRandFile(dstPath string, size int64) error {
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	_, err = io.CopyN(f, rand.Reader, size)
	if err != nil {
		return err
	}

	return nil
}
