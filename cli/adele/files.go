package main

import (
	"embed"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

//go:embed templates
var templateFS embed.FS

func copyFileFromTemplate(templatePath, targetFile string) error {
	if fileExists(targetFile) {
		return errors.New(targetFile + " already exists!")
	}

	data, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return err
	}

	err = copyDataToFile(data, targetFile)
	if err != nil {
		return err
	}

	return nil
}

func copyDataToFile(data []byte, to string) error {
	err := ioutil.WriteFile(to, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func fileExists(fileToCheck string) bool {
	if _, err := os.Stat(fileToCheck); os.IsNotExist(err) {
		return false
	}
	return true
}

func deleteFile(fileToDelete string) error {
	return os.Remove(fileToDelete)
}

// IsAdeleApp reports whether the current working directory looks like the root
// of an adele application. We use the framework module path in ./go.mod as the
// marker — it's written by `adele new` from templates/go.mod.txt and is
// authoritative for "this is an adele project."
func IsAdeleApp() bool {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "github.com/cidekar/adele-framework")
}
