package main

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
)

/* Storage implentation of the Storage interface using sqlite3 as the backend. */
type Storage interface {
	Get(key string) ([]byte, error)
	Put(key string, val []byte) error
	Del(key string) error
}

// SqliteStorage is a storage that uses sqlite as backend
type SqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage(dbPath string) (Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS storage (key TEXT PRIMARY KEY, val TEXT)")
	if err != nil {
		return nil, err
	}
	return &SqliteStorage{
		db: db,
	}, nil
}

func (s *SqliteStorage) Get(key string) ([]byte, error) {
	var val string
	err := s.db.QueryRow("SELECT val FROM storage WHERE key = ?", key).Scan(&val)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func (s *SqliteStorage) Put(key string, val []byte) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO storage (key, val) VALUES (?, ?)", key, string(val))
	return err
}

func (s *SqliteStorage) Del(key string) error {
	_, err := s.db.Exec("DELETE FROM storage WHERE key = ?", key)
	return err
}

var _globalStorage Storage

func GetStorage() Storage {
	return _globalStorage
}
