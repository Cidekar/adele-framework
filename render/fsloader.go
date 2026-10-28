package render

import (
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/CloudyKit/jet/v6"
)

// fsLoader adapts any fs.FS to the jet.Loader interface. Accepting the fs.FS
// interface (not the concrete embed.FS) is the strictly more flexible seam:
// embed.FS, os.DirFS, fstest.MapFS, and any custom fs.FS all satisfy it, so one
// loader serves both embedded and runtime-disk templates.
type fsLoader struct {
	fsys fs.FS
	dir  string
}

// resolve turns the name jet asks for into an fs.FS key.
//
// jet builds an absolute slash path (e.g. "/issue.md.jet") before a lookup. It
// uses path.Join and fs.Stat/fsys.Open, NEVER filepath.*: fs.FS and embed.FS
// require forward-slash keys always, but filepath.Join/FromSlash produce
// backslashes on Windows, which breaks the lookup. jet's own embedfs loader has
// this Windows bug; this loader is deliberately slash-only and correct.
//
// path.Join("", "/issue.md.jet") keeps the leading slash, but embed.FS and
// fstest.MapFS reject keys with a leading slash. The TrimPrefix guard strips it
// so empty-dir callers resolve. (When dir is non-empty, path.Join already
// strips the leading slash, so the guard is a no-op.)
//
// dir is a hard floor: an absolute ".." name (e.g. "/../secret.md.jet") must not
// climb to a sibling of dir inside fsys. When the join escapes dir, fall back to
// dir + base name so the key stays within dir.
func (l *fsLoader) resolve(name string) string {
	key := path.Join(l.dir, name)
	if l.dir != "" && key != l.dir && !strings.HasPrefix(key, l.dir+"/") {
		key = path.Join(l.dir, path.Base(name))
	}
	return strings.TrimPrefix(key, "/")
}

// Open returns the template's contents. jet closes the reader.
func (l *fsLoader) Open(name string) (io.ReadCloser, error) {
	return l.fsys.Open(l.resolve(name))
}

// Exists reports whether a (non-directory) template exists under name.
//
// It must be total and non-panicking: jet probes several candidate names per
// lookup (it loops its extension list and calls Exists(name+ext) for each), and
// only calls Open on the first Exists-true hit. A missing file yields a non-nil
// fs.Stat error, which this returns as false; Exists never errors or panics on a
// non-existent name.
func (l *fsLoader) Exists(name string) bool {
	st, err := fs.Stat(l.fsys, l.resolve(name))
	return err == nil && !st.IsDir()
}

// NewFSLoader returns a jet.Loader that loads templates from fsys, rooted at dir.
//
// Pass dir "" to load from the root of fsys, or a subdirectory (e.g. "templates")
// to root lookups there. fsys may be any fs.FS: an embed.FS from //go:embed,
// os.DirFS for a runtime directory, fstest.MapFS for tests, or a custom backing.
//
// Template names are caller-supplied and trusted; they MUST NOT be
// end-user-controlled, because an absolute name containing ".." can resolve to a
// sibling file inside fsys (it cannot escape the FS root, but it can reach a
// sibling of dir).
func NewFSLoader(fsys fs.FS, dir string) jet.Loader {
	return &fsLoader{fsys: fsys, dir: dir}
}
