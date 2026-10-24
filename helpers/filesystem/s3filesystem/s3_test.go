package s3filesystem

import (
	"reflect"
	"testing"
)

var disk = S3{
	Endpoint: "s3.amazonaws.com",
	Bucket:   "adele",
}

func TestFilesystem_S3_GetCredentials(t *testing.T) {
	creds := disk.getCredentials()
	if reflect.TypeOf(creds).String() != "*credentials.Credentials" {
		t.Error("filesystem was not able to get basic credentials")
	}
}

func TestFilesystem_S3_Put(t *testing.T) {
	name := "adele.text"
	folder := "./"

	err := disk.Put(name, folder)

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_S3_list(t *testing.T) {
	_, err := disk.List("adele")

	if err == nil {
		t.Error("put object did not return an error when it was expected")
	}
}

func TestFilesystem_S3_get(t *testing.T) {
	err := disk.Get("adele")

	if err != nil {
		t.Error("get objects returned an error when it was not expected")
	}
}

func TestFilesystem_S3_delete(t *testing.T) {
	deleted := disk.Delete([]string{"adele"})

	if deleted != false {
		t.Error("get objects returned an error when it was not expected")
	}
}
