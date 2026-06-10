// Package cache defines the cache backend interface and shared helpers for the
// framework's caching layer.
//
// It declares the Cache interface implemented by concrete drivers such as
// badgerdriver and redisdriver, the Entry type for cached values, JSON
// encode/decode helpers for portable storage, and environment-based detection
// of which backend is configured.
package cache

import (
	"encoding/json"
	"os"
)

// Checks if any framework service is configured to use Badger.
// Returns true if the CACHE environment variable
func UsesBadger() bool {
	return os.Getenv("CACHE") == "badger"
}

// Checks if any framework service is configured to use Redis.
// Returns true if any of the CACHE, SESSION_TYPE, or QUEUE_TYPE environment variable
func UsesRedis() bool {
	redisServices := []string{"CACHE", "SESSION_TYPE", "QUEUE_TYPE"}
	for _, service := range redisServices {
		if os.Getenv(service) == "redis" {
			return true
		}
	}
	return false
}

// Encode serializes a cache Entry (map[string]interface{}) into JSON bytes for storage.
// This function converts cache entries into a portable format that can be stored
// in various cache backends (Redis, Badger, etc.) and survives application restarts.
// Returns the JSON-encoded byte slice or an error if marshaling fails.
func Encode(item Entry) ([]byte, error) {
	return json.Marshal(item)
}

// Decode deserializes JSON bytes back into a cache Entry for use by the application.
// This function reconstructs cache entries from their stored JSON representation,
// allowing cached data to be retrieved and used regardless of the underlying storage backend.
// Returns the decoded Entry or an error if the data is not valid JSON or cannot be unmarshaled.
func Decode(data []byte) (Entry, error) {
	var item Entry
	err := json.Unmarshal(data, &item)
	return item, err
}
