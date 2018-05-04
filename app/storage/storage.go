package storage

import (
	"fmt"

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

func (s Storage) LoadUser(login string) (User, error) {
	u := &User{}
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return fmt.Errorf("bucket users not found")
		}

		val := bucket.Get([]byte(login))
		return u.Unmarshal(val)
	})

	return *u, err
}

func (s Storage) CreateUser(u User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		if b == nil {
			return fmt.Errorf("bucket users not found")
		}

		id, _ := b.NextSequence()
		u.ID = int(id)
		val, err := u.Marshal()
		if err != nil {
			return err
		}

		return b.Put([]byte(u.Login), val)
	})
}
