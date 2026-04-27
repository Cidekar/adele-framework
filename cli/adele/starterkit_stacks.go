package main

import (
	"maps"
	"slices"
)

// tailwindOnlyDestinations are files that only get written when Tailwind is
// enabled. Every variant treats these the same way, so we declare them once
// and reference the shared map from each variant's tailwindOnly field.
var tailwindOnlyDestinations = map[string]struct{}{
	"tailwind.config.js": {},
	"postcss.config.js":  {},
}

// sharedNoTailwindStyles is the no-tailwind styles.css used by every variant.
var sharedNoTailwindStyles = map[string]string{
	"resources/css/styles.css": "templates/vanilla/styles.css.notailwind",
}

// sharedAuthViews are the public auth views (and their layout) ported from
// aerra. Every variant ships them verbatim; they reuse the starter-kit's
// CSS/JS via {{VITE_ASSET}} so no per-variant overrides are needed.
var sharedAuthViews = map[string]string{
	"resources/views/layouts/application.jet": "templates/aerra/views/layouts/application.jet",
	"resources/views/login.jet":               "templates/aerra/views/login.jet",
	"resources/views/registration.jet":        "templates/aerra/views/registration.jet",
	"resources/views/forgot.jet":              "templates/aerra/views/forgot.jet",
	"resources/views/reset-password.jet":      "templates/aerra/views/reset-password.jet",
}

// sharedAuthCode are the Go source files (handlers, middleware, data models,
// routes) ported from aerra. They sit alongside the user-app skeleton's
// existing handlers/handlers.go and middleware/middleware.go and rely on the
// `$APPNAME$` token being substituted at install time for `$APPNAME$/models`
// imports.
var sharedAuthCode = map[string]string{
	"handlers/aerra.go":           "templates/aerra/code/handlers/aerra.go",
	"handlers/convenience.go":     "templates/aerra/code/handlers/convenience.go",
	"main.go":                     "templates/aerra/code/main.go",
	"middleware/authenticated.go": "templates/aerra/code/middleware/authenticated.go",
	"middleware/remember.go":      "templates/aerra/code/middleware/remember.go",
	"models/user.go":              "templates/aerra/code/models/user.go",
	"models/remember_token.go":    "templates/aerra/code/models/remember_token.go",
	"models/models.go":            "templates/aerra/code/models/models.go",
	"routes-web.go":               "templates/aerra/code/routes-web.go",
	"Makefile":                    "templates/aerra/code/Makefile",
}

// sharedAuthMigrations are the postgres migrations for the users and
// remember_tokens tables backing the auth flow. golang-migrate naming.
var sharedAuthMigrations = map[string]string{
	"migrations/0001_create_users_table.up.sql":             "templates/aerra/migrations/0001_create_users_table.up.sql",
	"migrations/0001_create_users_table.down.sql":           "templates/aerra/migrations/0001_create_users_table.down.sql",
	"migrations/0002_create_remember_tokens_table.up.sql":   "templates/aerra/migrations/0002_create_remember_tokens_table.up.sql",
	"migrations/0002_create_remember_tokens_table.down.sql": "templates/aerra/migrations/0002_create_remember_tokens_table.down.sql",
}

// sharedDashboardViews are the post-login dashboard partials (home, profile,
// header, menu). The OAuth-related views were intentionally dropped for v1.
var sharedDashboardViews = map[string]string{
	"resources/views/dashboard/home.jet":    "templates/aerra/views/dashboard/home.jet",
	"resources/views/dashboard/profile.jet": "templates/aerra/views/dashboard/profile.jet",
	"resources/views/dashboard/header.jet":  "templates/aerra/views/dashboard/header.jet",
	"resources/views/dashboard/menu.jet":    "templates/aerra/views/dashboard/menu.jet",
}

// sharedAuthOverrides replaces base entries that the aerra views can't render
// against. When --with-auth is set, the aerra views need a richer Tailwind
// palette (pink-50..pink-1250, teal-50..teal-950) and custom utility classes
// (.card, .btn, .advatar, .alert, .centered-vh) than the basic starter ships.
// These overrides replace the matching base entries with aerra-flavored
// versions; merged last in resolveFiles() so the colliding keys win.
var sharedAuthOverrides = map[string]string{
	"resources/css/styles.css": "templates/aerra/css/styles.css",
	"tailwind.config.js":       "templates/aerra/tailwind.config.js",
}

// mergeMaps returns a new map containing every key from each input. Later inputs
// override earlier ones on key collision; here all callers pass disjoint keys.
func mergeMaps(sources ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, src := range sources {
		maps.Copy(out, src)
	}
	return out
}

// stackVariant describes a starter-kit variant. Each variant resolves to a flat
// destination -> embed-path map at install time.
type stackVariant struct {
	name string
	// base is the file map written when tailwind is enabled (the default).
	base map[string]string
	// noTailwind overrides selected base entries when --no-tailwind is set.
	// Keys present here REPLACE the same key in base. Keys in base whose value
	// refers to a tailwind-only embed path should also be REMOVED for the
	// no-tailwind run; this is handled by listing them in tailwindOnly.
	noTailwind map[string]string
	// tailwindOnly is the set of destination paths that ONLY exist in the
	// tailwind variant (e.g. tailwind.config.js, postcss.config.js). When
	// --no-tailwind is set, these destinations are dropped from the merged map.
	tailwindOnly map[string]struct{}
}

// The aerra shared maps (sharedAuthViews, sharedAuthCode, sharedAuthMigrations,
// sharedDashboardViews) are intentionally NOT merged into any variant's base.
// They are gated behind --with-auth and merged in at resolve time only when the
// caller opts in (currently vanilla-only; vue/vue2 support is planned).
var vanillaVariant = stackVariant{
	name: "vanilla",
	base: map[string]string{
		"resources/views/layouts/base.jet": "templates/vanilla/base.jet",
		"resources/views/home.jet":         "templates/vanilla/home.jet",
		"resources/css/styles.css":         "templates/vanilla/styles.css",
		"resources/js/script.ts":           "templates/vanilla/script.ts",
		"package.json":                     "templates/vanilla/package.json",
		"vite.config.ts":                   "templates/vanilla/vite.config.ts",
		"tailwind.config.js":               "templates/tailwind/tailwind.config.js",
		"postcss.config.js":                "templates/tailwind/postcss.config.js",
	},
	noTailwind: mergeMaps(sharedNoTailwindStyles, map[string]string{
		"package.json": "templates/vanilla/package.json.notailwind",
	}),
	tailwindOnly: tailwindOnlyDestinations,
}

var vue3Variant = stackVariant{
	name: "vue3",
	base: map[string]string{
		"resources/views/layouts/base.jet": "templates/vue/base.jet",
		"resources/views/home.jet":         "templates/vue/home.jet",
		"resources/css/styles.css":         "templates/vanilla/styles.css",
		"resources/js/main.ts":             "templates/vue3/main.ts",
		"resources/js/App.vue":             "templates/vue/App.vue",
		"package.json":                     "templates/vue3/package.json",
		"vite.config.ts":                   "templates/vue3/vite.config.ts",
		"tailwind.config.js":               "templates/tailwind/tailwind.config.js",
		"postcss.config.js":                "templates/tailwind/postcss.config.js",
	},
	noTailwind: mergeMaps(sharedNoTailwindStyles, map[string]string{
		"package.json": "templates/vue3/package.json.notailwind",
	}),
	tailwindOnly: tailwindOnlyDestinations,
}

var vue2Variant = stackVariant{
	name: "vue2",
	base: map[string]string{
		"resources/views/layouts/base.jet": "templates/vue/base.jet",
		"resources/views/home.jet":         "templates/vue/home.jet",
		"resources/css/styles.css":         "templates/vanilla/styles.css",
		"resources/js/main.ts":             "templates/vue2/main.ts",
		"resources/js/App.vue":             "templates/vue2/App.vue",
		"package.json":                     "templates/vue2/package.json",
		"vite.config.ts":                   "templates/vue2/vite.config.ts",
		"tailwind.config.js":               "templates/tailwind/tailwind.config.js",
		"postcss.config.js":                "templates/tailwind/postcss.config.js",
	},
	noTailwind: mergeMaps(sharedNoTailwindStyles, map[string]string{
		"package.json": "templates/vue2/package.json.notailwind",
	}),
	tailwindOnly: tailwindOnlyDestinations,
}

var allVariants = []stackVariant{vanillaVariant, vue3Variant, vue2Variant}

// managedFiles returns the deduplicated union of every destination path across
// every variant. It is the authoritative list of files the install command
// considers "owned" so the cleanup step always removes the full superset before
// re-staging, regardless of which variant is being installed. The union spans
// both base and noTailwind so a re-run with a different toggle still sweeps
// files written by a previous run.
func managedFiles() []string {
	set := map[string]struct{}{}
	for _, v := range allVariants {
		for dest := range v.base {
			set[dest] = struct{}{}
		}
		for dest := range v.noTailwind {
			set[dest] = struct{}{}
		}
	}
	// Aerra extras are gated by --with-auth at install time, but cleanup still
	// owns these paths — re-running install without --with-auth must sweep any
	// prior aerra install away.
	for dest := range sharedAuthViews {
		set[dest] = struct{}{}
	}
	for dest := range sharedAuthCode {
		set[dest] = struct{}{}
	}
	for dest := range sharedAuthMigrations {
		set[dest] = struct{}{}
	}
	for dest := range sharedDashboardViews {
		set[dest] = struct{}{}
	}
	out := slices.Collect(maps.Keys(set))
	slices.Sort(out)
	return out
}
