package miniofilesystem

import (
	"reflect"
	"testing"
)

var disk = Minio{
	Endpoint: "s3.amazonaws.com",
	UseSSL:   true,
	Bucket:   "adele",
}

func TestFilesystem_minio_GetCredentials(t *testing.T) {
	creds := disk.getCredentials()
	if reflect.TypeOf(creds).String() != "*minio.Client" {
		t.Error("filesystem was not able to get basic credentials")
	}
}

func TestFilesystem_minio_Put(t *testing.T) {
	name := "adele.text"
	folder := "./"

	err := disk.Put(name, folder)

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_minio_list(t *testing.T) {
	_, err := disk.List("adele")

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_minio_get(t *testing.T) {
	err := disk.Get("adele")

	if err != nil {
		t.Error("get objects returned an error when it was not expected")
	}
}

func TestFilesystem_minio_delete(t *testing.T) {
	deleted := disk.Delete([]string{"adele"})

	if deleted != false {
		t.Error("get objects returned an error when it was not expected")
	}
}
