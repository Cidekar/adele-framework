package webdavfilesystem

import (
	"reflect"
	"testing"
)

var disk = WebDAV{
	Host:     "localhost",
	User:     "test",
	Password: "test",
}

func TestFilesystem_WebDAV_GetCredentials(t *testing.T) {
	creds := disk.getCredentials()
	if reflect.TypeOf(creds).String() != "*gowebdav.Client" {
		t.Error("filesystem was not able to get basic credentials")
	}
}

func TestFilesystem_WebDAV_Put(t *testing.T) {
	name := "adele.text"
	folder := "./"

	err := disk.Put(name, folder)

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_WebDAV_list(t *testing.T) {
	_, err := disk.List("adele")

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_WebDAV_get(t *testing.T) {
	err := disk.Get("adele")

	if err != nil {
		t.Error("get objects returned an error when it was not expected")
	}
}

func TestFilesystem_WebDAV_delete(t *testing.T) {
	deleted := disk.Delete([]string{"adele"})

	if deleted != false {
		t.Error("get objects returned an error when it was not expected")
	}
}
