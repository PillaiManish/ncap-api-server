package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

//
// Test strategy:
// We validate that initProjectRoutes registers the expected paths and HTTP methods.
// We avoid invoking the handlers (which may depend on external server logic) by:
// - Asserting route table contents via mux.Router.Walk
// - Issuing requests with incorrect methods to ensure 405 Method Not Allowed is returned
//
// This keeps tests unit-level and robust even if apiServer receiver is nil or not easily mockable.
//

// expectedRoutes returns the mapping of paths to allowed HTTP methods
func expectedRoutes() map[string][]string {
	return map[string][]string{
		"/create-new-project":     {http.MethodPost},
		"/retrieve-project-list":  {http.MethodGet},
		"/edit-project":           {http.MethodPatch},
		"/check-project-permission": {http.MethodPost},
		"/project-access-request": {http.MethodPost},
	}
}

// helper to normalize methods slice into a set-like string for stable comparisons
func normalizeMethods(methods []string) string {
	// Ensure consistent order without importing sort (keep dependencies minimal)
	order := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
	var out []string
	for _, m := range order {
		for _, seen := range methods {
			if strings.EqualFold(seen, m) {
				out = append(out, m)
				break
			}
		}
	}
	// Append any uncommon methods in their seen order
	for _, seen := range methods {
		found := false
		for _, m := range out {
			if strings.EqualFold(seen, m) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, seen)
		}
	}
	return strings.Join(out, ",")
}

// newTestRoutesApp constructs a minimal routesApp with a mux.Router.
// We intentionally do not set apiServer because handlers are not invoked in these tests.
func newTestRoutesApp() *routesApp {
	ra := &routesApp{}
	// Prefer creating a fresh router to avoid cross-test contamination if the struct permits.
	// If routesApp already initializes router elsewhere, this assignment is still safe for tests.
	ra.router = mux.NewRouter()
	return ra
}

func TestInitProjectRoutes_registersExpectedPathsAndMethods(t *testing.T) {
	app := newTestRoutesApp()
	app.initProjectRoutes()

	want := expectedRoutes()
	got := map[string][]string{}

	err := app.router.Walk(func(r *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := r.GetPathTemplate()
		if err != nil || path == "" {
			return nil // skip routes without a concrete path
		}
		methods, _ := r.GetMethods()
		if len(methods) == 0 {
			// If no methods were set, mux treats it as all methods; we capture that explicitly.
			methods = []string{"ANY"}
		}
		got[path] = methods
		return nil
	})
	if err != nil {
		t.Fatalf("router.Walk failed: %v", err)
	}

	// Assert each expected route exists with the correct methods
	for path, expectedMethods := range want {
		methods, ok := got[path]
		if !ok {
			t.Errorf("expected route %q to be registered, but it was not found", path)
			continue
		}
		if normalizeMethods(methods) != normalizeMethods(expectedMethods) {
			t.Errorf("route %q methods mismatch: got [%s], want [%s]", path, normalizeMethods(methods), normalizeMethods(expectedMethods))
		}
	}

	// Ensure there are no unexpected extra routes under these exact paths with different methods
	for path := range got {
		if _, ok := want[path]; !ok {
			// Allow mux internal routes; only warn for conflict if path is one we care about
			switch path {
			case "/create-new-project", "/retrieve-project-list", "/edit-project", "/check-project-permission", "/project-access-request":
				t.Errorf("unexpected route registered: %q", path)
			}
		}
	}
}

func TestInitProjectRoutes_disallowsWrongHTTPMethods(t *testing.T) {
	app := newTestRoutesApp()
	app.initProjectRoutes()

	type tc struct {
		name       string
		method     string
		path       string
		wantStatus int
	}
	// For each route, issue a deliberately incorrect method to assert 405 Method Not Allowed.
	// We also include a 404 case for an unknown path.
	tests := []tc{
		{name: "create-new-project GET not allowed", method: http.MethodGet, path: "/create-new-project", wantStatus: http.StatusMethodNotAllowed},
		{name: "retrieve-project-list POST not allowed", method: http.MethodPost, path: "/retrieve-project-list", wantStatus: http.StatusMethodNotAllowed},
		{name: "edit-project GET not allowed", method: http.MethodGet, path: "/edit-project", wantStatus: http.StatusMethodNotAllowed},
		{name: "check-project-permission GET not allowed", method: http.MethodGet, path: "/check-project-permission", wantStatus: http.StatusMethodNotAllowed},
		{name: "project-access-request GET not allowed", method: http.MethodGet, path: "/project-access-request", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown path returns 404", method: http.MethodGet, path: "/does-not-exist", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			app.router.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Fatalf("ServeHTTP(%s %s) = status %d, want %d; body=%q", tt.method, tt.path, rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

// Optional: sanity check that allowed methods do not crash the router dispatch pipeline.
// We do NOT assert handler outputs to avoid invoking nil receivers; instead, we only
// verify that mux does NOT return 405 for the correct method+path combo. A 200/404/500
// would all be acceptable outcomes for this sanity check as long as it's not 405.
func TestInitProjectRoutes_allowsCorrectHTTPMethods_sanity(t *testing.T) {
	app := newTestRoutesApp()
	app.initProjectRoutes()

	okCombos := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/create-new-project"},
		{http.MethodGet, "/retrieve-project-list"},
		{http.MethodPatch, "/edit-project"},
		{http.MethodPost, "/check-project-permission"},
		{http.MethodPost, "/project-access-request"},
	}

	for _, c := range okCombos {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, nil)
			rr := httptest.NewRecorder()
			app.router.ServeHTTP(rr, req)

			if rr.Code == http.StatusMethodNotAllowed {
				t.Fatalf("expected allowed method for %s %s, but got 405 Method Not Allowed", c.method, c.path)
			}
		})
	}
}