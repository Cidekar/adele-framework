package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/fatih/color"
)

//go:embed *.go
var efs embed.FS

// main is the CLI application entry point that handles command-line argument parsing and routing.
// Supports help flag detection, command validation, and fallback to help table for invalid commands.
// Uses a two-stage approach: registry lookup for command validation, then CLI handler for execution.
//
// Flow:
//  1. Parse command-line arguments
//  2. Check if command exists in registry
//  3. Handle help flags (--help, -h) with detailed command info
//  4. Execute command through CLI handler with validated arguments
//  5. Display help table if no command provided or command not found
//
// Example usage:
//
//	./adele version --help     // Shows version command details
//	./adele new myproject      // Creates new project
//	./adele invalidcmd         // Shows "Command not found" + help table
func main() {

	args := os.Args[1:]

	if len(args) > 0 {
		if cmd, exists := Registry.GetCommand(args[0]); exists {
			Registry.ParseCmdArgs()
			if HasOption("--help") || HasOption("-h") {
				PrintCommandDetails(cmd)
				os.Exit(0)
			} else {
				c := &Cli{}
				err := c.Handle()
				if err != nil {
					color.Red("Error: %v\n", err)
					os.Exit(0)
				}
			}

		} else {
			if args[0] == "help" {
				PrintHelpTable()
				os.Exit(0)
			}
			fmt.Printf("Command '%s' not found\n\n", args[0])
			PrintHelpTable()
		}
	} else {
		PrintHelpTable()
	}
}

// Handle executes CLI commands by dispatching to appropriate command handlers based on the
// first argument. Uses a switch statement to route commands to their respective handler
// implementations. Returns errors from command execution for display by the caller.
//
// Parameters represent positional arguments after the command name, allowing handlers
// to access command-specific parameters in a structured way.
//
// Example:
//
//	Handle()     // Calls handle with internal pick from Registry to isolate command
func (c *Cli) Handle() error {

	switch Registry.GetCurrentCmd() {

	case "new":
		c := NewApp{}
		err := c.Handle()
		if err != nil {
			return err
		}

	case "version":
		c := NewVersion()
		err := c.Handle()
		if err != nil {
			return err
		}

	case "install":
		c := NewInstall()
		err := c.Handle()
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintCommandDetails displays formatted help information for a specific command with color-coded
// sections. Shows description, usage, examples (if present), and options (if present) with
// proper indentation and alignment. Uses the color library for yellow section headers.
//
// Example output:
//
//	Description:
//	  Run database migrations
//
//	Usage:
//	  adele migrate [options]
//
//	Examples:
//	  adele migrate
//	  adele migrate --rollback
//
//	Options:
//	  -f,--force      skip confirmation prompts
func PrintCommandDetails(cmd *Command) {
	color.Yellow("Description:")
	fmt.Printf("  %s\n\n", cmd.Description)

	color.Yellow("Usage:")
	fmt.Printf("  %s\n\n", cmd.Usage)

	if len(cmd.Examples) > 0 {
		color.Yellow("Examples:")
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", example)
		}
		fmt.Println()
	}

	if len(cmd.Options) > 0 {
		color.Yellow("Options:")
		for flag, desc := range cmd.Options {
			fmt.Printf("  %-15s %s\n", flag, desc)
		}
	}
}

// PrintHelpTable displays a summary of all available commands in the global registry.
// Shows command names in a left-aligned column with their brief help text. Uses color-coded
// header and formats output in a table-like structure for easy scanning.
//
// Example output:
//
//	Available Commands:
//	  version         Show framework version
//	  migrate         Run database migrations
//	  make:controller Generate new controller
func PrintHelpTable() {
	color.Yellow("Available Commands:")
	for name, cmd := range Registry.GetAllCommands() {
		coloredName := color.GreenString(name)
		padding := 15 - len(name)
		spaces := strings.Repeat(" ", padding)
		fmt.Printf("  %s%s %s\n", coloredName, spaces, cmd.Help)
	}
}

// ParseAndRegister validates and adds a command to the registry using reflection-based validation.
// Checks struct field requirements via tags and applies field-specific validation rules.
// Returns error if command fails validation, otherwise stores in registry map.
//
// Example:
//
//	cmd := &Command{Name: "test", Description: "Test command"}
//	err := registry.ParseAndRegister(cmd)
//	if err != nil {
//	    log.Fatal("Invalid command:", err)
//	}
func (cr *CommandRegistry) ParseAndRegister(cmd *Command) error {
	if err := cr.validateCommand(cmd); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	cr.commands[cmd.Name] = cmd
	return nil
}

// validateCommand performs comprehensive validation of command struct using reflection.
// Checks required fields based on struct tags and applies field-specific business rules.
// Used internally by ParseAndRegister to ensure command integrity before registration.
func (cr *CommandRegistry) validateCommand(cmd *Command) error {
	cmdType := reflect.TypeOf(*cmd)
	cmdValue := reflect.ValueOf(*cmd)

	for i := 0; i < cmdType.NumField(); i++ {
		field := cmdType.Field(i)
		value := cmdValue.Field(i)

		// Check required fields
		if required := field.Tag.Get("required"); required == "true" {
			if cr.isEmptyValue(value) {
				return fmt.Errorf("required field %s is empty", field.Name)
			}
		}

		// Validate specific field rules
		if err := cr.validateField(field.Name, value); err != nil {
			return err
		}
	}

	return nil
}

// validateField applies business logic validation rules for specific command fields.
// Enforces naming conventions, format requirements, and content standards for
// individual struct fields. Used by validateCommand during registration process.
func (cr *CommandRegistry) validateField(fieldName string, value reflect.Value) error {
	switch fieldName {
	case "Name":
		name := value.String()
		if strings.Contains(name, " ") {
			return fmt.Errorf("command name cannot contain spaces")
		}
		if len(name) < 2 {
			return fmt.Errorf("command name must be at least 2 characters")
		}
	case "Description":
		description := value.String()
		if strings.TrimSpace(description) == "" {
			return fmt.Errorf("command description must contain a string")
		}
	case "Usage":
		usage := value.String()
		if strings.TrimSpace(usage) == "" {
			return fmt.Errorf("command usage must contain a string")
		}
	}
	return nil
}

// isEmptyValue checks if a reflected value represents an empty/zero state for its type.
// Handles strings, maps, slices, interfaces, and pointers with appropriate empty checks.
// Used by validation logic to determine if required fields contain meaningful data.
func (cr *CommandRegistry) isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// GenerateHelpText creates a formatted help text string for a specific command using
// reflection-based registry lookup. Builds a structured help display including command
// name, description, usage, examples, and options with proper formatting and alignment.
//
// Example:
//
//	registry := NewCommandRegistry()
//	registry.Register(&Command{
//	    Name: "migrate",
//	    Description: "Run database migrations",
//	    Usage: "adele migrate [options]",
//	    Examples: []string{"adele migrate", "adele migrate --rollback"},
//	    Options: map[string]string{"-f,--force": "skip confirmation"},
//	})
//
//	helpText, err := registry.GenerateHelpText("migrate")
//		Returns:
//
// Command: migrate
//
// Description:
//
//	Run database migrations
//
// Usage:
//
//	adele migrate [options]
//
// Examples:
//
//	adele migrate
//	adele migrate --rollback
//
// Options:
//
//	-f,--force      skip confirmation
func (cr *CommandRegistry) GenerateHelpText(cmdName string) (string, error) {
	cmd, exists := cr.commands[cmdName]
	if !exists {
		return "", fmt.Errorf("command %s not found", cmdName)
	}

	var help strings.Builder
	help.WriteString(fmt.Sprintf("Command: %s\n\n", cmd.Name))

	if cmd.Description != "" {
		help.WriteString(fmt.Sprintf("Description:\n  %s\n\n", cmd.Description))
	}

	if cmd.Usage != "" {
		help.WriteString(fmt.Sprintf("Usage:\n  %s\n\n", cmd.Usage))
	}

	if len(cmd.Examples) > 0 {
		help.WriteString("Examples:\n")
		for _, example := range cmd.Examples {
			help.WriteString(fmt.Sprintf("  %s\n", example))
		}
		help.WriteString("\n")
	}

	if len(cmd.Options) > 0 {
		help.WriteString("Options:\n")
		for flag, desc := range cmd.Options {
			help.WriteString(fmt.Sprintf("  %-15s %s\n", flag, desc))
		}
		help.WriteString("\n")
	}

	return help.String(), nil
}

// PrintCommandTable displays detailed information for a specific command using reflection-based
// command registry. Creates a new registry, registers all available commands with validation,
// and generates formatted help text. If the command is not found, returns silently.
//
// Example:
//
//	commands := CommandsHelper{
//	    1: {Name: "version", Description: "Show version info", Usage: "adele version"},
//	    2: {Name: "migrate", Description: "Run migrations", Usage: "adele migrate"},
//	}
//	PrintCommandTable(commands, "version")
//
// Output:
// Command: version
//
// Description:
//
//	Show version info
//
// Usage:
//
//	adele version
func PrintCommandTable(commands CommandsHelper, commandName string) {
	registry := NewCommandRegistry()

	// Register all commands
	for _, cmd := range commands {
		if err := registry.ParseAndRegister(&cmd); err != nil {
			fmt.Printf("Error registering command %s: %v\n", cmd.Name, err)
			continue
		}
	}

	// Generate and print help
	helpText, err := registry.GenerateHelpText(commandName)
	if err != nil {
		PrintHelpTable()
		return
	}

	fmt.Print(helpText)
}

// HasOption checks command line arguments for a given flag and returns true if it exists.
// Supports various flag formats including short flags, long flags, and bare option names.
//
// Example:
//
//	HasOption("--help")  // matches --help
//	HasOption("-h")      // matches -h
//	HasOption("help")    // matches -h, --help
//	HasOption("h")       // matches -h, --help
func HasOption(option string) bool {
	// Normalize the target option (remove leading dashes)
	targetOption := strings.TrimLeft(option, "-")

	for _, arg := range Registry.GetOptions() {
		//for _, arg := range os.Args[1:] {
		// Split on '=' to handle --flag=value format
		flagPart := strings.Split(arg, "=")[0]

		// Remove leading dashes from the argument
		normalizedArg := strings.TrimLeft(flagPart, "-")

		// Compare normalized versions
		if normalizedArg == targetOption {
			return true
		}
	}

	return false
}

// GetOption retrieves the value for a command line flag. Returns the flag's value
// if it exists, or an error if the flag is not found. Handles various flag formats
// including --flag=value, --flag value, -f=value, and -f value.
//
// Example:
//
//	GetOption("help")     // matches --help or -h, returns flag name if no value
//	GetOption("file")     // matches --file=config.json, returns "config.json"
//	GetOption("v")        // matches -v=debug or --verbose=debug, returns "debug"
//	GetOption("port")     // matches --port 8080, returns "8080"
func GetOption(option string) (string, error) {
	targetOption := strings.TrimLeft(option, "-")
	args := os.Args[1:]

	for i, arg := range Registry.GetOptions() {
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Handle --flag=value format
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flagName := strings.TrimLeft(parts[0], "-")
			if flagName == targetOption {
				return parts[1], nil
			}
		} else {
			// Handle --flag or -f format (value might be next arg)
			flagName := strings.TrimLeft(arg, "-")
			if flagName == targetOption {
				// Check if next argument is the value (doesn't start with -)
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					return args[i+1], nil
				}
				// Flag exists but has no value
				return flagName, nil
			}
		}
	}

	return "", errors.New("flag not found: " + option)
}
