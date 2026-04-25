package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type StarterKit struct{}

func NewStarterKit() *StarterKit {
	return &StarterKit{}
}

func (c *StarterKit) Help() error {
	color.Yellow("The starter kit installation is complete, but additional work is required:")
	color.Green("  1. Install the Tailwinds and Vite package dependencies:")
	fmt.Printf("    - $ npm install")
	color.Green("\n  2. Bundle assets for deployment:")
	fmt.Printf("    - $ npm run build")
	fmt.Printf("\n")
	return nil
}

func (c *StarterKit) WriteFiles() error {
	if fileExists("resources/views/layouts/base.jet") {
		err := deleteFile("resources/views/layouts/base.jet")
		if err != nil {
			return err
		}
		color.Green("  Removed asset... base.jet")
	}
	if fileExists("resources/views/home.jet") {
		err := deleteFile("resources/views/home.jet")
		if err != nil {
			return err
		}
		color.Green("  Removed asset... home.jet")
	}
	color.Green("  Copying assets...")
	err := copyFileFromTemplate("templates/vite/base.jet", "resources/views/layouts/base.jet")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/home.jet", "resources/views/home.jet")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/styles.css", "resources/css/styles.css")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/script.ts", "resources/js/script.ts")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/package.json", "package.json")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/vite.config.ts", "vite.config.ts")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/tailwind.config.js", "tailwind.config.js")
	if err != nil {
		return err
	}
	err = copyFileFromTemplate("templates/vite/postcss.config.js", "postcss.config.js")
	if err != nil {
		return err
	}
	return nil
}

func (c *StarterKit) WriteDir() error {
	color.Green("  Creating directories...")
	dirs := []string{"resources/js", "resources/css"}
	for _, path := range dirs {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			color.Yellow("%s", err)
		}
		color.Green("    " + path)
	}
	return nil
}

func (c *StarterKit) Handle() error {
	fmt.Printf("Adele Starter Kit\n\n")
	color.Yellow("This command will replace existing jet templates:")
	color.Yellow(" - /resources/views/layouts/base.jet")
	color.Yellow(" - /resources/views/home.jet")
	color.White("\nDo you wish to continue? [y/N]")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(input)) != "y" {
		return errors.New("command cancelled")
	}

	err := c.WriteDir()
	if err != nil {
		return err
	}
	err = c.WriteFiles()
	if err != nil {
		return err
	}
	err = c.Help()
	if err != nil {
		return err
	}
	return nil
}
