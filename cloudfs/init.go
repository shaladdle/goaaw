package cloudfs

import (
	"fmt"
	"os"
	"path"

	"github.com/shaladdle/goaaw/cloudfs/metastore"

	_ "github.com/mattn/go-sqlite3"
)

const mkdirPerm = 0755

// checkLocal does a preliminary check to see if all local resources appear to
// be at the specified path.
func checkLocal(dpath string) error {
	if info, err := os.Stat(dpath); err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%v exists, but is not a directory")
	}

	if err := metastore.CheckMetaDB(dpath); err != nil {
		return err
	}

	return nil
}

func initLocal(dpath string) error {
	if err := checkLocal(dpath); err == nil {
		return nil
	}

	if err := os.MkdirAll(dpath, mkdirPerm); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(dpath, metaName), mkdirPerm); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(dpath, stagingName), mkdirPerm); err != nil {
		return err
	}

	if _, err := metastore.CreateMetaDB(dpath); err != nil {
		return err
	}

	return nil
}
