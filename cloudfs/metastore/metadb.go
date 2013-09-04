package metastore

import (
	"database/sql"
	"fmt"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

const isbigTableName = "isbig"

const sqliteDBName = "cloudfs.sqlite"

type MetaStore interface {
	IsBig(key string) (bool, error)
	SetBig(key string, value bool) error
	Remove(key string) error
	Close() error
}

type inMem map[string]bool

func NewInMem() MetaStore {
	return inMem(make(map[string]bool))
}

func (m inMem) IsBig(key string) (bool, error) {
	ret, exists := m[key]
	if !exists {
		return false, fmt.Errorf("key %v does not exist", key)
	}

	return ret, nil
}

func (m inMem) SetBig(key string, value bool) error {
	m[key] = value
	return nil
}

func (m inMem) Remove(key string) error {
	delete(m, key)
	return nil
}

func (inMem) Close() error {
	return nil
}

type metadb struct {
	db *sql.DB
}

func CheckMetaDB(dpath string) error {
	db, err := sql.Open("sqlite3", getDBPath(dpath))
	if err != nil {
		return err
	}
	defer db.Close()

	var tableName string
	row := db.QueryRow("select name from sqlite_master where type='table'")
	if err := row.Scan(&tableName); err != nil {
		return err
	}

	if tableName != isbigTableName {
		return fmt.Errorf("table %v unexpected, wanted table %v", tableName, isbigTableName)
	}

	return nil
}

func getDBPath(dpath string) string {
	return path.Join(dpath, sqliteDBName)
}

func CreateMetaDB(dpath string) (string, error) {
	fpath := getDBPath(dpath)
	db, err := sql.Open("sqlite3", fpath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("create table %v (fkey char(40) primary key, big boolean)",
		isbigTableName))

	return fpath, err
}

func NewMetaDB(fpath string) (MetaStore, error) {
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
