package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type StarterKit struct {
	variant      stackVariant
	skip         bool
	withTailwind bool
	// preconfirmed suppresses the "overwrite or remove?" prompt. Set when the
	// caller has already obtained explicit consent for the destructive write
	// (e.g. ensureAdeleApp just scaffolded the app this run).
	preconfirmed bool
	// withAuth opts in to the aerra auth scaffold (handlers, middleware, data,
	// migrations, dashboard views, and replaced routes-web). The caller is
	// responsible for ensuring the chosen variant supports auth (vanilla only
	// today); install.go enforces that gate before constructing this struct.
	withAuth bool
}

func NewStarterKit(variant stackVariant, skip, withTailwind, preconfirmed, withAuth bool) *StarterKit {
	return &StarterKit{
		variant:      variant,
		skip:         skip,
		withTailwind: withTailwind,
		preconfirmed: preconfirmed,
		withAuth:     withAuth,
	}
}

// resolveFiles returns the destination -> embed-path map for this install run.
// When tailwind is disabled, tailwind-only destinations are dropped and the
// noTailwind overrides replace the matching base entries. When withAuth is
// true the aerra shared maps (views, code, migrations, dashboard views) are
// merged in on top.
func (c *StarterKit) resolveFiles() map[string]string {
	out := make(map[string]string, len(c.variant.base))
	maps.Copy(out, c.variant.base)
	if !c.withTailwind {
		for dest := range c.variant.tailwindOnly {
			delete(out, dest)
		}
		maps.Copy(out, c.variant.noTailwind)
	}
	if c.withAuth {
		maps.Copy(out, sharedAuthViews)
		maps.Copy(out, sharedAuthCode)
		maps.Copy(out, sharedAuthMigrations)
		maps.Copy(out, sharedDashboardViews)
		maps.Copy(out, sharedAuthOverrides)
	}
	return out
}

// confirmDestructive requires the literal full word "yes" (case-insensitive) to
// proceed. Anything else (including a bare "y") cancels — overwriting templates
// is destructive enough that we want a deliberate confirmation, not a stray
// keystroke.
func (c *StarterKit) confirmDestructive(affected []string) bool {
	color.Yellow("Adele Starter Kit (%s)", c.variant.name)
	color.Yellow("These files will be overwritten or removed:")
	for _, f := range affected {
		color.Yellow("  - %s", f)
	}
	color.Yellow("Any local edits to these files will be lost.")
	color.White("\nType 'yes' to confirm replacement, anything else to cancel: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(line)) == "yes"
}

func (c *StarterKit) Handle() error {
	fmt.Printf("Adele Starter Kit\n\n")

	if c.skip {
		return c.handleSkip()
	}

	files := c.resolveFiles()
	managed := managedFiles()

	var existing []string
	for _, p := range managed {
		if fileExists(p) {
			existing = append(existing, p)
		}
	}

	return c.install(files, managed, existing)
}

// readAppName parses ./go.mod and returns the value of the `module` directive.
// Substituted into staged template files in place of the `$APPNAME$` token so
// imports like `$APPNAME$/data` resolve to the user's actual module path.
// Empty return is treated as a no-op by stageFiles — the staged file is left
// with the literal token if the module line is missing or malformed.
func readAppName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

func (c *StarterKit) handleSkip() error {
	color.Yellow("--skip set: keeping your existing templates.")
	color.Yellow("You will need to wire up the build toolchain yourself:")
	color.Yellow("  - install %s deps in your package.json", c.variant.name)
	color.Yellow("  - configure vite.config.ts with the appropriate plugin")
	color.Yellow("  - import Tailwind in your CSS if you use it")

	return c.UpdateGitignore()
}

// install stages every file under a temp dir, snapshots existing managed files
// into the staging dir's backup/ subdir, removes the originals, then promotes
// the staged files into the project. Any failure during the move triggers a
// rollback from the snapshot so the project is never left in a half-written
// state.
func (c *StarterKit) install(files map[string]string, managed []string, existing []string) error {
	if len(existing) > 0 && !c.preconfirmed {
		if !c.confirmDestructive(existing) {
			color.Yellow("Replacement cancelled.")
			color.Yellow("Re-run with --skip to keep your templates and wire up the toolchain manually.")
			return errors.New("install cancelled")
		}
	}

	if err := c.WriteDir(); err != nil {
		return err
	}

	stageRoot, err := os.MkdirTemp("", "adele-starterkit-*")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(stageRoot)

	appName := readAppName()

	if err := c.stageFiles(stageRoot, files, appName); err != nil {
		return fmt.Errorf("staging failed; project untouched: %w", err)
	}

	backupRoot := filepath.Join(stageRoot, "backup")
	if err := c.snapshotExisting(backupRoot, existing); err != nil {
		return fmt.Errorf("backup failed; project untouched: %w", err)
	}

	for _, p := range managed {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			_ = c.rollback(backupRoot, existing)
			return fmt.Errorf("failed to remove %s during cleanup; rolled back: %w", p, err)
		}
	}

	committed, err := c.commitStaged(stageRoot, files)
	if err != nil {
		c.removeOrphans(committed, existing)
		_ = c.rollback(backupRoot, existing)
		return fmt.Errorf("commit failed; rolled back: %w", err)
	}

	if err := c.UpdateGitignore(); err != nil {
		return err
	}

	color.Green("Installed: %s", c.variant.name)
	return c.Help()
}

func (c *StarterKit) stageFiles(stageRoot string, files map[string]string, appName string) error {
	for dest, embedPath := range files {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", embedPath, err)
		}
		if appName != "" {
			data = bytes.ReplaceAll(data, []byte("$APPNAME$"), []byte(appName))
		}
		// The aerra .go templates carry `//go:build aerra_template` so the
		// in-repo `go test ./...` walk skips them. Once installed they need to
		// compile in the user's app, so strip that constraint (and any
		// immediately following blank line) on the way out.
		data = stripAerraBuildTag(data)
		target := filepath.Join(stageRoot, dest)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

// stripAerraBuildTag removes a leading `//go:build aerra_template` line (and
// at most one trailing blank line) from the embedded template payload. The
// constraint exists only to keep the in-repo `go test ./...` from trying to
// compile template Go files; once we copy them into the user's project they
// must compile, so the tag is stripped at install time.
func stripAerraBuildTag(data []byte) []byte {
	const tag = "//go:build aerra_template"
	if !bytes.HasPrefix(data, []byte(tag)) {
		return data
	}
	rest := data[len(tag):]
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	}
	return rest
}

func (c *StarterKit) snapshotExisting(backupRoot string, existing []string) error {
	for _, p := range existing {
		target := filepath.Join(backupRoot, p)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

// commitStaged writes every staged file into its final destination. It returns
// the list of destinations that were successfully written so the caller can
// remove orphans (files that did not pre-exist) on rollback.
func (c *StarterKit) commitStaged(stageRoot string, files map[string]string) ([]string, error) {
	var written []string
	for dest := range files {
		src := filepath.Join(stageRoot, dest)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return written, err
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return written, err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return written, err
		}
		written = append(written, dest)
	}
	return written, nil
}

// removeOrphans deletes every committed destination that did not exist before
// the install. Combined with rollback() (which restores pre-existing files from
// the backup snapshot), this guarantees a failed mid-commit leaves the project
// in its pre-install state.
func (c *StarterKit) removeOrphans(committed, existing []string) {
	pre := make(map[string]struct{}, len(existing))
	for _, p := range existing {
		pre[p] = struct{}{}
	}
	for _, p := range committed {
		if _, wasPresent := pre[p]; wasPresent {
			continue
		}
		_ = os.Remove(p)
	}
}

func (c *StarterKit) rollback(backupRoot string, existing []string) error {
	var lastErr error
	for _, p := range existing {
		src := filepath.Join(backupRoot, p)
		data, err := os.ReadFile(src)
		if err != nil {
			lastErr = err
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			lastErr = err
			continue
		}
		if err := os.WriteFile(p, data, 0644); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

const (
	gitignoreHeader     = "# Adele starter-kit"
	gitignoreNodeMod    = "node_modules/"
	gitignorePublicDist = "public/dist/"
	gitignorePath       = ".gitignore"
)

// scanGitignore reports which of the adele-managed entries are already present
// in an existing .gitignore. All checks use trimmed-line equality so a substring
// inside a comment doesn't fool us.
func scanGitignore(data []byte) (hasNode, hasDist, hasHeader bool, err error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		switch strings.TrimSpace(scanner.Text()) {
		case gitignoreNodeMod:
			hasNode = true
		case gitignorePublicDist:
			hasDist = true
		case gitignoreHeader:
			hasHeader = true
		}
	}
	return hasNode, hasDist, hasHeader, scanner.Err()
}

func (c *StarterKit) UpdateGitignore() error {
	color.Green("  Updating .gitignore...")

	var existing []byte
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existing = data
	} else if !os.IsNotExist(err) {
		return err
	}

	hasNode, hasDist, hasHeader, err := scanGitignore(existing)
	if err != nil {
		return err
	}

	if hasNode && hasDist {
		color.Green("    .gitignore already up to date")
		return nil
	}

	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteString("\n")
	}

	bothMissing := !hasNode && !hasDist
	if bothMissing && !hasHeader {
		if len(existing) > 0 {
			b.WriteString("\n")
		}
		b.WriteString(gitignoreHeader)
		b.WriteString("\n")
	}

	if !hasNode {
		b.WriteString(gitignoreNodeMod)
		b.WriteString("\n")
		color.Green("    + " + gitignoreNodeMod)
	}
	if !hasDist {
		b.WriteString(gitignorePublicDist)
		b.WriteString("\n")
		color.Green("    + " + gitignorePublicDist)
	}

	return os.WriteFile(gitignorePath, []byte(b.String()), 0644)
}

func (c *StarterKit) WriteDir() error {
	color.Green("  Creating directories...")
	dirs := []string{"resources/js", "resources/css", "resources/views/layouts"}
	for _, p := range dirs {
		if err := os.MkdirAll(p, 0755); err != nil {
			color.Yellow("%s", err)
		}
		color.Green("    " + p)
	}
	return nil
}

func (c *StarterKit) Help() error {
	color.Yellow("\nThe %s starter kit installation is complete. Next steps:", c.variant.name)
	color.Green("  1. Resolve Go dependencies:")
	fmt.Println("     - $ go mod tidy")
	color.Green("  2. Install package dependencies:")
	fmt.Println("     - $ npm install")
	color.Green("  3. Bundle assets for deployment:")
	fmt.Println("     - $ npm run build")
	if c.withAuth {
		color.Green("  4. Configure your database in .env:")
		fmt.Println("     - DATABASE_TYPE=postgres (or mysql)")
		fmt.Println("     - DATABASE_HOST, DATABASE_PORT, DATABASE_USER,")
		fmt.Println("       DATABASE_PASSWORD, DATABASE_NAME")
		color.Green("  5. Run database migrations to create the users + remember_tokens tables:")
		fmt.Println("     - $ make migrate:up")
		fmt.Println("     - or directly: $ adele migrate up")
		color.Yellow("\nWithout step 5, /registration and /login will 500 with \"relation \\\"users\\\" does not exist\".")
	}
	fmt.Println()
	return nil
}
