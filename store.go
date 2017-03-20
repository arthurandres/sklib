package sklib

import "github.com/boltdb/bolt"

const (
	cacheLocation = "cache.db"
	bucketName    = "cache"
)

type CacheStore interface {
	Get(key string) (data []byte)
	Set(key string, data []byte) error
}

type BoltStore struct {
	DB     *bolt.DB
	Bucket string
}

type WriteOnlyStore struct {
	Store CacheStore
}

func (m *BoltStore) Get(key string) []byte {
	var value []byte
	m.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.Bucket))
		if b == nil {
			return nil
		}
		value = b.Get([]byte(key))
		return nil
	})
	return value
}

func (m *BoltStore) Set(key string, data []byte) error {
	var err error
	m.DB.Update(func(tx *bolt.Tx) error {
		var b *bolt.Bucket
		b, err = tx.CreateBucketIfNotExists([]byte(m.Bucket))
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), data)
		return err
	})
	return err
}

func (m *WriteOnlyStore) Get(key string) []byte {
	return nil
}

func (m *WriteOnlyStore) Set(key string, data []byte) error {
	return m.Store.Set(key, data)
}

func CreateDB() *bolt.DB {
	var db *bolt.DB
	var err error
	db, err = bolt.Open(cacheLocation, 0600, nil)
	if err != nil {
		panic(err)
	}
	return db
}
