// Package filesystem defines the common FS interface and Listing type shared by
// Adele's remote storage backends (S3, MinIO, SFTP, and WebDAV).
//
// Each backend implements FS to put, get, list, and delete files using a uniform
// API, decoupling application code from any specific storage provider.
package filesystem

import "time"

// The interface for the filesystem that must be implemented
type FS interface {
	Put(fileName string, folder string, acl ...string) error
	Get(destination string, items ...string) error
	List(prefix string) ([]Listing, error)
	Delete(itemsToDelete []string) bool
}

// Describes one file on a remote file system
type Listing struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}
