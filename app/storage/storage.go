package storage

import (
	"github.com/boltdb/bolt"
)

type Storage struct {
	db *bolt.DB
}

func New(dbFile string) (*Storage, error) {
	var s Storage
	db, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		return nil, err
	}

	s.db = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		return err
	})

	return &s, err
}
