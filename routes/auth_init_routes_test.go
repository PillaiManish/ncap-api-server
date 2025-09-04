package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type fakeAPIServer struct {
	loginCalled    bool
	registerCalled bool
	lastMethod     string
}

func (f *fakeAPIServer) Login(w http.ResponseWriter, r *http.Request) {
	f.loginCalled = true
	f.lastMethod = r.Method
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "login ok")
}

func (f *fakeAPIServer) Register(w http.ResponseWriter, r *http.Request) {
	f.registerCalled = true
	f.lastMethod = r.Method
	w.WriteHeader(http.StatusCreated)
	_, _ = io.WriteString(w, "register ok")
}

// routesApp minimal test shim: depend on fields used by initAuthRoutes.
// If the production routesApp has more fields, embedding them here is harmless for tests.
type routesApp struct {
	router    *mux.Router
	apiServer interface {
		Login(http.ResponseWriter, *http.Request)
		Register(http.ResponseWriter, *http.Request)
	}
}

func TestInitAuthRoutes_registersPostHandlers(t *testing.T) {
	fake := &fakeAPIServer{}
	app := &routesApp{
		router:    mux.NewRouter(),
		apiServer: fake,
	}
	app.initAuthRoutes()

	// POST /login
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /login status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, "login ok") {
		t.Fatalf("POST /login body = %q, want contains %q", body, "login ok")
	}
	if !fake.loginCalled || fake.lastMethod != http.MethodPost {
		t.Fatalf("login handler not called correctly: called=%v method=%q", fake.loginCalled, fake.lastMethod)
	}

	// POST /register
	req2 := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{}`))
	rec2 := httptest.NewRecorder()
	app.router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusCreated {
		t.Fatalf("POST /register status = %d, want %d", rec2.Code, http.StatusCreated)
	}
	if body := rec2.Body.String(); !strings.Contains(body, "register ok") {
		t.Fatalf("POST /register body = %q, want contains %q", body, "register ok")
	}
	if !fake.registerCalled || fake.lastMethod != http.MethodPost {
		t.Fatalf("register handler not called correctly: called=%v method=%q", fake.registerCalled, fake.lastMethod)
	}
}

func TestInitAuthRoutes_rejectsNonPostMethods(t *testing.T) {
	fake := &fakeAPIServer{}
	app := &routesApp{
		router:    mux.NewRouter(),
		apiServer: fake,
	}
	app.initAuthRoutes()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET login not allowed", http.MethodGet, "/login"},
		{"PUT login not allowed", http.MethodPut, "/login"},
		{"DELETE login not allowed", http.MethodDelete, "/login"},
		{"GET register not allowed", http.MethodGet, "/register"},
		{"PUT register not allowed", http.MethodPut, "/register"},
		{"PATCH register not allowed", http.MethodPatch, "/register"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			app.router.ServeHTTP(rec, req)

			// mux.Router returns 405 for wrong method on an existing path with method constraints.
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("%s: status = %d, want %d", tc.name, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestInitAuthRoutes_unknownPathReturns404(t *testing.T) {
	fake := &fakeAPIServer{}
	app := &routesApp{
		router:    mux.NewRouter(),
		apiServer: fake,
	}
	app.initAuthRoutes()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	app.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /does-not-exist status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}