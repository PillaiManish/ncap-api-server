package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ncap-api-server/common"
	"ncap-api-server/model"
)

type fakeLogger struct{}
func (f fakeLogger) Debugf(string, ...interface{}) {}
func (f fakeLogger) Errorf(string, ...interface{}) {}

type fakeDB struct {
	createStatus int
	createErr    error
	editStatus   int
	editErr      error
	listProjects []model.ProjectModel
	listStatus   int
	listErr      error
	accessErr    error

	// call trackers
	createCalled bool
	editCalled   bool
	listCalled   bool
	accessCalled bool

	lastCreate *model.ProjectModel
	lastEdit   *model.ProjectModel
	lastUser   string
	lastProjID string
	lastLimit  int
}

func (f *fakeDB) CreateNewProject(ctx context.Context, p *model.ProjectModel, limit int) (int, error) {
	f.createCalled = true
	f.lastCreate = p
	f.lastLimit = limit
	return f.createStatus, f.createErr
}
func (f *fakeDB) EditProject(ctx context.Context, p *model.ProjectModel) (int, error) {
	f.editCalled = true
	f.lastEdit = p
	return f.editStatus, f.editErr
}
func (f *fakeDB) RetrieveProjectList(ctx context.Context, userId string) ([]model.ProjectModel, int, error) {
	f.listCalled = true
	f.lastUser = userId
	return f.listProjects, f.listStatus, f.listErr
}
func (f *fakeDB) ProjectAccessRequest(ctx context.Context, userId, projectId string) error {
	f.accessCalled = true
	f.lastUser = userId
	f.lastProjID = projectId
	return f.accessErr
}

type fakeResponder struct {
	status int
	body   []byte
}
func (fr *fakeResponder) Response(w http.ResponseWriter, code int, payload interface{}) {
	fr.status = code
	switch v := payload.(type) {
	case nil:
		// no body
	case string:
		io.WriteString(w, v)
	case []byte:
		w.Write(v)
	default:
		enc := json.NewEncoder(w)
		_ = enc.Encode(v)
	}
	w.WriteHeader(code)
}

// testServer embeds APIServerModel to reuse methods, but overrides only the
// methods used by handlers that are hard to control (authorization/checkPermission/Response).
type testServer struct {
	APIServerModel
	responder *fakeResponder
	authUser  string
	authErr   error
	permOK    bool
	permErr   error
}
func (ts *testServer) authorization(w http.ResponseWriter, token, endpoint string) (string, error) {
	return ts.authUser, ts.authErr
}
func (ts *testServer) checkPermission(userId, projectId string, role common.Role) (bool, error) {
	return ts.permOK, ts.permErr
}
func (ts *testServer) Response(w http.ResponseWriter, code int, payload interface{}) {
	ts.responder.Response(w, code, payload)
}

func newTestServer(db *fakeDB) *testServer {
	ts := &testServer{
		APIServerModel: APIServerModel{
			ctx:     context.Background(),
			database: db,
			config: &ServerConfig{
				ProjectConfig: ProjectConfig{
					Limit: 5,
				},
			},
			logger: fakeLogger{},
		},
		responder: &fakeResponder{},
		authUser:  "user-123",
	}
	return ts
}

func TestCreateNewProject_Success(t *testing.T) {
	db := &fakeDB{createStatus: http.StatusCreated}
	ts := newTestServer(db)

	body := model.ProjectModel{Name: "proj", Desc: "desc"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer ok")
	w := httptest.NewRecorder()

	ts.CreateNewProject(w, req)

	if !db.createCalled {
		t.Fatalf("expected CreateNewProject to be called")
	}
	if db.lastCreate.UserId != ts.authUser {
		t.Fatalf("expected UserId %q to be set on project, got %q", ts.authUser, db.lastCreate.UserId)
	}
	if db.lastLimit != 5 {
		t.Fatalf("expected limit 5, got %d", db.lastLimit)
	}
	if ts.responder.status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, ts.responder.status)
	}
}

func TestCreateNewProject_Unauthorized(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)
	ts.authErr = errors.New("not authorized")

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()

	ts.CreateNewProject(w, req)

	if db.createCalled {
		t.Fatalf("database should not be called on auth failure")
	}
	if ts.responder.status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", ts.responder.status)
	}
}

func TestCreateNewProject_BadJSON(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("{bad")))
	w := httptest.NewRecorder()

	ts.CreateNewProject(w, req)

	if ts.responder.status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", ts.responder.status)
	}
	if db.createCalled {
		t.Fatalf("database should not be called on bad json")
	}
}

func TestCreateNewProject_DBError(t *testing.T) {
	db := &fakeDB{createStatus: http.StatusTooManyRequests, createErr: errors.New("limit reached")}
	ts := newTestServer(db)

	body := model.ProjectModel{Name: "p", Desc: "d"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(b))
	w := httptest.NewRecorder()

	ts.CreateNewProject(w, req)

	if ts.responder.status != http.StatusTooManyRequests {
		t.Fatalf("expected passthrough status, got %d", ts.responder.status)
	}
}

func TestEditProject_Success(t *testing.T) {
	db := &fakeDB{editStatus: http.StatusOK}
	ts := newTestServer(db)

	payload := model.ProjectModel{Id: "p1", Name: "n", Desc: "d"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/projects", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer ok")
	w := httptest.NewRecorder()

	ts.EditProject(w, req)

	if !db.editCalled {
		t.Fatalf("expected EditProject to be called")
	}
	if db.lastEdit.UserId != ts.authUser {
		t.Fatalf("expected UserId to be set on edit payload")
	}
	if ts.responder.status != http.StatusOK {
		t.Fatalf("expected 200, got %d", ts.responder.status)
	}
}

func TestEditProject_MissingFields(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)

	// Missing Desc
	payload := model.ProjectModel{Id: "p1", Name: "n"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/projects", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer ok")
	w := httptest.NewRecorder()

	ts.EditProject(w, req)

	if db.editCalled {
		t.Fatalf("db should not be called when request invalid")
	}
	if ts.responder.status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", ts.responder.status)
	}
}

func TestEditProject_BadBodyRead(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)

	// Simulate body read error via a reader that errors
	errReader := io.NopCloser(&failingReader{})
	req := httptest.NewRequest(http.MethodPut, "/projects", errReader)
	req.Header.Set("Authorization", "Bearer ok")
	w := httptest.NewRecorder()

	ts.EditProject(w, req)

	if ts.responder.status != http.StatusBadRequest {
		t.Fatalf("expected 400 for body read error, got %d", ts.responder.status)
	}
}

type failingReader struct{}
func (f *failingReader) Read(p []byte) (int, error) { return 0, errors.New("boom") }

func TestEditProject_DBError(t *testing.T) {
	db := &fakeDB{editStatus: http.StatusConflict, editErr: errors.New("conflict")}
	ts := newTestServer(db)

	payload := model.ProjectModel{Id: "p1", Name: "n", Desc: "d"}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/projects", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer ok")
	w := httptest.NewRecorder()

	ts.EditProject(w, req)

	if ts.responder.status != http.StatusConflict {
		t.Fatalf("expected 409 passthrough, got %d", ts.responder.status)
	}
}

func TestCheckProjectPermission_Success(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)
	ts.permOK = true

	req := httptest.NewRequest(http.MethodGet, "/perm?projectId=pid1", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.CheckProjectPermission(w, req)

	if ts.responder.status != http.StatusOK {
		t.Fatalf("expected 200, got %d", ts.responder.status)
	}
}

func TestCheckProjectPermission_AuthFail(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)
	ts.authErr = errors.New("bad token")

	req := httptest.NewRequest(http.MethodGet, "/perm?projectId=pid1", nil)
	w := httptest.NewRecorder()

	ts.CheckProjectPermission(w, req)

	if ts.responder.status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", ts.responder.status)
	}
}

func TestCheckProjectPermission_PermErr(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)
	ts.permErr = errors.New("nope")

	req := httptest.NewRequest(http.MethodGet, "/perm?projectId=pid1", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.CheckProjectPermission(w, req)

	if ts.responder.status != http.StatusUnauthorized {
		t.Fatalf("expected 401 on perm error, got %d", ts.responder.status)
	}
}

func TestProjectAccessRequest_Success(t *testing.T) {
	db := &fakeDB{}
	ts := newTestServer(db)

	req := httptest.NewRequest(http.MethodPost, "/access?projectId=pid1", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.ProjectAccessRequest(w, req)

	if !db.accessCalled {
		t.Fatalf("expected ProjectAccessRequest to be called")
	}
	if db.lastProjID != "pid1" || db.lastUser != ts.authUser {
		t.Fatalf("expected correct args to DB, got user=%s project=%s", db.lastUser, db.lastProjID)
	}
	if ts.responder.status != http.StatusOK {
		t.Fatalf("expected 200, got %d", ts.responder.status)
	}
}

func TestProjectAccessRequest_DBErr(t *testing.T) {
	db := &fakeDB{accessErr: errors.New("db down")}
	ts := newTestServer(db)

	req := httptest.NewRequest(http.MethodPost, "/access?projectId=pid1", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.ProjectAccessRequest(w, req)

	if ts.responder.status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", ts.responder.status)
	}
}

func TestRetrieveProjectList_Success(t *testing.T) {
	db := &fakeDB{
		listProjects: []model.ProjectModel{{Id: "1", Name: "A"}, {Id: "2", Name: "B"}},
		listStatus:   http.StatusOK,
	}
	ts := newTestServer(db)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.RetrieveProjectList(w, req)

	if !db.listCalled {
		t.Fatalf("expected RetrieveProjectList to be called")
	}
	if db.lastUser != ts.authUser {
		t.Fatalf("expected user id to be passed to DB")
	}
	if ts.responder.status != http.StatusOK {
		t.Fatalf("expected 200, got %d", ts.responder.status)
	}
}

func TestRetrieveProjectList_DBErr(t *testing.T) {
	db := &fakeDB{
		listStatus: http.StatusServiceUnavailable,
		listErr:    errors.New("svc down"),
	}
	ts := newTestServer(db)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "ok")
	w := httptest.NewRecorder()

	ts.RetrieveProjectList(w, req)

	if ts.responder.status != http.StatusServiceUnavailable {
		t.Fatalf("expected status passthrough, got %d", ts.responder.status)
	}
}

// Minimal config structs to satisfy APIServerModel usage in handlers.
// These mirror only the fields referenced by the handlers under test.
type ServerConfig struct {
	ProjectConfig ProjectConfig
}
type ProjectConfig struct {
	Limit int
}

// Note: We assume APIServerModel has fields: ctx, database, config, logger.
// If actual names differ, adjust test helpers accordingly.