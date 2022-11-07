package cache

import (
	"github.com/dgraph-io/badger/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"time"
)

type Cache struct {
	db *badger.DB
}

var (
	cache = new(Cache)
	json  = jsoniter.ConfigCompatibleWithStandardLibrary
)

func New() error {
	opt := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(logrus.StandardLogger()).
		WithLoggingLevel(badger.INFO)
	db, err := badger.Open(opt)
	if err != nil {
		return err
	}
	cache.db = db
	return nil
}

func (c *Cache) Add(key string, value interface{}, TTL time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Update(func(txn *badger.Txn) error {
		if TTL != 0 {
			return txn.SetEntry(badger.NewEntry([]byte(key), data).WithTTL(TTL))
		}
		return txn.Set([]byte(key), data)
	})
}

func (c *Cache) Get(key string, value interface{}) error {
	item, err := c.get(key)
	if err != nil {
		return err
	}
	val, err := c.getItemValue(item)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, value)
}

func (c *Cache) get(key string) (*badger.Item, error) {
	var item = new(badger.Item)
	err := c.db.View(func(txn *badger.Txn) (err error) {
		item, err = txn.Get([]byte(key))
		return err
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (c *Cache) getItemValue(item *badger.Item) (val []byte, err error) {
	var v []byte
	err = item.Value(func(val []byte) error {
		v = append(v, val...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return v, err
}

func (c *Cache) Has(key string) bool {
	_, err := c.get(key)
	if err != nil {
		return false
	}
	return true
}

func (c *Cache) Delete(key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (c *Cache) Iterator(prefix string) map[string][]byte {
	var res = make(map[string][]byte)
	_ = c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				res[string(item.Key())] = v
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return res
}
