package main

import (
	"fmt"
)

var InstallCommand = &Command{
	Name:        "install",
	Help:        "Install a kit into the current project",
	Description: "Install a packaged kit (such as a frontend pipeline) into the current working directory",
	Usage:       "adele install <kit> [options]",
	Examples:    []string{"adele install starter-kit"},
	Options:     map[string]string{},
}

type Install struct{}

func NewInstall() *Install {
	return &Install{}
}

func (c *Install) Handle() error {
	args := Registry.GetArgs()
	if len(args) < 2 {
		return fmt.Errorf("missing kit name (available: starter-kit)\nusage: %s", InstallCommand.Usage)
	}

	kit := args[1]
	switch kit {
	case "starter-kit":
		k := NewStarterKit()
		return k.Handle()
	default:
		return fmt.Errorf("unknown kit %q (available: starter-kit)", kit)
	}
}

func init() {
	if err := Registry.Register(InstallCommand); err != nil {
		panic(fmt.Sprintf("Failed to register install command: %v", err))
	}
}
