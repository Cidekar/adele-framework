package vite

import (
	"strings"
	"testing"
)

func TestVite_New(t *testing.T) {

	v := New(Host, Port, RootDir, DistDir, Manifest, DevelopmentMode, InfoLog, ErrorLog)

	if v.Host != Host {
		t.Fatalf("vite symbols do not match %s and %s", v.Host, Host)
	}

	if v.Port != Port {
		t.Fatalf("vite symbols do not match %s and %s", v.Port, Port)
	}

	if v.Manifest != Manifest {
		t.Fatalf("vite symbols do not match %s and %s", v.Manifest, Manifest)
	}

	manifestPath := "./public/" + DistDir + "/" + Manifest
	if v.ManifestPath != manifestPath {
		t.Fatalf("vite symbols do not match %s and %s", v.ManifestPath, manifestPath)
	}

	clientPath := "//" + Host + ":" + Port + "/@vite/client"

	if v.ClientPath != clientPath {
		t.Fatalf("vite symbols do not match %s and %s", v.ClientPath, clientPath)
	}

	devServer := "//" + Host + ":" + Port
	if v.DevServer != devServer {
		t.Fatalf("vite symbols do not match %s and %s", v.DevServer, devServer)
	}

	if v.DistDir != "/public/"+DistDir {
		t.Fatalf("vite symbols do not match %s and %s", v.DistDir, DistDir)
	}

	if v.RootDir != RootDir {
		t.Fatalf("vite symbols do not match %s and %s", v.RootDir, RootDir)
	}
}

func TestVite_Defaults(t *testing.T) {

	v := New("", "", "", "", "", "", InfoLog, ErrorLog)

	if v.Host != "localhost" {
		t.Fatalf("vite symbols do not match %s and %s", v.Host, "localhost")
	}

	if v.Port != "4001" {
		t.Fatalf("vite symbols do not match %s and %s", v.Port, "4001")
	}

	if v.Manifest != "manifest.json" {
		t.Fatalf("vite symbols do not match %s and %s", v.Manifest, "manifest.json")
	}

	if v.ManifestPath != "./public/dist/manifest.json" {
		t.Fatalf("vite symbols do not match %s and %s", v.ManifestPath, "./public/dist/manifest.json")
	}

	if v.ClientPath != "//localhost:4001/@vite/client" {
		t.Fatalf("vite symbols do not match %s and %s", v.ClientPath, "//localhost:4001/@vite/client")
	}

	if v.DevServer != "" {
		t.Fatalf("vite symbols do not match %s and %s", v.DevServer, "")
	}

	if v.DistDir != "/public/dist" {
		t.Fatalf("vite symbols do not match %s and %s", v.DistDir, "/public/dist")
	}

	if v.RootDir != "resources/views" {
		t.Fatalf("vite symbols do not match %s and %s", v.RootDir, "resources/views")
	}
}

func TestVite_JetFastFunction(t *testing.T) {

	v := New(Host, Port, RootDir, DistDir, Manifest, "true", InfoLog, ErrorLog)

	assetPath := []byte("app.css")

	path := v.GetViteAssetPath(assetPath)
	if path.String() != v.DevServer+string(assetPath[:]) {
		t.Fatalf("vite did not return the development server path as expected: %s and %s", path, v.DevServer+string(assetPath[:]))
	}

	v.DevelopmentMode = false

	path = v.GetViteAssetPath(assetPath)
	if path.String() != "" {
		t.Fatalf("vite did not return the asset path as expected: %s and %s", path, v.DevServer+string(assetPath[:]))
	}

	if !strings.Contains(buf.String(), "jet was unable locate the vite manifest in") {
		t.Fatalf("vite did not log the expected error message: \n%s", buf.String())
	}
}

func TestVite_JetFastFunction_Manifest(t *testing.T) {

	buf.Reset()

	v := Vite{
		ClientPath:      "//localhost:4001/resources/views/@vite/client",
		DevelopmentMode: false,
		DevServer:       "//localhost:4001/resources/views/",
		InfoLog:         InfoLog,
		ErrorLog:        ErrorLog,
		Host:            "localhost",
		Manifest:        "manifest.json",
		ManifestPath:    "./testdata/manifest.json",
		Port:            "4001",
		DistDir:         "/public/dist",
		RootDir:         "resources/views",
	}

	assetPath := []byte("css/styles.css")

	path := v.GetViteAssetPath(assetPath)
	assetPathInManifest := v.DistDir
	for k, m := range TestManifest {
		if k == string(assetPath) {
			assetPathInManifest = assetPathInManifest + "/" + m.File
			break
		}
	}

	if assetPathInManifest != path.String() {
		t.Fatalf("vite did not return the asset path from the manifest expected: %s and %s", path.String(), assetPathInManifest)
	}

	assetPath = []byte("css/unknown-styles.css")

	path = v.GetViteAssetPath(assetPath)

	if path.String() != "" {
		t.Fatalf("vite did not return the asset path from the manifest expected: %s and %s", path.String(), assetPathInManifest)
	}
}

func TestVite_JetFastFunction_Manifest_Parse(t *testing.T) {

	buf.Reset()

	v := Vite{
		ClientPath:      "//localhost:4001/resources/views/@vite/client",
		DevelopmentMode: false,
		DevServer:       "//localhost:4001/resources/views/",
		InfoLog:         InfoLog,
		ErrorLog:        ErrorLog,
		Host:            "localhost",
		Manifest:        "manifest_bad.json",
		ManifestPath:    "./testdata/manifest_bad.json",
		Port:            "4001",
		DistDir:         "/public/dist",
		RootDir:         "resources/views",
	}

	assetPath := []byte("css/styles.css")

	path := v.GetViteAssetPath(assetPath)
	assetPathInManifest := v.DistDir
	for k, m := range TestManifest {
		if k == string(assetPath) {
			assetPathInManifest = assetPathInManifest + "/" + m.File
			break
		}
	}

	if path.String() != "" {
		t.Fatalf("vite did not return the asset path from the manifest expected: %s and %s", path.String(), assetPathInManifest)
	}

	if !strings.Contains(buf.String(), "jet was unable parse the vite manifest") {
		t.Fatalf("vite did not log the expected error message: \n%s", buf.String())
	}
}
