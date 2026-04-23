package redisdriver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cidekar/adele-framework/cache"
	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

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
