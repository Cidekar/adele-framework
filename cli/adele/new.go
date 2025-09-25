package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cidekar/adele-framework"
	"github.com/cidekar/adele-framework/helpers"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var NewApplicationCommand = &Command{
	Name:        "new",
	Help:        "Create a fresh application",
	Description: "Create a fresh application",
	Usage:       "new [arg] [options]",
	Examples:    []string{"adele new myapp", "adele new myapp -p=/my/app", "adele new myapp --version=1.2"},
	Options: map[string]string{
		"-b=, --branch=":  "specify a git branch of adele to use when installing",
		"-v=, --version=": "specify a version of adele to install",
	},
}

var repositoryRoot = "github.com/cidekar"

// Register command on package init
func init() {
	if err := Registry.Register(NewApplicationCommand); err != nil {
		panic(fmt.Sprintf("Failed to register new application command: %v", err))
	}
}

type CommandNewApplication interface {
	Handle() error
	AddBinary() error
	Clone(string) error
	Sanitize(string) string
	UpdateSource(string) error
	Validate(string) error
	Write(string, string) error
}

type NewApp struct {
	command CommandNewApplication
}

func NewApplication() *NewApp {
	return &NewApp{}
}

func (c *NewApp) AddBinary() error {
	return nil
}

func (c *NewApp) Clone(name string) error {
	path := fmt.Sprintf("./%s", name)
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return fmt.Errorf("%s already exsists", name)
	}

	options := &git.CloneOptions{
		URL:      "https://" + repositoryRoot + "/adele.git",
		Progress: os.Stdout,
		Depth:    1,
	}

	if HasOption("--branch") {
		b, err := GetOption("--branch")
		if err != nil {
			return err
		} else {
			options.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", b))
			options.SingleBranch = true
		}

	} else if HasOption("-b") {
		b, err := GetOption("-b")
		if err != nil {
			return err
		} else {
			options.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", b))
			options.SingleBranch = true
		}

	}

	message := fmt.Sprintf("Cloning repository %s", options.URL)

	if options.ReferenceName != "" {
		message += fmt.Sprintf(" (branch: %s)", options.ReferenceName.Short())
	}

	color.Yellow(message + "....")

	_, err := git.PlainClone(path, false, options)

	if err != nil {
		return err
	}

	err = os.RemoveAll(fmt.Sprintf("./%s/.git", name))
	if err != nil {
		return err
	}

	return nil
}

func (c *NewApp) Sanitize(filename string) (string, error) {

	// Keep only alphanumeric, spaces, dots, hyphens, underscores
	safe := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	sanitized := safe.ReplaceAllString(filename, "")

	// Replace multiple spaces with single space
	spaces := regexp.MustCompile(`\s+`)
	sanitized = spaces.ReplaceAllString(sanitized, " ")

	// Trim and handle empty result
	sanitized = strings.TrimSpace(sanitized)
	if sanitized == "" {
		return "", errors.New("invalid application name provided")
	}

	// Force a lowercase
	name := strings.ToLower(sanitized)

	return name, nil
}

func (c *NewApp) Validate() error {

	args := Registry.GetArgs()

	if len(args) == 1 {
		return errors.New("you must provide a name for the application")
	}

	return nil
}

func (c *NewApp) UpdateApplicationNameInSource(name string) error {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		return updateSourceFiles(path, info, err, name)
	})
	if err != nil {
		return err
	}
	return nil
}

func updateSourceFiles(path string, fi os.FileInfo, err error, name string) error {

	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	matched, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}
	if matched {
		read, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newCont := strings.Replace(string(read), "myapp", name, -1)
		err = os.WriteFile(path, []byte(newCont), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *NewApp) Write(name string) error {

	color.Yellow("Writing files to disk...")

	data, err := templateFS.ReadFile("templates/env.txt")
	if err != nil {
		return err
	}

	env := string(data)
	env = strings.ReplaceAll(env, "${APP_NAME}", name)

	helpers := helpers.Helpers{}
	env = strings.ReplaceAll(env, "${KEY}", helpers.RandomString(32))

	err = copyDataToFile([]byte(env), fmt.Sprintf("./%s/.env", name))
	if err != nil {
		return err
	}

	// Determine the source file based on OS
	var sourceFile string
	if runtime.GOOS == "windows" {
		sourceFile = fmt.Sprintf("./%s/Makefile.windows", name)
	} else {
		sourceFile = fmt.Sprintf("./%s/Makefile.mac", name)
	}

	// Copy the proper makefile into the project
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(fmt.Sprintf("./%s/Makefile", name))
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	_ = os.Remove("./" + name + "/Makefile.mac")
	_ = os.Remove("./" + name + "/Makefile.windows")

	// Copy the mod file into the application and update the default references for path and package
	_ = os.Remove("./" + name + "go.mod")
	data, err = templateFS.ReadFile("templates/go.mod.txt")
	if err != nil {
		return err
	}

	mod := string(data)
	mod = strings.ReplaceAll(mod, "${APP_MODULE_PATH}", name)
	mod = strings.ReplaceAll(mod, "${ADELE_PACKAGE_VERSION}", adele.Version)

	err = copyDataToFile([]byte(mod), "./"+name+"/go.mod")
	if err != nil {
		return err
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Change directories into the git clone and update the default application anme in all
	// soruce files to the name provided by the command during execution.
	// Example:
	// 	$ adele new awsomeapp
	// 	import myapp/models -> awsomeapp/modles
	os.Chdir(fmt.Sprintf("./%s", name))
	err = c.UpdateApplicationNameInSource(name)
	if err != nil {
		return err
	}

	// Restore original directory
	err = os.Chdir(originalDir)
	if err != nil {
		return err
	}

	return nil
}

func (c *NewApp) Handle() error {

	err := c.Validate()
	if err != nil {
		return (err)
	}

	args := Registry.GetArgs()
	appName := args[1]
	name, err := c.Sanitize(appName)
	if err != nil {
		return (err)
	}

	err = c.Clone(name)
	if err != nil {
		return err
	}

	err = c.Write(name)
	if err != nil {
		return (err)
	}

	if flag.Lookup("test.v") != nil {
		color.Yellow("\tRunning go test ... skip AddBinary")
		return nil
	} else {
		err = c.AddBinary()
		if err != nil {
			return err
		}
	}

	color.Yellow("Cleaning up...")
	color.Yellow("Done")
	color.White("Adele provides a solid foundation for building modern web applications. Congratulations—you've skipped writing a lot of boilerplate code, so now you can focus on creating your application. We hope you build something awesome!")

	return nil
}
