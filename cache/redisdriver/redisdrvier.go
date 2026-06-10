// Package redisdriver provides a Redis-backed implementation of the framework's cache.Cache interface (Has/Get/Set/Forget/EmptyByMatch/Empty).
package redisdriver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cidekar/adele-framework/cache"
	"github.com/gomodule/redigo/redis"
)

// RedisCache is a Redis-backed implementation of the cache.Cache interface.
// Conn is the underlying redis connection pool from which connections are borrowed,
// and Prefix namespaces all keys so multiple caches can share a single Redis instance.
type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

// Has reports whether the prefixed key exists in Redis using the EXISTS command.
func (c *RedisCache) Has(str string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	connection := c.Conn.Get()
	defer connection.Close()

	ok, err := redis.Bool(connection.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Get fetches the cached entry stored under the prefixed key and decodes it.
// The raw bytes are retrieved with GET and run through cache.Decode, which yields
// the entry map keyed by the prefixed key. Returns the stored value or an error if
// the key is missing or decoding fails.
func (c *RedisCache) Get(str string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	fmt.Println("try to get:", key)
	conn := c.Conn.Get()
	defer conn.Close()

	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := cache.Decode(cacheEntry)
	if err != nil {
		return nil, err
	}

	item := decoded[key]

	return item, nil
}

// Set encodes the given value and stores it under the prefixed key.
// The variadic expires argument is a TTL in seconds: when supplied the value is
// written with SETEX so it expires automatically, otherwise a non-expiring SET is used.
// Returns an error if encoding or the Redis command fails.
func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()

	defer conn.Close()

	entry := cache.Entry{}
	entry[key] = value
	encoded, err := cache.Encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		_, err := conn.Do("SETEX", key, expires[0], string(encoded))
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Do("SET", key, string(encoded))
		if err != nil {
			return err
		}
	}

	return nil
}

// Forget deletes the single prefixed key from Redis using the DEL command.
func (c *RedisCache) Forget(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}

// EmptyByMatch deletes all keys matching the given prefixed pattern.
// It collects the matching keys via getKeys (a SCAN cursor walk) and then issues a
// DEL for each. Returns an error if scanning or any deletion fails.
func (c *RedisCache) EmptyByMatch(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		_, err = conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}
	return nil
}

// Empty deletes every key under this cache's prefix.
// It scans for all keys beginning with the prefix via getKeys and issues a DEL for
// each one. Returns an error if scanning or any deletion fails.
func (c *RedisCache) Empty() error {
	key := fmt.Sprintf("%s:", c.Prefix)
	conn := c.Conn.Get()
	defer conn.Close()

	keys, err := c.getKeys(key)
	if err != nil {
		return err
	}

	for _, x := range keys {
		_, err = conn.Do("DEL", x)
		if err != nil {
			return err
		}
	}

	return nil
}

// getKeys performs a cursor-based SCAN to collect every key matching the given pattern.
// It repeatedly issues SCAN with a MATCH of pattern* until the returned cursor wraps back
// to zero, accumulating the keys from each batch. Returns the collected keys or an error.
func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := c.Conn.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}

	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return keys, err
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

// CreateRedisPool builds a redis.Pool from string configuration values.
// The idel, active, and timeout arguments are parsed into MaxIdle, MaxActive, and an
// IdleTimeout (in seconds); the dial function applies optional username/password auth and
// a PING-based TestOnBorrow health check. Returns an error if any numeric config string is invalid.
func CreateRedisPool(idel, active, timeout, host, username, password string) (*redis.Pool, error) {

	maxIdle, err := strconv.Atoi(idel)
	if err != nil {
		return nil, fmt.Errorf("invalid MaxIdle value: %v", err)
	}

	maxActive, err := strconv.Atoi(active)
	if err != nil {
		return nil, fmt.Errorf("invalid MaxActive value: %v", err)
	}

	idleTimeout, err := strconv.Atoi(timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid IdleTimeout value: %v", err)
	}

	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			opts := []redis.DialOption{}
			if username != "" {
				opts = append(opts, redis.DialUsername(username))
			}
			if password != "" {
				opts = append(opts, redis.DialPassword(password))
			}
			return redis.Dial("tcp", host, opts...)
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}, nil
}
