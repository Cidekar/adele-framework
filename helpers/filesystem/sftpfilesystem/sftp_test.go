package sftpfilesystem

import (
	"reflect"
	"testing"
)

var disk = SFTP{
	Host:     "localhost",
	User:     "test",
	Password: "test",
}

func TestFilesystem_Sftp_GetCredentials(t *testing.T) {
	creds, err := disk.getCredentials()
	if reflect.TypeOf(creds).String() != "*sftp.Client" {
		t.Error("filesystem was not able to get basic credentials")
	}

	if err == nil {
		t.Error("filesystem did not return an error when one was expected")
	}
}

func TestFilesystem_Sftp_Put(t *testing.T) {
	name := "adele.text"
	folder := "./"

	err := disk.Put(name, folder)

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_Sftp_list(t *testing.T) {
	_, err := disk.List("adele")

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_Sftp_get(t *testing.T) {
	err := disk.Get("adele")

	if err == nil {
		t.Error("get objects returned an error when it was not expected")
	}
}

func TestFilesystem_Sftp_delete(t *testing.T) {
	deleted := disk.Delete([]string{"adele"})

	if deleted != false {
		t.Error("get objects returned an error when it was not expected")
	}
}
