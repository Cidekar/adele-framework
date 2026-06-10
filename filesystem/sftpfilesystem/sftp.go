// Package sftpfilesystem implements the filesystem.FS interface backed by a
// remote SFTP server over SSH using the pkg/sftp client.
package sftpfilesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/cidekar/adele-framework/filesystem"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTP struct {
	Host     string
	User     string
	Password string
	Port     string
}

func (s *SFTP) getCredentials() (*sftp.Client, error) {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	cwd, err := client.Getwd()
	log.Println("Current working directory:", cwd)

	return client, nil

}

func (s *SFTP) Put(filename string, folder string, acl ...string) error {
	// Client
	client, err := s.getCredentials()
	if err != nil {
		return err
	}

	defer client.Close()

	// File
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f2, err := client.Create(fmt.Sprintf("%s/%s", folder, path.Base(filename)))
	if err != nil {
		return err
	}
	defer f2.Close()

	// Move the file
	if _, err := io.Copy(f2, f); err != nil {
		return err
	}

	return nil
}

func (s *SFTP) List(prefix string) ([]filesystem.Listing, error) {
	var listing []filesystem.Listing
	// Client
	client, err := s.getCredentials()
	if err != nil {
		return listing, err
	}

	defer client.Close()

	// Files
	files, err := client.ReadDir(prefix)
	if err != nil {
		return listing, err
	}
	for _, x := range files {
		var item filesystem.Listing

		if !strings.HasPrefix(x.Name(), ".") {
			b := float64(x.Size())
			kb := b / 1024
			mb := kb / 1024
			item.Key = x.Name()
			item.Size = mb
			item.LastModified = x.ModTime()
			item.IsDir = x.IsDir()
			listing = append(listing, item)
		}
	}
	return listing, nil
}

func (s *SFTP) Delete(itemsToDelete []string) bool {
	// Client
	client, err := s.getCredentials()
	if err != nil {
		return false
	}

	defer client.Close()

	// Range through the files and delete them
	for _, x := range itemsToDelete {
		deleteErr := client.Remove(x)
		if deleteErr != nil {
			return false
		}
	}

	return true
}

func (s *SFTP) Get(destination string, items ...string) error {
	// Client
	client, err := s.getCredentials()
	if err != nil {
		return err
	}

	defer client.Close()

	for _, item := range items {
		err := func() error {
			// create a dest file
			dstFile, err := os.Create(fmt.Sprintf("%s/%s", destination, path.Base(item)))
			if err != nil {
				return err
			}
			defer dstFile.Close()

			// open source
			srcFile, err := client.Open(item)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			// copy src to dst
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// flush in-memory
			err = dstFile.Sync()
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
