package cloudfs

import (
	"database/sql"
	"fmt"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

const sqliteDBName = "cloudfs.sqlite"

type metaStore interface {
	IsBig(key string) (bool, error)
	SetBig(key string, value bool) error
	Remove(key string) error
}

type memMetaStore map[string]bool

func newMemMetaStore() metaStore {
	return memMetaStore(make(map[string]bool))
}

func (m memMetaStore) IsBig(key string) (bool, error) {
	ret, exists := m[key]
	if !exists {
		return false, fmt.Errorf("key %v does not exist", key)
	}

	return ret, nil
}

func (m memMetaStore) SetBig(key string, value bool) error {
	m[key] = value
	return nil
}

func (m memMetaStore) Remove(key string) error {
	delete(m, key)
	return nil
}

type metadb struct {
	db *sql.DB
}

func createMetaDb(dpath string) (string, error) {
	fpath := path.Join(dpath, sqliteDBName)
	db, err := sql.Open("sqlite3", fpath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec("create table isbig (fkey char(40) primary key, big boolean)")
	return fpath, err
}

func newMetaDB(fpath string) (metaStore, error) {
	db, err := sql.Open("sqlite3", fpath)
	if err != nil {
		return nil, err
	}

	return &metadb{db}, nil
}

func (m *metadb) IsBig(key string) (bool, error) {
	rows, err := m.db.Query(fmt.Sprintf("select big from isbig where fkey='%v'", key))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// If next returns false, then there is no entry for this file in the
	// database, so return an error
	if !rows.Next() {
		return false, fmt.Errorf("no entry for file in database")
	}

	var ret bool
	if err = rows.Scan(&ret); err != nil {
		return false, err
	}

	return ret, nil
}

func (m *metadb) SetBig(key string, value bool) error {
	rows, err := m.db.Query(fmt.Sprintf("select big from isbig where fkey='%v'", key))
	if err != nil {
		return err
	}
	defer rows.Close()

	var val int
	if value {
		val = 1
	} else {
		val = 0
	}

	if rows.Next() {
		_, err := m.db.Exec(fmt.Sprintf("update isbig set big=%v where fkey='%v'", val, key))
		if err != nil {
			return err
		}
	} else {
		_, err := m.db.Exec(fmt.Sprintf("insert into isbig (fkey, big) values ('%v', %v)", key, val))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *metadb) Remove(key string) error {
	if _, err := m.db.Exec("delete from isbig where fkey=?", key); err != nil {
		return err
	}

	return nil
}

func (m *metadb) Close() error {
	return m.db.Close()
}
