package vite

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"
)

type ManifestEntry struct {
	File string `json:"file"`
}

var (
	buf             = &bytes.Buffer{}
	Host            = "localhost"
	Port            = "4001"
	RootDir         = "views"
	DistDir         = "dist"
	Manifest        = "manifest.json"
	DevelopmentMode = "false"
	InfoLog         = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog        = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	TestManifest    map[string]ManifestEntry
)

func TestMain(m *testing.M) {
	InfoLog.SetOutput(buf)
	ErrorLog.SetOutput(buf)

	data, err := os.ReadFile("./testdata/manifest.json")

	if err != nil {
		panic("please run vite build and copy the manifest.json into /vite/testdata")
	}

	err = json.Unmarshal(data, &TestManifest)

	if err != nil {
		panic("unable to parse the manifest.json")
	}

	os.Exit(m.Run())
}
