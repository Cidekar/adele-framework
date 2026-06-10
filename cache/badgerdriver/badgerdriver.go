// Package badgerdriver provides a Badger-backed embedded implementation of the framework's cache.Cache interface (Has/Get/Set/Forget/EmptyByMatch/Empty).
package badgerdriver

import (
	"time"

	"github.com/cidekar/adele-framework/cache"
	"github.com/dgraph-io/badger/v3"
)

// BadgerCache is a cache.Cache implementation backed by an embedded Badger key-value store.
// Conn holds the open *badger.DB connection used for all read and write transactions,
// and Prefix is an optional namespace applied to keys managed by this cache.
type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

// Has reports whether a key exists in the cache by attempting to Get it.
func (b *BadgerCache) Has(str string) (bool, error) {
	_, err := b.Get(str)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Get reads the raw bytes stored for a key inside a Badger read transaction (View) and decodes the stored entry.
// The decoded entry is a map keyed by the cache key, from which the value for str is extracted and returned.
// Returns an error if the key is missing or if reading or decoding the stored bytes fails.
func (b *BadgerCache) Get(str string) (interface{}, error) {
	var fromCache []byte

	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			fromCache = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	decoded, err := cache.Decode(fromCache)
	if err != nil {
		return nil, err
	}

	item := decoded[str]

	return item, nil
}

// Set encodes a value and stores it under the given key inside a Badger write transaction (Update).
// The variadic expires argument is a TTL in seconds applied via WithTTL when supplied; otherwise the entry has no expiry.
// Returns an error if encoding the value fails.
func (b *BadgerCache) Set(str string, value interface{}, expires ...int) error {
	entry := cache.Entry{}

	entry[str] = value
	encoded, err := cache.Encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded).WithTTL(time.Second * time.Duration(expires[0]))
			err = txn.SetEntry(e)
			return err
		})
	} else {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded)
			err = txn.SetEntry(e)
			return err
		})
	}

	return nil
}

// Forget deletes a single key from the cache inside a Badger write transaction (Update).
func (b *BadgerCache) Forget(str string) error {
	err := b.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(str))
		return err
	})

	return err
}

// EmptyByMatch deletes all keys with the given prefix, delegating to emptyByMatch.
func (b *BadgerCache) EmptyByMatch(str string) error {
	return b.emptyByMatch(str)
}

// Empty deletes all keys in the cache, delegating to emptyByMatch with an empty prefix that matches everything.
func (b *BadgerCache) Empty() error {
	return b.emptyByMatch("")
}

// emptyByMatch iterates over all keys matching the given prefix and deletes them.
// Keys are collected and removed in batches of collectSize to bound memory usage and the
// size of each write transaction. Returns an error if iteration or any batch deletion fails.
func (b *BadgerCache) emptyByMatch(str string) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := b.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000

	err := b.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek([]byte(str)); it.ValidForPrefix([]byte(str)); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// CreateBadgerPool opens (or creates) a Badger database at the given storage path with logging disabled.
// Returns the opened *badger.DB, or nil if the database cannot be opened.
func CreateBadgerPool(storagePath string) *badger.DB {
	pool, err := badger.Open(badger.DefaultOptions(storagePath).WithLogger(nil))
	if err != nil {
		return nil
	}
	return pool
}

// BadgerCacheClean runs Badger's value-log garbage collection to reclaim disk space from deleted data, keeping the database size under control.
func BadgerCacheClean(cache *BadgerCache) error {
	return cache.Conn.RunValueLogGC(0.7)
}
