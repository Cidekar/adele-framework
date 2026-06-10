// Package miniofilesystem implements the filesystem.FS interface backed by
// MinIO object storage using the minio-go client.
package miniofilesystem

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/cidekar/adele-framework/filesystem"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Minio struct {
	Endpoint string
	Key      string
	Secret   string
	UseSSL   bool
	Region   string
	Bucket   string
}

func (m *Minio) getCredentials() *minio.Client {
	client, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.Key, m.Secret, ""),
		Secure: m.UseSSL,
	})

	if err != nil {
		log.Println(err)
	}

	return client
}

func (m *Minio) Put(fileName string, folder string, acl ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Strip and get the filename without the path
	objectName := path.Base(fileName)

	client := m.getCredentials()

	// Let transport
	upload, err := client.FPutObject(ctx, m.Bucket, fmt.Sprintf("%s/%s", folder, objectName), fileName, minio.PutObjectOptions{})
	if err != nil {
		log.Println("Failed with FPutObject")
		log.Println(err)
		log.Println("upload info:", upload)
		return err
	}

	return nil
}

func (m *Minio) List(prefix string) ([]filesystem.Listing, error) {
	var listing []filesystem.Listing

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	objectCh := client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {

		if object.Err != nil {
			fmt.Println(object.Err)
			return listing, object.Err
		}

		// Ignore any hidden files, starts with a "."
		if !strings.HasPrefix(object.Key, ".") {
			b := float64(object.Size)
			kb := b / 1024
			mb := kb / 1024
			item := filesystem.Listing{
				Etag:         object.ETag,
				LastModified: object.LastModified,
				Key:          object.Key,
				Size:         mb,
			}

			listing = append(listing, item)

		}
	}

	return listing, nil
}

func (m *Minio) Delete(itemsToDelete []string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	options := minio.RemoveObjectOptions{
		// Delete object regardless of lock state
		GovernanceBypass: true,
	}

	for _, item := range itemsToDelete {
		err := client.RemoveObject(ctx, m.Bucket, item, options)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}

func (m *Minio) Get(destination string, items ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	for _, item := range items {
		err := client.FGetObject(ctx, m.Bucket, item, fmt.Sprintf("%s/%s", destination, path.Base(item)), minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
