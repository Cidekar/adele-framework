package render

import (
	"embed"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

//go:embed testdata/embedded/*
var embeddedViews embed.FS

// issueData is an arbitrary struct (NOT *TemplateData). It proves RenderToString
// accepts any data without the *TemplateData assertion that JetPage panics on.
type issueData struct {
	Title string
	Body  string
}

// TestRenderToString_OSLoader renders against the OS-loader set with a markdown
// fixture and asserts the exact string with data interpolation.
func TestRenderToString_OSLoader(t *testing.T) {
	got, err := RenderToString(views, "issue.md", issueData{Title: "Bug report", Body: "It broke."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Bug report\n\nIt broke.\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestRenderToString_MapFS renders against an in-memory fstest.MapFS set with no
// file on disk, via NewJetSetFromFS.
func TestRenderToString_MapFS(t *testing.T) {
	fsys := fstest.MapFS{
		"views/issue.md.jet": &fstest.MapFile{Data: []byte("# {{ .Title }}\n\n{{ .Body }}\n")},
	}
	set := NewJetSetFromFS(fsys, "views")

	got, err := RenderToString(set, "issue.md", issueData{Title: "In memory", Body: "No disk."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# In memory\n\nNo disk.\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestRenderToString_EmbedFS renders against a real //go:embed FS, proving the
// directive wiring works as an app would write it.
func TestRenderToString_EmbedFS(t *testing.T) {
	set := NewJetSetFromFS(embeddedViews, "testdata/embedded")

	type nameData struct{ Name string }
	got, err := RenderToString(set, "greeting.md", nameData{Name: "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello, World!" {
		t.Fatalf("got %q, want %q", got, "Hello, World!")
	}
}

// TestRenderToString_EmptyDirLeadingSlash proves the TrimPrefix guard: a MapFS
// key with no leading slash and an empty dir must resolve. Without the guard
// path.Join("", "/issue.md.jet") keeps the leading slash and the lookup fails.
func TestRenderToString_EmptyDirLeadingSlash(t *testing.T) {
	fsys := fstest.MapFS{
		"issue.md.jet": &fstest.MapFile{Data: []byte("# {{ .Title }}\n")},
	}
	set := NewJetSetFromFS(fsys, "")

	got, err := RenderToString(set, "issue.md", issueData{Title: "Root"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "# Root\n" {
		t.Fatalf("got %q, want %q", got, "# Root\n")
	}
}

// TestRenderToString_OSDirFS proves the runtime-disk backing resolves through the
// same loader as embedded.
func TestRenderToString_OSDirFS(t *testing.T) {
	set := NewJetSetFromFS(os.DirFS("./testdata/views"), "")

	got, err := RenderToString(set, "issue.md", issueData{Title: "On disk", Body: "Live read."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# On disk\n\nLive read.\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestRenderToString_UnknownTemplate asserts the error path returns ("", err).
func TestRenderToString_UnknownTemplate(t *testing.T) {
	got, err := RenderToString(views, "does-not-exist.md", nil)
	if err == nil {
		t.Fatal("expected an error for an unknown template, got nil")
	}
	if got != "" {
		t.Fatalf("expected empty string on error, got %q", got)
	}
}

// TestRenderToString_NilSet asserts a nil *jet.Set returns ("", err) and does not
// panic.
func TestRenderToString_NilSet(t *testing.T) {
	got, err := RenderToString(nil, "issue.md", nil)
	if err == nil {
		t.Fatal("expected an error for a nil set, got nil")
	}
	if got != "" {
		t.Fatalf("expected empty string on error, got %q", got)
	}
}

// TestRenderToString_TopLevelAssignment proves a top-level {{ x := ... }} renders
// without panic, locking the make(jet.VarMap) choice. A nil VarMap would panic.
func TestRenderToString_TopLevelAssignment(t *testing.T) {
	got, err := RenderToString(views, "toplevel", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(got) != "ok" {
		t.Fatalf("got %q, want %q", got, "ok")
	}
}

// TestRenderToString_ArbitraryStruct proves an arbitrary struct (not
// *TemplateData) renders without panic, the JetPage failure mode.
func TestRenderToString_ArbitraryStruct(t *testing.T) {
	type payload struct {
		Title string
		Body  string
	}

	got, err := RenderToString(views, "issue.md", payload{Title: "Arbitrary", Body: "struct."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "# Arbitrary\n\nstruct.\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
