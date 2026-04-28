package vite

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

type Vite struct {
	ClientPath      string
	DevelopmentMode bool
	DevServer       string
	ErrorLog        *log.Logger
	InfoLog         *log.Logger
	Host            string
	Manifest        string
	ManifestPath    string
	Port            string
	DistDir         string
	RootDir         string
}

type ViteManifest struct {
	Chunk map[string]ViteManifestEntry
}

type ViteManifestEntry struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	SrcPath string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	Css     []string `json:"css"`
	Assets  []string `json:"assets"`
}

func New(host, port, rootdir, distdir, manifest, developmentmode string, infoLog *log.Logger, errorLog *log.Logger) *Vite {
	v := new(Vite)

	v.InfoLog = infoLog
	v.ErrorLog = errorLog

	if host == "" {
		v.Host = "localhost"
	} else {
		v.Host = host
	}

	if port == "" {
		v.Port = "4001"
	} else {
		v.Port = port
	}

	if rootdir == "" {
		v.RootDir = "resources/views"
	} else {
		v.RootDir = rootdir
	}

	if manifest == "" {
		v.Manifest = "manifest.json"
	} else {
		v.Manifest = manifest
	}

	if distdir == "" {
		v.DistDir = "/public/dist"
	} else {
		v.DistDir = "/public/" + distdir
	}

	v.ClientPath = "//" + v.Host + ":" + v.Port + "/@vite/client"
	v.ManifestPath = "." + v.DistDir + "/" + v.Manifest

	if developmentmode == "" {
		v.DevelopmentMode = false
	} else {
		isDevServer, _ := strconv.ParseBool(developmentmode)
		v.DevelopmentMode = isDevServer
		v.DevServer = "//" + v.Host + ":" + v.Port
	}

	return v
}

// Return a path to an asset from the Vite development server in devlopment mode or lookup the path to a Vite resource though the Vite manifest
func (v *Vite) GetViteAssetPath(b []byte) reflect.Value {

	if v.DevelopmentMode {
		return reflect.ValueOf(v.DevServer + string(b[:]))
	}

	var manifest map[string]ViteManifestEntry

	data, err := os.ReadFile(v.ManifestPath)
	if err != nil {
		v.ErrorLog.Println("jet was unable locate the vite manifest in " + v.ManifestPath)
		return reflect.ValueOf("")
	}

	err = json.Unmarshal(data, &manifest)

	if err != nil {
		v.ErrorLog.Println("jet was unable parse the vite manifest")
		return reflect.ValueOf("")
	}

	for k, m := range manifest {
		fmt.Println(k, m, string(b[:]))
		if k == string(b[:]) {
			return reflect.ValueOf(v.DistDir + "/" + m.File)
		}
	}

	return reflect.ValueOf("")
}

// Read the Vite manifest file from disk and render a Vite manfiest during Jet template render.
func (v *Vite) ParseViteBuildManifest() ViteManifest {

	var manifest map[string]ViteManifestEntry
	data, err := os.ReadFile(v.ManifestPath)
	if err != nil {
		v.ErrorLog.Printf("jet was unable locate the vite manifest in %s\n", v.ManifestPath)
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		v.ErrorLog.Printf("jet was unable parse the vite manifest\n")
	}

	var viteManifest ViteManifest
	viteManifest.Chunk = map[string]ViteManifestEntry{}
	for k, m := range manifest {

		// Normalize file path
		m.File = filepath.Clean(fmt.Sprintf("%s/%s", v.DistDir, m.File))

		// Normalize css path
		if len(m.Css) > 0 {
			for c, cf := range m.Css {
				fmt.Println(c, cf)
				m.Css[c] = filepath.Clean(fmt.Sprintf("%s/%s", v.DistDir, cf))
			}
		}

		// Normalize assets path
		if len(m.Assets) > 0 {
			for a, af := range m.Assets {
				m.Assets[a] = filepath.Clean(fmt.Sprintf("%s/%s", v.DistDir, af))
			}
		}

		viteManifest.Chunk[k] = m

	}

	return viteManifest
}
