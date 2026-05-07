package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cidekar/adele-framework/helpers"
	"github.com/fatih/color"
)

// keyEnvVar is the name of the env var that holds the application key.
// Matches the placeholder substitution in templates/env.txt where APP_KEY=${KEY}.
const keyEnvVar = "APP_KEY"

// keyEnvFile is the env file the command reads/writes, relative to CWD.
const keyEnvFile = ".env"

// keyLength is the byte length passed to helpers.RandomString. Matches the
// length used by `adele new` when seeding the initial APP_KEY.
const keyLength = 32

// InstallKey generates a new application key, writes it into .env (creating or
// updating the APP_KEY entry), and prints the new key to stdout. Mirrors
// Laravel's `php artisan key:generate` workflow. Run from the root of an
// adele application; refuses to write outside one.
//
// If APP_KEY already has a non-empty value and --force is not set, the command
// prompts before overwriting so an accidental run does not invalidate signed
// session cookies in a live environment.
type InstallKey struct {
	Force bool
}

func NewInstallKey(force bool) *InstallKey {
	return &InstallKey{Force: force}
}

func (c *InstallKey) Handle() error {
	if !IsAdeleApp() {
		return errors.New("adele install key must be run from the root of an adele application (no go.mod referencing the framework)")
	}

	envPath := keyEnvFile
	exists := fileExists(envPath)

	// Prompt before overwriting an existing non-empty key unless --force.
	if exists && !c.Force {
		current, err := readEnvValue(envPath, keyEnvVar)
		if err != nil {
			return fmt.Errorf("read %s: %w", envPath, err)
		}
		if current != "" {
			confirm, err := confirmOverwriteKey()
			if err != nil {
				return err
			}
			if !confirm {
				return errors.New("key install cancelled — existing APP_KEY left unchanged")
			}
		}
	}

	h := helpers.Helpers{}
	newKey := h.RandomString(keyLength)

	if err := writeEnvValue(envPath, keyEnvVar, newKey); err != nil {
		return fmt.Errorf("write %s: %w", envPath, err)
	}

	color.Green("APP_KEY written to %s", envPath)
	fmt.Println(newKey)
	return nil
}

// readEnvValue returns the value of `key` from a dotenv-style file, or the
// empty string if the key is not present. Lines beginning with '#' are
// treated as comments. Quoted values have their surrounding quotes stripped
// to match godotenv's parsing.
func readEnvValue(path, key string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	prefix := key + "="
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		raw := strings.TrimPrefix(trimmed, prefix)
		raw = strings.TrimSpace(raw)
		// Strip a single pair of surrounding quotes if present.
		if len(raw) >= 2 && (raw[0] == '"' || raw[0] == '\'') && raw[0] == raw[len(raw)-1] {
			raw = raw[1 : len(raw)-1]
		}
		return raw, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// writeEnvValue sets `key=value` in a dotenv-style file. If the file does not
// exist it is created. If the key is already present, the existing line is
// rewritten in place to preserve surrounding lines and comments. If absent,
// the entry is appended. Writes via a temp file + rename so a partial write
// can never leave the env in a corrupt state.
func writeEnvValue(path, key, value string) error {
	var lines []string
	found := false

	if fileExists(path) {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(f)
		prefix := key + "="
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if !found && strings.HasPrefix(trimmed, prefix) {
				lines = append(lines, fmt.Sprintf("%s=%s", key, value))
				found = true
				continue
			}
			lines = append(lines, line)
		}
		if err := scanner.Err(); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	// Write atomically: tmp file in the same directory, then rename.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".env.tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		// Best-effort cleanup if rename never happened.
		os.Remove(tmpName)
	}()

	for _, line := range lines {
		if _, err := tmp.WriteString(line + "\n"); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// confirmOverwriteKey reads stdin and returns true only on "yes". Any other
// answer (or read error) returns false. Prompt phrasing makes the production
// risk explicit so an operator does not click through habitually.
func confirmOverwriteKey() (bool, error) {
	fmt.Println("APP_KEY already has a value in .env.")
	fmt.Println("Overwriting will invalidate any data signed with the current key")
	fmt.Println("(session cookies, encrypted columns, signed URLs).")
	fmt.Print("Type 'yes' to overwrite, anything else to cancel: ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(answer)) == "yes", nil
}
