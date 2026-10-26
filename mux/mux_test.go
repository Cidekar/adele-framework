package mux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// helper methods for mux testing
func testHandler(t *testing.T, h http.Handler, method, path string, body io.Reader) (*http.Response, string) {
	r, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result(), w.Body.String()
}

func TestMux_New(t *testing.T) {
	mux := NewRouter()

	if reflect.TypeOf(mux).String() != "*mux.Mux" {
		t.Error("mux new did not return expected type")
	}
}

func TestMux_GetScopes(t *testing.T) {
	scope := "ping pong"
	path := "/ping"
	annotation := "scopes[ping pong]"

	MuxRouterTree = append(MuxRouterTree, MuxRouteInfo{
		Route:      path,
		Annotation: annotation,
		Scope:      scope,
	})

	mux := NewRouter()

	muxRouteScopes := mux.GetScopes(path)

	if scope != strings.Join(muxRouteScopes.Scope, " ") {
		t.Error("scope not found on expected path")
	}
}

func TestMux_With(t *testing.T) {

	mf := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}

	mux := NewRouter()

	r := mux.With(mf)

	if reflect.TypeOf(r).String() != "*chi.Mux" {
		t.Error("mux with did not return pointer to internal router")
	}

}

func TestMux_Use(t *testing.T) {

	mf := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	}

	mux := NewRouter()

	mux.Use(mf)

	if len(mux.Mux.Middlewares()) != 1 {
		t.Error("mux use did not return expected number of middleware")
	}
}

func TestMux_Handle(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Handle(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p {
			added = true
		}
	}

	if added == false {
		t.Error("handle did not add unique handler to mux")
	}
}

// HandleFunc
func TestMux_HandleFunc(t *testing.T) {

	p := "/foo/bar/bat"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.HandleFunc(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p {
			added = true
		}
	}

	if added == false {
		t.Error("handle did not add unique handler to mux")
	}
}

func TestMux_Match(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Handle(p, mf)

	ctx := chi.Context{}

	method := "GET"

	found := mux.Match(&ctx, method, p)

	if found == false {
		t.Error("mux wan not able to match the method and path")
	}
}

func TestMux_Method(t *testing.T) {

	p := "/foo/bar/cax"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Method("GET", p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p {
			added = true
		}
	}

	if added == false {
		t.Error("mux method did not add the route pattern and method")
	}
}

func TestMux_MethodFunc(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.MethodFunc("GET", p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p {
			added = true
		}
	}

	if added == false {
		t.Error("mux method function did not add the route pattern and method")
	}
}

func TestMux_Connect(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Connect(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["CONNECT"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux connect function did not add the route pattern and method")
	}
}

func TestMux_Head(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Head(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["HEAD"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux head function did not add the route pattern and method")
	}
}

func TestMux_Get(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Get(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["GET"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux get function did not add the route pattern and method")
	}
}

func TestMux_Post(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Post(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["POST"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux post function did not add the route pattern and method")
	}
}

func TestMux_Put(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Put(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["PUT"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux put function did not add the route pattern and method")
	}
}

func TestMux_Patch(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Patch(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["PATCH"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux patch function did not add the route pattern and method")
	}
}

func TestMux_Delete(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Delete(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["DELETE"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux delete function did not add the route pattern and method")
	}
}

func TestMux_Trace(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Trace(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["TRACE"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux trace function did not add the route pattern and method")
	}
}

func TestMux_Options(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Options(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["OPTIONS"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux trace function did not add the route pattern and method")
	}
}

func TestMux_NotFound(t *testing.T) {

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Adele", 404)
		return
	})

	mux := NewRouter()

	mux.Get("/adele", mf)

	mux.NotFound(mf)

	res, body := testHandler(t, mux, "GET", "/unknown", nil)

	if res.StatusCode != 404 && body != "Adele" {
		t.Error("mux not found function did not apply the not found handler")
	}
}

func TestMux_MethodNotAllowed(t *testing.T) {

	status := http.StatusText(405) + " Adele"
	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, status, 404)
		return
	})

	mux := NewRouter()

	mux.Get("/adele", mf)

	mux.MethodNotAllowed(mf)

	res, body := testHandler(t, mux, "GET", "/unknown", nil)

	if res.StatusCode != 404 && body != status {
		t.Error("mux not found function did not apply the not found handler")
	}
}

func TestMux_Group(t *testing.T) {

	status := "Adele"
	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(status))
	})

	mux := NewRouter()

	mux.Group(func(r chi.Router) {
		mux.Get("/adele", mf)
	})

	res, _ := testHandler(t, mux, "GET", "/adele", nil)

	if res.StatusCode != 200 {
		t.Error("mux group function did create a new inline-mux with a fresh handler")
	}
}

func TestMux_Route(t *testing.T) {

	pattern := "/adele"

	mux := NewRouter()

	mux.Route(pattern, func(r chi.Router) {

		mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		mux.Get("/", mf)
	})

	added := false
	for _, route := range mux.Routes() {
		if route.Pattern == pattern+"/*" {
			added = true
		}
	}

	if added == false {
		t.Error("mux trace function did not add the route pattern and method")
	}
}

func TestMux_Mount(t *testing.T) {

	pattern := "/adele"

	mux := NewRouter()

	mux.Route(pattern, func(r chi.Router) {

		mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		mux.Get("/", mf)
	})

	m := NewRouter()

	m.Mount("/share", mux)

	added := false
	for _, route := range mux.Routes() {
		if route.Pattern == pattern+"/*" {
			added = true
		}
	}

	if added == false {
		t.Error("mux trace function did not add the route pattern and method")
	}
}

func TestMux_Middlewares(t *testing.T) {

	mux := NewRouter()

	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	})

	if len(mux.Middlewares()) != 1 {
		t.Error("mux use did not return expected number of middleware")
	}
}

func TestMux_Routes(t *testing.T) {

	pattern := "/foo/bar/baz"

	mux := NewRouter()

	mux.Get(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	if len(mux.Routes()) != 1 {
		t.Error("mux routes function did return all middleware")
	}
}

func TestMux_ServeHTTP(t *testing.T) {

	pattern := "/adele"

	mux := NewRouter()

	mux.Get(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r, _ := http.NewRequest("GET", pattern, nil)

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	if w.Result().StatusCode != 200 {
		t.Error("mux server http function did not properly apply the routing context")
	}
}

func TestMux_CheckAnnotation(t *testing.T) {

	mux := NewRouter()

	pattern := "/adele"

	annotation := "[scopes:Foo Bar]"

	mux.Post(pattern+annotation, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r, _ := http.NewRequest("POST", pattern, nil)

	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	t.Log(w.Result().StatusCode)

}

func TestMux_Query(t *testing.T) {

	p := "/foo/bar/baz"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux := NewRouter()

	mux.Query(p, mf)

	added := false
	for _, route := range mux.Routes() {

		if route.Pattern == p && route.Handlers["QUERY"] != nil {
			added = true
		}
	}

	if added == false {
		t.Error("mux query function did not add the route pattern and method")
	}
}

func TestMux_Query_Annotation(t *testing.T) {

	mux := NewRouter()

	pattern := "/adele"

	annotation := "[scopes:Foo Bar]"

	mf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	mux.Query(pattern+annotation, mf)

	clean := false
	raw := false
	for _, route := range mux.Routes() {

		if route.Pattern == pattern && route.Handlers["QUERY"] != nil {
			clean = true
		}

		if route.Pattern == pattern+annotation {
			raw = true
		}
	}

	if clean == false {
		t.Error("mux query function did not add the route under the clean pattern")
	}

	if raw == true {
		t.Error("mux query function did not strip the scope annotation from the pattern")
	}

	recorded := false
	for _, router := range MuxRouterTree {

		if router.Route == pattern && router.Annotation == pattern+annotation {
			recorded = true
		}
	}

	if recorded == false {
		t.Error("mux query function did not record the scope annotation for the route")
	}
}
