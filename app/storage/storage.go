package storage

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type Serializable interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Storage struct {
	db *bolt.DB
}

var Buckets = []string{"users", "converters", "stats", "history", "jobs"}

func New(dbFile string) (*Storage, error) {
	var s Storage
	db, err := bolt.Open(dbFile, 0644, nil)
	if err != nil {
		return nil, err
	}

	s.db = db

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range Buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	})

	return &s, err
}

func (s Storage) Load(b, key string, v Serializable) error {
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", b)
		}

		val := bucket.Get([]byte(key))
		return v.Unmarshal(val)
	})

	return err
}

func (s Storage) Save(b, key string, v Serializable) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(b))
		if b == nil {
			return fmt.Errorf("bucket %s not found", b)
		}

		val, err := v.Marshal()
		if err != nil {
			return err
		}

		return b.Put([]byte(key), val)
	})
}
