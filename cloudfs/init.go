package cloudfs

import (
	"github.com/shaladdle/goaaw/cloudfs/metastore"

	_ "github.com/mattn/go-sqlite3"
)

// checkLocal does a preliminary check to see if all local resources appear to
// be at the specified path.
func checkLocal(dpath string) error {
	if err := metastore.CheckMetaDB(dpath); err != nil {
		return err
	}

	return nil
}

func (fs *cloudFileSystem) initLocal(dpath string) error {
	if err := checkLocal(dpath); err == nil {
		return nil
	}

	if _, err := metastore.CreateMetaDB(dpath); err != nil {
		return err
	}

	return nil
}
