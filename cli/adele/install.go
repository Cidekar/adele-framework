package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var InstallCommand = &Command{
	Name:        "install",
	Help:        "Install a kit into the current project",
	Description: "Install a packaged kit (such as a frontend pipeline) into the current working directory",
	Usage:       "adele install <kit> [options]",
	Examples: []string{
		"adele install starter-kit",
		"adele install starter-kit --skip",
		"adele install starter-kit --no-tailwind",
		"adele install starter-kit --vue",
		"adele install starter-kit --vue=3 --no-tailwind",
		"adele install starter-kit --vue=2",
		"adele install starter-kit --with-auth",
	},
	Options: map[string]string{
		"--skip":        "keep your existing templates; you must wire up the toolchain manually",
		"--no-tailwind": "skip Tailwind CSS scaffolding (default: included)",
		"--vue":         "scaffold a Vue starter (default: Vue 3)",
		"--vue=N":       "scaffold Vue version N (2 or 3)",
		"--with-auth":   "scaffold a working password-auth flow (vanilla only; vue/vue2 coming soon)",
	},
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
		// Resolve flags BEFORE the adele-app gate so an invalid value (e.g.
		// --vue=4) errors out without first prompting the user to scaffold a
		// new app and waiting through a git clone.
		skip := HasOption("--skip")
		withTailwind := !HasOption("--no-tailwind")
		withAuth := HasOption("--with-auth")
		variant, err := resolveStarterKitVariant()
		if err != nil {
			return err
		}
		if withAuth && variant.name != "vanilla" {
			return fmt.Errorf("--with-auth is currently supported only with vanilla; vue and vue2 support is coming soon")
		}
		justScaffolded, err := ensureAdeleApp()
		if err != nil {
			return err
		}
		// If we just scaffolded the app, the user already consented to the
		// install in the scaffold prompt — skipping the second "overwrite or
		// remove?" gate avoids friction on the empty target.
		return NewStarterKit(variant, skip, withTailwind, justScaffolded, withAuth).Handle()
	default:
		return fmt.Errorf("unknown kit %q (available: starter-kit)", kit)
	}
}

// ensureAdeleApp gates `adele install starter-kit` on being inside an adele
// application root. If the CWD is not an adele app, the user can opt into
// scaffolding one via `adele new <name>` and chdir into it; declining (or
// providing an empty name) returns a hard error so the install never lands
// in a foreign project.
//
// Returns justScaffolded=true when this call created the project — the caller
// uses that to suppress the redundant "overwrite or remove?" confirmation
// downstream (the user already consented at the scaffold prompt).
func ensureAdeleApp() (justScaffolded bool, err error) {
	if IsAdeleApp() {
		return false, nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Adele install needs to run inside an adele application root, but the current")
	fmt.Println("directory does not look like one (no go.mod referencing the framework).")
	fmt.Println()
	fmt.Println("Would you like to scaffold a new adele application here? This will run")
	fmt.Println("`adele new <name>` and then continue with the starter-kit install.")
	fmt.Println()
	fmt.Print("Type 'yes' to scaffold, anything else to cancel: ")

	answer, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(answer)) != "yes" {
		return false, errors.New("adele install starter-kit must be run from the root of an adele application")
	}

	fmt.Print("Application name (alphanumeric + dashes/underscores): ")
	nameLine, _ := reader.ReadString('\n')
	appName := strings.TrimSpace(nameLine)
	if appName == "" {
		return false, errors.New("scaffold cancelled: empty application name")
	}

	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)
	Registry.SetArgs([]string{"new", appName})

	if err := NewApplication().Handle(); err != nil {
		return false, fmt.Errorf("scaffold failed: %w", err)
	}

	if err := os.Chdir("./" + appName); err != nil {
		return false, fmt.Errorf("scaffold succeeded but chdir into %q failed: %w", appName, err)
	}

	return true, nil
}

// resolveStarterKitVariant maps the --vue flag to a stackVariant. Bare --vue
// (no value) and --vue=3 both select Vue 3; --vue=2 selects Vue 2.
//
// We parse Registry.GetOptions() directly rather than calling GetOption("--vue")
// because GetOption's bare-flag fallback indexes into os.Args[1:] using the
// position from the filtered options list, which leaks neighbor positional args
// (e.g. "starter-kit") as the value.
func resolveStarterKitVariant() (stackVariant, error) {
	for _, opt := range Registry.GetOptions() {
		flag, value, hasValue := strings.Cut(opt, "=")
		if strings.TrimLeft(flag, "-") != "vue" {
			continue
		}
		if !hasValue {
			return vue3Variant, nil
		}
		switch value {
		case "", "3":
			return vue3Variant, nil
		case "2":
			return vue2Variant, nil
		default:
			return stackVariant{}, fmt.Errorf("invalid --vue=%s; valid values are 2 or 3", value)
		}
	}
	return vanillaVariant, nil
}

func init() {
	if err := Registry.Register(InstallCommand); err != nil {
		panic(fmt.Sprintf("Failed to register install command: %v", err))
	}
}
