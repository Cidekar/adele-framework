package main

import (
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

// withStdin replaces os.Stdin for the duration of the test with a pipe whose
// read end emits the given input. The original stdin is restored on cleanup.
func withStdin(t *testing.T, input string) {
	t.Helper()

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("Failed to write to stdin pipe: %v", err)
	}
	w.Close()

	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
	})
}

// expectedStarterKitFiles is the list of files the vanilla starter kit handler
// is expected to write into the working directory on a default install (no
// --with-auth). The aerra auth scaffold is opt-in, so its files are listed
// separately in expectedAerraAuthFiles.
var expectedStarterKitFiles = []string{
	"resources/views/layouts/base.jet",
	"resources/views/home.jet",
	"resources/css/styles.css",
	"resources/js/script.ts",
	"package.json",
	"vite.config.ts",
	"tailwind.config.js",
	"postcss.config.js",
}

// expectedAerraAuthFiles is the additional file set that --with-auth pulls in
// on top of the variant base. Used both to assert presence on with-auth
// installs and to assert absence on default installs.
var expectedAerraAuthFiles = []string{
	"resources/views/layouts/application.jet",
	"resources/views/login.jet",
	"resources/views/registration.jet",
	"resources/views/forgot.jet",
	"resources/views/reset-password.jet",
	"resources/views/dashboard/home.jet",
	"resources/views/dashboard/profile.jet",
	"resources/views/dashboard/header.jet",
	"resources/views/dashboard/menu.jet",
	"handlers/aerra.go",
	"handlers/convenience.go",
	"main.go",
	"middleware/authenticated.go",
	"middleware/remember.go",
	"models/user.go",
	"models/remember_token.go",
	"models/models.go",
	"routes-web.go",
	"Makefile",
	"migrations/0001_create_users_table.up.sql",
	"migrations/0001_create_users_table.down.sql",
	"migrations/0002_create_remember_tokens_table.up.sql",
	"migrations/0002_create_remember_tokens_table.down.sql",
}

// expectedVanillaWithAuthFiles is the union of vanilla base files and the
// aerra auth scaffold — what `adele install starter-kit --with-auth` writes.
var expectedVanillaWithAuthFiles = append(append([]string{}, expectedStarterKitFiles...), expectedAerraAuthFiles...)

func TestStarterKit_Handle_CancelOnN(t *testing.T) {
	t.Chdir(t.TempDir())

	// Pre-create one managed file so the prompt fires; an empty project skips
	// the confirmation step entirely.
	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
		t.Fatalf("Failed to seed stub: %v", err)
	}

	withStdin(t, "n\n")

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	err := k.Handle()
	if err == nil {
		t.Fatal("Expected error when user cancels with 'n'")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
		t.Errorf("Expected error to contain 'cancelled', got: %v", err)
	}

	// Stub file should be untouched (not overwritten).
	got, err := os.ReadFile("resources/views/layouts/base.jet")
	if err != nil {
		t.Fatalf("Failed to read base.jet: %v", err)
	}
	if string(got) != "STUB" {
		t.Errorf("Expected stub content preserved on cancel, got: %s", string(got))
	}
}

func TestStarterKit_Handle_CancelOnEmpty(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
		t.Fatalf("Failed to seed stub: %v", err)
	}

	withStdin(t, "\n")

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	err := k.Handle()
	if err == nil {
		t.Fatal("Expected error when user provides empty input")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
		t.Errorf("Expected error to contain 'cancelled', got: %v", err)
	}
}

func TestStarterKit_Handle_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create resources/views/layouts: %v", err)
	}

	// No managed files exist yet, so no prompt fires; stdin is unused.
	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	err := k.Handle()
	if err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedStarterKitFiles {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written", f)
			continue
		}

		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}

	// Spot-check package.json contains "vite" (matches embedded template).
	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkg), "vite") {
		t.Errorf("Expected package.json to mention 'vite', got: %s", string(pkg))
	}

	if !fileExists(".gitignore") {
		t.Fatal("Expected .gitignore to be written by Handle()")
	}
	gi, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	giStr := string(gi)
	if !strings.Contains(giStr, "node_modules/") {
		t.Errorf("Expected .gitignore to contain 'node_modules/', got: %s", giStr)
	}
	if !strings.Contains(giStr, "public/dist/") {
		t.Errorf("Expected .gitignore to contain 'public/dist/', got: %s", giStr)
	}
}

// TestStarterKit_Handle_Preconfirmed_SkipsPrompt verifies that when the caller
// flags the install as preconfirmed (e.g. ensureAdeleApp just scaffolded the
// app), the destructive-confirmation prompt does not fire even when managed
// files exist on disk. The starter-kit just overwrites them.
func TestStarterKit_Handle_Preconfirmed_SkipsPrompt(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	// Seed a managed file. With preconfirmed=false the install would prompt
	// and (with empty stdin) cancel; with preconfirmed=true the prompt is
	// skipped and the install proceeds.
	if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
		t.Fatalf("Failed to seed stub: %v", err)
	}

	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, true, true, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error with preconfirmed=true: %v", err)
	}

	// The stub should have been replaced with the embedded template content.
	got, err := os.ReadFile("resources/views/layouts/base.jet")
	if err != nil {
		t.Fatalf("Failed to read base.jet: %v", err)
	}
	if string(got) == "STUB" {
		t.Error("Expected stub to be overwritten when preconfirmed=true")
	}
}

func TestStarterKit_UpdateGitignore_CreatesNewFile(t *testing.T) {
	t.Chdir(t.TempDir())

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.UpdateGitignore(); err != nil {
		t.Fatalf("UpdateGitignore returned unexpected error: %v", err)
	}

	if !fileExists(".gitignore") {
		t.Fatal("Expected .gitignore to be created")
	}
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Adele starter-kit") {
		t.Errorf("Expected header comment in new .gitignore, got: %s", content)
	}
	if !containsLine(content, "node_modules/") {
		t.Errorf("Expected line 'node_modules/' in new .gitignore, got: %s", content)
	}
	if !containsLine(content, "public/dist/") {
		t.Errorf("Expected line 'public/dist/' in new .gitignore, got: %s", content)
	}
}

func TestStarterKit_UpdateGitignore_AppendsMissingEntries(t *testing.T) {
	t.Chdir(t.TempDir())

	original := "*.log\n"
	if err := os.WriteFile(".gitignore", []byte(original), 0644); err != nil {
		t.Fatalf("Failed to seed .gitignore: %v", err)
	}

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.UpdateGitignore(); err != nil {
		t.Fatalf("UpdateGitignore returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	content := string(data)

	if !strings.HasPrefix(content, original) {
		t.Errorf("Expected original content preserved at start, got: %q", content)
	}
	if strings.Contains(content, "*.lognode_modules/") {
		t.Errorf("Found unintended concatenation '*.lognode_modules/' in: %q", content)
	}
	if !containsLine(content, "*.log") {
		t.Errorf("Expected '*.log' line preserved, got: %q", content)
	}
	if !containsLine(content, "node_modules/") {
		t.Errorf("Expected 'node_modules/' line, got: %q", content)
	}
	if !containsLine(content, "public/dist/") {
		t.Errorf("Expected 'public/dist/' line, got: %q", content)
	}
	if !strings.Contains(content, "# Adele starter-kit") {
		t.Errorf("Expected header comment when both entries are missing, got: %q", content)
	}
}

func TestStarterKit_UpdateGitignore_PartiallyPresent_AppendsOnlyMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	original := "*.log\nnode_modules/\n"
	if err := os.WriteFile(".gitignore", []byte(original), 0644); err != nil {
		t.Fatalf("Failed to seed .gitignore: %v", err)
	}

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.UpdateGitignore(); err != nil {
		t.Fatalf("UpdateGitignore returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	content := string(data)

	if got := countLine(content, "node_modules/"); got != 1 {
		t.Errorf("Expected exactly 1 'node_modules/' line, got %d in: %q", got, content)
	}
	if got := countLine(content, "public/dist/"); got != 1 {
		t.Errorf("Expected exactly 1 'public/dist/' line, got %d in: %q", got, content)
	}
	if strings.Contains(content, "# Adele starter-kit") {
		t.Errorf("Did not expect header comment when only one entry was missing, got: %q", content)
	}
}

func TestStarterKit_UpdateGitignore_FullyPresent_NoOp(t *testing.T) {
	t.Chdir(t.TempDir())

	original := "*.log\nnode_modules/\npublic/dist/\n"
	if err := os.WriteFile(".gitignore", []byte(original), 0644); err != nil {
		t.Fatalf("Failed to seed .gitignore: %v", err)
	}

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.UpdateGitignore(); err != nil {
		t.Fatalf("UpdateGitignore returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	if string(data) != original {
		t.Errorf("Expected .gitignore to be unchanged.\nWant: %q\nGot:  %q", original, string(data))
	}
}

func containsLine(content, target string) bool {
	for line := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(line) == target {
			return true
		}
	}
	return false
}

func countLine(content, target string) int {
	n := 0
	for line := range strings.SplitSeq(content, "\n") {
		if strings.TrimSpace(line) == target {
			n++
		}
	}
	return n
}

func TestStarterKit_Handle_ReplacesExistingJetTemplates(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}

	stub := []byte("STUB CONTENT")
	if err := os.WriteFile("resources/views/layouts/base.jet", stub, 0644); err != nil {
		t.Fatalf("Failed to seed base.jet: %v", err)
	}
	if err := os.WriteFile("resources/views/home.jet", stub, 0644); err != nil {
		t.Fatalf("Failed to seed home.jet: %v", err)
	}

	withStdin(t, "yes\n")

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	err := k.Handle()
	if err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	baseContent, err := os.ReadFile("resources/views/layouts/base.jet")
	if err != nil {
		t.Fatalf("Failed to read base.jet: %v", err)
	}
	if string(baseContent) == string(stub) {
		t.Error("Expected base.jet to be replaced with embedded template content")
	}

	homeContent, err := os.ReadFile("resources/views/home.jet")
	if err != nil {
		t.Fatalf("Failed to read home.jet: %v", err)
	}
	if string(homeContent) == string(stub) {
		t.Error("Expected home.jet to be replaced with embedded template content")
	}
}

func TestStarterKit_Handle_RerunSucceedsAfterCleanup(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create resources/views/layouts: %v", err)
	}

	// First run: nothing on disk, no prompt.
	withStdin(t, "")
	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Second run: managed files now exist, so prompt fires; full-word "yes" must
	// confirm and the install must overwrite cleanly without "already exists".
	withStdin(t, "yes\n")
	k2 := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k2.Handle(); err != nil {
		t.Fatalf("Second run failed: %v", err)
	}

	// Every file present and matches the embedded template byte-for-byte
	// after the install-time transforms (build-tag strip, $APPNAME$ subst).
	// We did not seed a go.mod so $APPNAME$ remains literal in the staged
	// output; only the build-tag strip applies here.
	for dest, embedPath := range vanillaVariant.base {
		got, err := os.ReadFile(dest)
		if err != nil {
			t.Errorf("Failed to read %q after rerun: %v", dest, err)
			continue
		}
		want, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("Failed to read embedded %q: %v", embedPath, err)
			continue
		}
		want = stripAerraBuildTag(want)
		if string(got) != string(want) {
			t.Errorf("Rerun did not overwrite %q with embedded template content", dest)
		}
	}
}

func TestStarterKit_Handle_RequiresFullWordYes(t *testing.T) {
	// Bare "y" must NOT confirm.
	t.Run("bare y cancels", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
			t.Fatalf("Failed to create dirs: %v", err)
		}
		if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
			t.Fatalf("Failed to seed stub: %v", err)
		}

		withStdin(t, "y\n")

		k := NewStarterKit(vanillaVariant, false, true, false, false)
		err := k.Handle()
		if err == nil {
			t.Fatal("Expected cancellation when input is bare 'y'")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	})

	// "Yes" (any case) must confirm.
	t.Run("Yes confirms", func(t *testing.T) {
		t.Chdir(t.TempDir())
		if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
			t.Fatalf("Failed to create dirs: %v", err)
		}
		if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
			t.Fatalf("Failed to seed stub: %v", err)
		}

		withStdin(t, "Yes\n")

		k := NewStarterKit(vanillaVariant, false, true, false, false)
		if err := k.Handle(); err != nil {
			t.Fatalf("Expected 'Yes' to confirm, got error: %v", err)
		}

		got, err := os.ReadFile("resources/views/layouts/base.jet")
		if err != nil {
			t.Fatalf("Failed to read base.jet: %v", err)
		}
		if string(got) == "STUB" {
			t.Error("Expected base.jet to be overwritten when user typed 'Yes'")
		}
	})
}

func TestStarterKit_Handle_SkipFlag_DoesNotWriteTemplates(t *testing.T) {
	t.Chdir(t.TempDir())

	k := NewStarterKit(vanillaVariant, true, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle(skip=true) returned unexpected error: %v", err)
	}

	for _, f := range expectedStarterKitFiles {
		if fileExists(f) {
			t.Errorf("Expected file %q NOT to be written when --skip is set", f)
		}
	}
}

func TestStarterKit_Handle_SkipFlag_StillUpdatesGitignore(t *testing.T) {
	t.Chdir(t.TempDir())

	if fileExists(".gitignore") {
		t.Fatal("Test precondition violated: .gitignore should not exist in fresh tempdir")
	}

	k := NewStarterKit(vanillaVariant, true, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle(skip=true) returned unexpected error: %v", err)
	}

	if !fileExists(".gitignore") {
		t.Fatal("Expected .gitignore to be created when --skip is set")
	}
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}
	content := string(data)
	if !containsLine(content, "node_modules/") {
		t.Errorf("Expected 'node_modules/' line, got: %q", content)
	}
	if !containsLine(content, "public/dist/") {
		t.Errorf("Expected 'public/dist/' line, got: %q", content)
	}
}

func TestStarterKit_WriteDir_CreatesResourceDirs(t *testing.T) {
	t.Chdir(t.TempDir())

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.WriteDir(); err != nil {
		t.Fatalf("WriteDir returned unexpected error: %v", err)
	}

	for _, dir := range []string{"resources/js", "resources/css", "resources/views/layouts"} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Expected %q to be a directory", dir)
		}
	}
}

func TestStarterKit_Help_ReturnsNil(t *testing.T) {
	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.Help(); err != nil {
		t.Errorf("Expected Help() to return nil, got: %v", err)
	}
}

func TestNewStarterKit_NotNil(t *testing.T) {
	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if k == nil {
		t.Fatal("Expected NewStarterKit() to return non-nil")
	}
}

func TestStarterKit_ExpectedFiles_IncludeWriteDirOutputs(t *testing.T) {
	for _, dir := range []string{"resources/js", "resources/css"} {
		found := false
		for _, f := range expectedStarterKitFiles {
			if strings.HasPrefix(f, dir+"/") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected at least one file under %q in expectedStarterKitFiles", dir)
		}
	}
}

func TestManagedFiles_DeduplicatesAndSortsStable(t *testing.T) {
	got := managedFiles()

	seen := map[string]int{}
	for _, p := range got {
		seen[p]++
	}
	for p, n := range seen {
		if n != 1 {
			t.Errorf("Expected %q to appear exactly once in managedFiles(), got %d", p, n)
		}
	}

	wantSet := map[string]struct{}{}
	for _, v := range allVariants {
		for dest := range v.base {
			wantSet[dest] = struct{}{}
		}
		for dest := range v.noTailwind {
			wantSet[dest] = struct{}{}
		}
	}
	// managedFiles() also reserves the aerra extras for cleanup so a re-run
	// without --with-auth sweeps a prior auth install away.
	for dest := range sharedAuthViews {
		wantSet[dest] = struct{}{}
	}
	for dest := range sharedAuthCode {
		wantSet[dest] = struct{}{}
	}
	for dest := range sharedAuthMigrations {
		wantSet[dest] = struct{}{}
	}
	for dest := range sharedDashboardViews {
		wantSet[dest] = struct{}{}
	}
	for dest := range sharedAuthVue3 {
		wantSet[dest] = struct{}{}
	}
	for dest := range sharedAuthVue3ViewOverride {
		wantSet[dest] = struct{}{}
	}
	want := make([]string, 0, len(wantSet))
	for dest := range wantSet {
		want = append(want, dest)
	}
	sort.Strings(want)

	gotSorted := append([]string(nil), got...)
	sort.Strings(gotSorted)

	if len(gotSorted) != len(want) {
		t.Fatalf("managedFiles() length mismatch: got %d, want %d (got=%v)", len(gotSorted), len(want), gotSorted)
	}
	for i := range want {
		if gotSorted[i] != want[i] {
			t.Errorf("managedFiles()[%d] = %q, want %q", i, gotSorted[i], want[i])
		}
	}
}

func TestVanillaVariant_AllEmbedPathsResolve(t *testing.T) {
	for dest, embedPath := range vanillaVariant.base {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vanillaVariant.base[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vanillaVariant.base[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
}

func TestVanillaVariant_NoTailwindEmbedPathsResolve(t *testing.T) {
	for dest, embedPath := range vanillaVariant.noTailwind {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vanillaVariant.noTailwind[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vanillaVariant.noTailwind[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
}

func TestManagedFiles_IncludesNoTailwindDestinations(t *testing.T) {
	got := managedFiles()
	for dest := range vanillaVariant.noTailwind {
		if !slices.Contains(got, dest) {
			t.Errorf("Expected managedFiles() to include noTailwind destination %q", dest)
		}
	}
}

func TestStarterKit_Handle_NoTailwind_DoesNotWriteTailwindConfigFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, false, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	if fileExists("tailwind.config.js") {
		t.Error("Expected tailwind.config.js NOT to be written when tailwind=false")
	}
	if fileExists("postcss.config.js") {
		t.Error("Expected postcss.config.js NOT to be written when tailwind=false")
	}

	if !fileExists("package.json") {
		t.Fatal("Expected package.json to be written")
	}
	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if strings.Contains(string(pkg), "tailwindcss") {
		t.Errorf("Expected package.json NOT to contain 'tailwindcss', got: %s", string(pkg))
	}
	if !strings.Contains(string(pkg), "vite") {
		t.Errorf("Expected package.json to contain 'vite', got: %s", string(pkg))
	}

	if !fileExists("resources/css/styles.css") {
		t.Fatal("Expected resources/css/styles.css to be written")
	}
	css, err := os.ReadFile("resources/css/styles.css")
	if err != nil {
		t.Fatalf("Failed to read styles.css: %v", err)
	}
	if strings.Contains(string(css), "@tailwind") {
		t.Errorf("Expected styles.css NOT to contain '@tailwind', got: %s", string(css))
	}
}

func TestStarterKit_Handle_NoTailwind_AfterTailwindInstall_RemovesConfigFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("First install (tailwind=true) failed: %v", err)
	}
	if !fileExists("tailwind.config.js") {
		t.Fatal("Expected tailwind.config.js after tailwind=true install")
	}

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, false, false, false).Handle(); err != nil {
		t.Fatalf("Second install (tailwind=false) failed: %v", err)
	}

	if fileExists("tailwind.config.js") {
		t.Error("Expected tailwind.config.js to be removed after re-install with tailwind=false")
	}
	if fileExists("postcss.config.js") {
		t.Error("Expected postcss.config.js to be removed after re-install with tailwind=false")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if strings.Contains(string(pkg), "tailwindcss") {
		t.Errorf("Expected package.json NOT to contain 'tailwindcss' after re-install with tailwind=false, got: %s", string(pkg))
	}
}

var expectedVue3Files = []string{
	"resources/views/layouts/base.jet",
	"resources/views/home.jet",
	"resources/css/styles.css",
	"resources/js/main.ts",
	"resources/js/App.vue",
	"package.json",
	"vite.config.ts",
	"tailwind.config.js",
	"postcss.config.js",
}

func TestStarterKit_Handle_Vue3_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vue3Variant, false, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedVue3Files {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written", f)
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgStr := string(pkg)
	if !strings.Contains(pkgStr, `"vue"`) {
		t.Errorf("Expected package.json to contain '\"vue\"', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, "@vitejs/plugin-vue") {
		t.Errorf("Expected package.json to contain '@vitejs/plugin-vue', got: %s", pkgStr)
	}

	app, err := os.ReadFile("resources/js/App.vue")
	if err != nil {
		t.Fatalf("Failed to read App.vue: %v", err)
	}
	appStr := string(app)
	if !strings.Contains(appStr, "<script setup") {
		t.Errorf("Expected App.vue to contain '<script setup', got: %s", appStr)
	}
	if !strings.Contains(appStr, "Vue") {
		t.Errorf("Expected App.vue to contain 'Vue' label, got: %s", appStr)
	}

	main, err := os.ReadFile("resources/js/main.ts")
	if err != nil {
		t.Fatalf("Failed to read main.ts: %v", err)
	}
	if !strings.Contains(string(main), "createApp(App)") {
		t.Errorf("Expected main.ts to contain 'createApp(App)', got: %s", string(main))
	}
}

func TestStarterKit_Handle_Vue3_NoTailwind(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vue3Variant, false, false, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	if fileExists("tailwind.config.js") {
		t.Error("Expected tailwind.config.js NOT to be written when tailwind=false")
	}
	if fileExists("postcss.config.js") {
		t.Error("Expected postcss.config.js NOT to be written when tailwind=false")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgStr := string(pkg)
	if strings.Contains(pkgStr, "tailwindcss") {
		t.Errorf("Expected package.json NOT to contain 'tailwindcss', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, `"vue"`) {
		t.Errorf("Expected package.json to contain '\"vue\"', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, "@vitejs/plugin-vue") {
		t.Errorf("Expected package.json to contain '@vitejs/plugin-vue', got: %s", pkgStr)
	}
}

func TestStarterKit_Handle_VanillaToVue3_VariantSwitch(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vanilla install failed: %v", err)
	}
	if !fileExists("resources/js/script.ts") {
		t.Fatal("Expected script.ts after vanilla install")
	}

	withStdin(t, "yes\n")
	if err := NewStarterKit(vue3Variant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vue3 install failed: %v", err)
	}

	if fileExists("resources/js/script.ts") {
		t.Error("Expected script.ts to be removed after switching to vue3")
	}
	if !fileExists("resources/js/main.ts") {
		t.Error("Expected main.ts to exist after switching to vue3")
	}
	if !fileExists("resources/js/App.vue") {
		t.Error("Expected App.vue to exist after switching to vue3")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkg), `"vue"`) {
		t.Errorf("Expected package.json to contain '\"vue\"' after switch, got: %s", string(pkg))
	}
}

func TestStarterKit_Handle_Vue3ToVanilla_VariantSwitch(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n")
	if err := NewStarterKit(vue3Variant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vue3 install failed: %v", err)
	}
	if !fileExists("resources/js/App.vue") {
		t.Fatal("Expected App.vue after vue3 install")
	}

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vanilla install failed: %v", err)
	}

	if fileExists("resources/js/App.vue") {
		t.Error("Expected App.vue to be removed after switching to vanilla")
	}
	if fileExists("resources/js/main.ts") {
		t.Error("Expected main.ts to be removed after switching to vanilla")
	}
	if !fileExists("resources/js/script.ts") {
		t.Error("Expected script.ts to exist after switching to vanilla")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if strings.Contains(string(pkg), `"vue"`) {
		t.Errorf("Expected package.json NOT to contain '\"vue\"' after switch to vanilla, got: %s", string(pkg))
	}
}

func TestVue3Variant_AllEmbedPathsResolve(t *testing.T) {
	for dest, embedPath := range vue3Variant.base {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vue3Variant.base[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vue3Variant.base[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
	for dest, embedPath := range vue3Variant.noTailwind {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vue3Variant.noTailwind[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vue3Variant.noTailwind[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
}

func TestManagedFiles_IncludesVue3Destinations(t *testing.T) {
	got := managedFiles()
	// 5 shared auth views + 10 shared auth code (incl. Makefile) +
	// 4 shared auth migrations + 4 shared dashboard views + 8 vanilla-unique
	// base entries + resources/js/main.ts + resources/js/App.vue from the vue
	// variants = 33; plus 12 net-new paths from sharedAuthVue3 (router.ts,
	// api.ts, banner.ts, AlertBanner.vue, and 8 page components — main.ts,
	// App.vue, package.json already in the union) = 45. sharedAuthVue3ViewOverride
	// keys are all already present via sharedAuthViews + sharedDashboardViews
	// + the home jet base entry, so they add no new paths.
	if len(got) != 45 {
		t.Errorf("Expected managedFiles() to return 45 paths, got %d: %v", len(got), got)
	}
	if !slices.Contains(got, "resources/js/main.ts") {
		t.Errorf("Expected managedFiles() to include 'resources/js/main.ts', got: %v", got)
	}
	if !slices.Contains(got, "resources/js/App.vue") {
		t.Errorf("Expected managedFiles() to include 'resources/js/App.vue', got: %v", got)
	}
	if !slices.Contains(got, "resources/js/router.ts") {
		t.Errorf("Expected managedFiles() to include 'resources/js/router.ts', got: %v", got)
	}
	if !slices.Contains(got, "resources/js/components/Login.vue") {
		t.Errorf("Expected managedFiles() to include 'resources/js/components/Login.vue', got: %v", got)
	}
}

func TestStarterKit_Handle_TailwindAfterNoTailwind(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, false, false, false).Handle(); err != nil {
		t.Fatalf("First install (tailwind=false) failed: %v", err)
	}
	if fileExists("tailwind.config.js") {
		t.Fatal("Expected tailwind.config.js NOT to exist after tailwind=false install")
	}

	withStdin(t, "yes\n")
	if err := NewStarterKit(vanillaVariant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Second install (tailwind=true) failed: %v", err)
	}

	if !fileExists("tailwind.config.js") {
		t.Error("Expected tailwind.config.js to exist after re-install with tailwind=true")
	}
	if !fileExists("postcss.config.js") {
		t.Error("Expected postcss.config.js to exist after re-install with tailwind=true")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkg), "tailwindcss") {
		t.Errorf("Expected package.json to contain 'tailwindcss' after re-install with tailwind=true, got: %s", string(pkg))
	}
}

var expectedVue2Files = []string{
	"resources/views/layouts/base.jet",
	"resources/views/home.jet",
	"resources/css/styles.css",
	"resources/js/main.ts",
	"resources/js/App.vue",
	"package.json",
	"vite.config.ts",
	"tailwind.config.js",
	"postcss.config.js",
}

func TestStarterKit_Handle_Vue2_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vue2Variant, false, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedVue2Files {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written", f)
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgStr := string(pkg)
	if !strings.Contains(pkgStr, `"vue": "^2`) {
		t.Errorf("Expected package.json to contain '\"vue\": \"^2', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, "@vitejs/plugin-vue2") {
		t.Errorf("Expected package.json to contain '@vitejs/plugin-vue2', got: %s", pkgStr)
	}

	app, err := os.ReadFile("resources/js/App.vue")
	if err != nil {
		t.Fatalf("Failed to read App.vue: %v", err)
	}
	if !strings.Contains(string(app), "Vue.extend(") {
		t.Errorf("Expected App.vue to contain 'Vue.extend(', got: %s", string(app))
	}

	vite, err := os.ReadFile("vite.config.ts")
	if err != nil {
		t.Fatalf("Failed to read vite.config.ts: %v", err)
	}
	if !strings.Contains(string(vite), "@vitejs/plugin-vue2") {
		t.Errorf("Expected vite.config.ts to reference '@vitejs/plugin-vue2', got: %s", string(vite))
	}

	main, err := os.ReadFile("resources/js/main.ts")
	if err != nil {
		t.Fatalf("Failed to read main.ts: %v", err)
	}
	if !strings.Contains(string(main), "new Vue(") {
		t.Errorf("Expected main.ts to contain 'new Vue(', got: %s", string(main))
	}
}

func TestStarterKit_Handle_Vue2_NoTailwind(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vue2Variant, false, false, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	if fileExists("tailwind.config.js") {
		t.Error("Expected tailwind.config.js NOT to be written when tailwind=false")
	}
	if fileExists("postcss.config.js") {
		t.Error("Expected postcss.config.js NOT to be written when tailwind=false")
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgStr := string(pkg)
	if strings.Contains(pkgStr, "tailwindcss") {
		t.Errorf("Expected package.json NOT to contain 'tailwindcss', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, `"vue": "^2`) {
		t.Errorf("Expected package.json to contain '\"vue\": \"^2', got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, "@vitejs/plugin-vue2") {
		t.Errorf("Expected package.json to contain '@vitejs/plugin-vue2', got: %s", pkgStr)
	}
}

func TestStarterKit_Handle_Vue2ToVue3_VariantSwitch(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n")
	if err := NewStarterKit(vue2Variant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vue2 install failed: %v", err)
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json after vue2 install: %v", err)
	}
	if !strings.Contains(string(pkg), `"vue": "^2`) {
		t.Fatalf("Expected vue2 package.json to contain '\"vue\": \"^2', got: %s", string(pkg))
	}

	withStdin(t, "yes\n")
	if err := NewStarterKit(vue3Variant, false, true, false, false).Handle(); err != nil {
		t.Fatalf("Vue3 install failed: %v", err)
	}

	if !fileExists("resources/js/main.ts") {
		t.Error("Expected main.ts to exist after switching to vue3")
	}
	if !fileExists("resources/js/App.vue") {
		t.Error("Expected App.vue to exist after switching to vue3")
	}

	pkg, err = os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json after switch: %v", err)
	}
	pkgStr := string(pkg)
	if !strings.Contains(pkgStr, `"vue": "^3`) {
		t.Errorf("Expected package.json to contain '\"vue\": \"^3' after switch, got: %s", pkgStr)
	}
	if strings.Contains(pkgStr, `"vue": "^2`) {
		t.Errorf("Expected package.json NOT to contain '\"vue\": \"^2' after switch, got: %s", pkgStr)
	}

	vite, err := os.ReadFile("vite.config.ts")
	if err != nil {
		t.Fatalf("Failed to read vite.config.ts: %v", err)
	}
	viteStr := string(vite)
	if strings.Contains(viteStr, "@vitejs/plugin-vue2") {
		t.Errorf("Expected vite.config.ts NOT to reference '@vitejs/plugin-vue2' after switch, got: %s", viteStr)
	}
	if !strings.Contains(viteStr, "@vitejs/plugin-vue") {
		t.Errorf("Expected vite.config.ts to reference '@vitejs/plugin-vue' after switch, got: %s", viteStr)
	}

	app, err := os.ReadFile("resources/js/App.vue")
	if err != nil {
		t.Fatalf("Failed to read App.vue: %v", err)
	}
	appStr := string(app)
	if !strings.Contains(appStr, "<script setup") {
		t.Errorf("Expected App.vue to contain '<script setup' after switch, got: %s", appStr)
	}
	if strings.Contains(appStr, "Vue.extend(") {
		t.Errorf("Expected App.vue NOT to contain 'Vue.extend(' after switch, got: %s", appStr)
	}
}

func TestVue2Variant_AllEmbedPathsResolve(t *testing.T) {
	for dest, embedPath := range vue2Variant.base {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vue2Variant.base[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vue2Variant.base[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
	for dest, embedPath := range vue2Variant.noTailwind {
		data, err := templateFS.ReadFile(embedPath)
		if err != nil {
			t.Errorf("vue2Variant.noTailwind[%q] -> %q: failed to read embedded template: %v", dest, embedPath, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("vue2Variant.noTailwind[%q] -> %q: embedded template is empty", dest, embedPath)
		}
	}
}

// TestStarterKit_StageFiles_SubstitutesAppName verifies that the `$APPNAME$`
// token in embedded templates is replaced with the module name parsed from
// ./go.mod when files are staged. This is the bridge that lets the user-app's
// `$APPNAME$/models` imports resolve to the actual module path.
func TestStarterKit_StageFiles_SubstitutesAppName(t *testing.T) {
	t.Chdir(t.TempDir())

	// Seed a go.mod whose module line is the value we want pasted in. The
	// integration uses the real handlers/aerra.go template which
	// contains the literal `$APPNAME$/models` import.
	goMod := "module myproject/foo\n\ngo 1.25\n"
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	withStdin(t, "")

	// $APPNAME$ tokens only appear inside the aerra auth scaffold, so this
	// test must opt in via withAuth=true; otherwise no aerra files are written.
	k := NewStarterKit(vanillaVariant, false, true, false, true)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	got, err := os.ReadFile("handlers/aerra.go")
	if err != nil {
		t.Fatalf("Failed to read handlers/aerra.go: %v", err)
	}
	gotStr := string(got)

	if strings.Contains(gotStr, "$APPNAME$") {
		t.Errorf("Expected `$APPNAME$` token to be substituted, got: %s", gotStr)
	}
	if !strings.Contains(gotStr, `"myproject/foo/models"`) {
		t.Errorf("Expected import to contain 'myproject/foo/models', got: %s", gotStr)
	}

	// remember.go also has `$APPNAME$/models` — confirm it was substituted there
	// too, since stageFiles iterates the whole map.
	rem, err := os.ReadFile("middleware/remember.go")
	if err != nil {
		t.Fatalf("Failed to read middleware/remember.go: %v", err)
	}
	if strings.Contains(string(rem), "$APPNAME$") {
		t.Errorf("Expected `$APPNAME$` token to be substituted in remember.go, got: %s", string(rem))
	}
	if !strings.Contains(string(rem), `"myproject/foo/models"`) {
		t.Errorf("Expected remember.go import to contain 'myproject/foo/models', got: %s", string(rem))
	}
}

// TestStarterKit_ReadAppName_FromGoMod verifies the helper that pulls the
// module name out of ./go.mod. The starter-kit relies on this to feed the
// `$APPNAME$` substitution; failing here would silently leave the token in the
// installed Go source.
func TestStarterKit_ReadAppName_FromGoMod(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile("go.mod", []byte("module myapp\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	if got := readAppName(); got != "myapp" {
		t.Errorf("readAppName() = %q, want %q", got, "myapp")
	}
}

// TestStarterKit_RemoveOrphans_DeletesOnlyNewDestinations directly exercises
// the orphan-cleanup helper that runs when commitStaged fails partway through.
// Files that pre-existed (in `existing`) must be left in place so rollback can
// restore them; files that did not pre-exist must be removed so a failed
// install does not leak partial state into the project.
func TestStarterKit_RemoveOrphans_DeletesOnlyNewDestinations(t *testing.T) {
	t.Chdir(t.TempDir())

	preExisted := "package.json"
	orphan := "resources/js/App.vue"

	if err := os.WriteFile(preExisted, []byte("ORIG"), 0644); err != nil {
		t.Fatalf("Failed to seed pre-existing: %v", err)
	}
	if err := os.MkdirAll("resources/js", 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(orphan, []byte("NEW"), 0644); err != nil {
		t.Fatalf("Failed to seed orphan: %v", err)
	}

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	k.removeOrphans(
		[]string{preExisted, orphan},
		[]string{preExisted},
	)

	if got, err := os.ReadFile(preExisted); err != nil {
		t.Errorf("Pre-existing file should still exist, got read error: %v", err)
	} else if string(got) != "ORIG" {
		t.Errorf("Pre-existing file content should be untouched, got: %s", string(got))
	}

	if fileExists(orphan) {
		t.Errorf("Orphan %q should have been removed by removeOrphans", orphan)
	}
}

// TestStarterKit_Handle_Vanilla_NoAuthByDefault confirms that a default vanilla
// install (no --with-auth) writes ONLY the starter-kit chrome and leaves the
// aerra auth scaffold off disk entirely. The aerra files are an opt-in extra.
func TestStarterKit_Handle_Vanilla_NoAuthByDefault(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, true, false, false)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedStarterKitFiles {
		if !fileExists(f) {
			t.Errorf("Expected vanilla starter file %q to be written", f)
		}
	}

	for _, f := range expectedAerraAuthFiles {
		if fileExists(f) {
			t.Errorf("Expected aerra auth file %q NOT to be written without --with-auth", f)
		}
	}
}

// TestStarterKit_Handle_VanillaWithAuth_HappyPath confirms that --with-auth on
// vanilla writes both the starter-kit chrome and the full aerra auth scaffold
// (handlers, middleware, data, migrations, dashboard views, replaced
// routes-web.go, and the public auth views).
func TestStarterKit_Handle_VanillaWithAuth_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, true, false, true)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedVanillaWithAuthFiles {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written by --with-auth install", f)
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}
}

// expectedVue3WithAuthFiles is the union of vue3 base files and the aerra
// auth scaffold + the vue3 SPA frontend overrides. The SPA replaces the
// per-page Jet views with a single shell, ships Vue Router + components,
// and adds vue-router to package.json.
var expectedVue3WithAuthFiles = []string{
	"resources/views/layouts/base.jet",
	"resources/views/home.jet",
	"resources/views/layouts/application.jet",
	"resources/views/login.jet",
	"resources/views/registration.jet",
	"resources/views/forgot.jet",
	"resources/views/reset-password.jet",
	"resources/views/dashboard/home.jet",
	"resources/views/dashboard/profile.jet",
	"resources/views/dashboard/header.jet",
	"resources/views/dashboard/menu.jet",
	"resources/css/styles.css",
	"resources/js/main.ts",
	"resources/js/App.vue",
	"resources/js/router.ts",
	"resources/js/api.ts",
	"resources/js/banner.ts",
	"resources/js/components/AppLayout.vue",
	"resources/js/components/AlertBanner.vue",
	"resources/js/components/Login.vue",
	"resources/js/components/Registration.vue",
	"resources/js/components/Forgot.vue",
	"resources/js/components/ResetPassword.vue",
	"resources/js/components/Home.vue",
	"resources/js/components/DashboardHome.vue",
	"resources/js/components/DashboardProfile.vue",
	"package.json",
	"vite.config.ts",
	"tailwind.config.js",
	"postcss.config.js",
	"handlers/aerra.go",
	"handlers/convenience.go",
	"main.go",
	"middleware/authenticated.go",
	"middleware/remember.go",
	"models/user.go",
	"models/remember_token.go",
	"models/models.go",
	"routes-web.go",
	"Makefile",
	"migrations/0001_create_users_table.up.sql",
	"migrations/0001_create_users_table.down.sql",
	"migrations/0002_create_remember_tokens_table.up.sql",
	"migrations/0002_create_remember_tokens_table.down.sql",
}

// TestStarterKit_Handle_Vue3WithAuth_HappyPath confirms that --with-auth on
// vue3 writes the full SPA scaffold: starter-kit chrome, aerra auth backend,
// SPA shell views, Vue Router, components, fetch wrapper, and a package.json
// with vue-router pinned.
func TestStarterKit_Handle_Vue3WithAuth_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vue3Variant, false, true, false, true)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedVue3WithAuthFiles {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written by --vue3 --with-auth install", f)
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}

	main, err := os.ReadFile("resources/js/main.ts")
	if err != nil {
		t.Fatalf("Failed to read main.ts: %v", err)
	}
	mainStr := string(main)
	if !strings.Contains(mainStr, "createApp") {
		t.Errorf("Expected SPA main.ts to contain 'createApp', got: %s", mainStr)
	}
	if !strings.Contains(mainStr, "app.use(router)") {
		t.Errorf("Expected SPA main.ts to wire the router, got: %s", mainStr)
	}

	router, err := os.ReadFile("resources/js/router.ts")
	if err != nil {
		t.Fatalf("Failed to read router.ts: %v", err)
	}
	routerStr := string(router)
	for _, route := range []string{"'/login'", "'/registration'", "'/forgot'", "'/reset-password'", "'/dashboard/home'", "'/dashboard/profile'"} {
		if !strings.Contains(routerStr, route) {
			t.Errorf("Expected router.ts to register route %s, got: %s", route, routerStr)
		}
	}

	loginJet, err := os.ReadFile("resources/views/login.jet")
	if err != nil {
		t.Fatalf("Failed to read login.jet: %v", err)
	}
	if !strings.Contains(string(loginJet), `<div id="app">`) {
		t.Errorf("Expected login.jet to be replaced with SPA shell, got: %s", string(loginJet))
	}

	dashboardHomeJet, err := os.ReadFile("resources/views/dashboard/home.jet")
	if err != nil {
		t.Fatalf("Failed to read dashboard/home.jet: %v", err)
	}
	if !strings.Contains(string(dashboardHomeJet), `<div id="app">`) {
		t.Errorf("Expected dashboard/home.jet to be replaced with SPA shell, got: %s", string(dashboardHomeJet))
	}
	if !strings.Contains(string(dashboardHomeJet), `../layouts/application.jet`) {
		t.Errorf("Expected nested SPA shell to extend ../layouts/application.jet, got: %s", string(dashboardHomeJet))
	}

	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	pkgStr := string(pkg)
	if !strings.Contains(pkgStr, `"vue-router"`) {
		t.Errorf("Expected package.json to include vue-router, got: %s", pkgStr)
	}
	if !strings.Contains(pkgStr, `"vue"`) {
		t.Errorf("Expected package.json to include vue, got: %s", pkgStr)
	}
}

// TestStarterKit_Handle_WithAuth_OverridesStylesAndTailwindConfig confirms that
// --with-auth replaces the base styles.css and tailwind.config.js with the
// aerra-flavored versions. The aerra views require utility classes (.card) and
// palette entries (pink-1000) that don't exist in the basic starter assets, so
// the override map in resolveFiles() must win on key collision.
func TestStarterKit_Handle_WithAuth_OverridesStylesAndTailwindConfig(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "")

	k := NewStarterKit(vanillaVariant, false, true, true, true)
	if err := k.Handle(); err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	styles, err := os.ReadFile("resources/css/styles.css")
	if err != nil {
		t.Fatalf("Failed to read styles.css: %v", err)
	}
	if !strings.Contains(string(styles), ".card") {
		t.Errorf("Expected aerra styles.css to contain '.card' utility class, got: %s", string(styles))
	}

	tw, err := os.ReadFile("tailwind.config.js")
	if err != nil {
		t.Fatalf("Failed to read tailwind.config.js: %v", err)
	}
	if !strings.Contains(string(tw), "pink-1000") {
		t.Errorf("Expected aerra tailwind.config.js to contain 'pink-1000', got: %s", string(tw))
	}
}
