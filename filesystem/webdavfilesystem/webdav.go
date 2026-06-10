// Package webdavfilesystem implements the filesystem.FS interface backed by a
// remote WebDAV server using the gowebdav client.
package webdavfilesystem

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cidekar/adele-framework/filesystem"
	"github.com/studio-b12/gowebdav"
)

type WebDAV struct {
	Host     string
	User     string
	Password string
}

func (s *WebDAV) getCredentials() *gowebdav.Client {
	client := gowebdav.NewClient(s.Host, s.User, s.Password)
	return client
}

func (s *WebDAV) Put(fileName string, folder string, acl ...string) error {
	client := s.getCredentials()
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = client.WriteStream(fmt.Sprintf("%s/%s", folder, path.Base(fileName)), file, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (s *WebDAV) List(prefix string) ([]filesystem.Listing, error) {
	var listing []filesystem.Listing
	client := s.getCredentials()
	files, err := client.ReadDir(prefix)
	if err != nil {
		return listing, err
	}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			b := float64(file.Size())
			kb := b / 1024
			mb := kb / 1024
			current := filesystem.Listing{
				LastModified: file.ModTime(),
				Key:          file.Name(),
				Size:         mb,
				IsDir:        file.IsDir(),
			}
			listing = append(listing, current)
		}
	}

	return listing, nil
}

func (s *WebDAV) Delete(itemsToDelete []string) bool {
	client := s.getCredentials()
	for _, item := range itemsToDelete {
		err := client.Remove(item)
		if err != nil {
			return false
		}
	}
	return true
}

func (s *WebDAV) Get(destination string, items ...string) error {
	client := s.getCredentials()

	for _, item := range items {
		err := func() error {
			webdavFilePath := item
			localFilePath := fmt.Sprintf("%s/%s", destination, path.Base(item))

			// Get a reader
			reader, err := client.ReadStream(webdavFilePath)
			if err != nil {
				return err
			}

			// Create empty file
			file, err := os.Create(localFilePath)
			if err != nil {
				return err
			}
			defer file.Close()

			// DO the copy
			_, err = io.Copy(file, reader)
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
