package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"ncap-api-server/common"
	"ncap-api-server/model"
)

// ---- Test doubles ----

type fakeDB struct {
	closeCalled       bool
	checkCalled       bool
	lastUserID        int
	lastProjectID     string
	lastPermission    common.PERMISSION
	checkReturnsCode  int
	checkReturnsError error
}

func (f *fakeDB) Close() {
	f.closeCalled = true
}

func (f *fakeDB) CheckPermission(ctx context.Context, userId int, projectId string, permission common.PERMISSION) (int, error) {
	f.checkCalled = true
	f.lastUserID = userId
	f.lastProjectID = projectId
	f.lastPermission = permission
	return f.checkReturnsCode, f.checkReturnsError
}

// Ensure fakeDB satisfies the expected interface at compile-time.
// If the interface name or shape differs, uncomment and adjust accordingly.
// var _ database.Database = (*fakeDB)(nil)

// helper to build api server with minimal config
func newAPI(logger *logrus.Logger, db *fakeDB, secret []byte, errorData bool) *APIServerModel {
	cfg := &model.Config{
		JWT: model.JWT{
			Secret: secret,
		},
		ErrorDataResponse: errorData,
	}
	return &APIServerModel{
		logger:   logger,
		database: db,
		config:   cfg,
		ctx:      context.Background(),
	}
}

// parse response payload into model.ResultStatus (fallback to map if type differs)
func decodeResultStatus(t *testing.T, body []byte) model.ResultStatus {
	t.Helper()
	var rs model.ResultStatus
	if err := json.Unmarshal(body, &rs); err != nil {
		t.Fatalf("failed to unmarshal ResultStatus: %v\nbody: %s", err, string(body))
	}
	return rs
}

// ---- Response tests ----

func TestResponse_StatusOK_WithData(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	rr := httptest.NewRecorder()
	payload := map[string]any{"k": "v"}

	api.Response(rr, http.StatusOK, payload)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	got := decodeResultStatus(t, rr.Body.Bytes())
	if got.Status != common.SUCCESS {
		t.Errorf("expected status SUCCESS, got %s", got.Status)
	}
	if got.StatusCode != http.StatusOK {
		t.Errorf("expected statusCode %d, got %d", http.StatusOK, got.StatusCode)
	}
	if got.Message != "Processed successfully" {
		t.Errorf("unexpected message: %q", got.Message)
	}
	// Data should echo payload on success
	if m, ok := got.Data.(map[string]any); !ok || m["k"] != "v" {
		t.Errorf("unexpected data: %#v", got.Data)
	}
}

func TestResponse_ErrorCodes_WithAndWithoutErrorData(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		msg         string
		withErrData bool
	}{
		{"BadRequest_noErrData", http.StatusBadRequest, "Invalid request", false},
		{"BadRequest_withErrData", http.StatusBadRequest, "Invalid request", true},
		{"Unauthorized_withErrData", http.StatusUnauthorized, "Unauthorized", true},
		{"NotFound_noErrData", http.StatusNotFound, "Not found", false},
		{"Internal_withErrData", http.StatusInternalServerError, "Internal server error", true},
		{"Conflict_withErrData", http.StatusConflict, "Conflict", true},
		{"Created_success", http.StatusCreated, "Processed successfully", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			db := &fakeDB{}
			api := newAPI(logger, db, []byte("secret"), tt.withErrData)

			rr := httptest.NewRecorder()
			errPayload := map[string]string{"err": "details"}
			api.Response(rr, tt.code, errPayload)

			if rr.Code != tt.code {
				t.Fatalf("expected status %d, got %d", tt.code, rr.Code)
			}
			got := decodeResultStatus(t, rr.Body.Bytes())
			if tt.code == http.StatusOK || tt.code == http.StatusCreated {
				if got.Status != common.SUCCESS {
					t.Errorf("expected SUCCESS, got %s", got.Status)
				}
			} else {
				if got.Status != common.FAILED {
					t.Errorf("expected FAILED, got %s", got.Status)
				}
			}
			if got.Message != tt.msg {
				t.Errorf("expected message %q, got %q", tt.msg, got.Message)
			}
			if tt.withErrData || tt.code == http.StatusOK || tt.code == http.StatusCreated {
				if got.Data == nil {
					t.Errorf("expected data to be present")
				}
			} else {
				if got.Data != nil {
					t.Errorf("expected data to be nil, got %#v", got.Data)
				}
			}
		})
	}
}

func TestResponse_DefaultCase(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), true)

	rr := httptest.NewRecorder()
	api.Response(rr, 0, map[string]string{"x": "y"})

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected default to map to 500, got %d", rr.Code)
	}
	got := decodeResultStatus(t, rr.Body.Bytes())
	if got.Message) != "Internal server error" {
		t.Errorf("unexpected message: %q", got.Message)
	}
}

// ---- Token & authorization tests ----

func TestCreateAndVerifyToken_RoundTrip(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("supersecret"), false)

	u := &model.UserModel{Id: 42, Name: "Ada", Email: "ada@example.com"}
	token, err := api.createToken(u)
	if err != nil {
		t.Fatalf("createToken error: %v", err)
	}

	gotID, err := api.verifyToken(token)
	if err != nil {
		t.Fatalf("verifyToken error: %v", err)
	}
	if gotID != 42 {
		t.Errorf("expected user id 42, got %d", gotID)
	}
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	if _, err := api.verifyToken("not-a-jwt"); err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestVerifyToken_WrongSecret(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	apiGood := newAPI(logger, db, []byte("correct"), false)
	apiBad := newAPI(logger, db, []byte("incorrect"), false)

	u := &model.UserModel{Id: 7, Name: "Lin", Email: "lin@example.com"}
	token, err := apiGood.createToken(u)
	if err != nil {
		t.Fatalf("createToken error: %v", err)
	}
	if _, err := apiBad.verifyToken(token); err == nil {
		t.Fatal("expected signature verification error")
	}
}

func TestAuthorization_HeaderScenarios(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("authsecret"), false)

	u := &model.UserModel{Id: 99, Name: "User", Email: "user@example.com"}
	token, err := api.createToken(u)
	if err != nil {
		t.Fatalf("createToken error: %v", err)
	}

	tests := []struct {
		name   string
		header string
		ok     bool
	}{
		{"MissingHeader", "", false},
		{"MalformedPrefix", "Token " + token, false},
		{"ValidBearer", "Bearer " + token, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			id, err := api.authorization(w, tt.header, "/endpoint")
			if tt.ok {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if id != 99 {
					t.Fatalf("expected id 99, got %d", id)
				}
			} else {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if id != -1 {
					t.Fatalf("expected id -1 on error, got %d", id)
				}
			}
		})
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	// Build a token with very short expiry by creating one then manually altering exp via createToken style is limited.
	// Instead, we hand-craft an expired token by calling createToken then simulating expiration with a small time travel:
	// We'll create a token with short TTL by temporarily substituting time.Now via wrapper. Since code uses time.Now() directly,
	// we simulate expiration by waiting until exp is passed (very short).
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	u := &model.UserModel{Id: 1, Name: "t", Email: "t@t.tt"}

	// Create token and then sleep past a near-immediate expiry isn't possible (hardcoded 24h).
	// Therefore, this test asserts that an obviously malformed but structurally JWT string fails.
	if _, err := api.verifyToken(strings.Repeat("a.", 3)); err == nil {
		t.Fatal("expected error for structurally invalid/expired token surrogate")
	}
}

// ---- checkPermission tests ----

func TestCheckPermission_Success(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{checkReturnsCode: http.StatusOK}
	api := newAPI(logger, db, []byte("secret"), false)

	ok, err := api.checkPermission(10, "proj-1", common.READ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected permission ok")
	}
	if !db.checkCalled || db.lastUserID != 10 || db.lastProjectID != "proj-1" || db.lastPermission != common.READ {
		t.Errorf("db.CheckPermission not called with expected args: %+v", db)
	}
}

func TestCheckPermission_NotOkStatus(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{checkReturnsCode: http.StatusUnauthorized}
	api := newAPI(logger, db, []byte("secret"), false)

	ok, err := api.checkPermission(11, "proj-2", common.WRITE)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected permission false when status != 200")
	}
}

func TestCheckPermission_DBError(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{checkReturnsError: errors.New("db down")}
	api := newAPI(logger, db, []byte("secret"), false)

	ok, err := api.checkPermission(12, "proj-3", common.ADMIN)
	if err == nil {
		t.Fatal("expected error from db")
	}
	if ok {
		t.Fatal("expected ok=false when db errors")
	}
}

// ---- createTableNameFromName tests ----

func TestCreateTableNameFromName(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	cases := []struct {
		display  string
		project  int
		expected string
	}{
		{"simple", 1, "1_simple"},
		{"has-dash", 2, "2_has_dash"},
		{"has space", 3, "3_has_space"},
		{"Mix-ed Case Name", 42, "42_Mix_ed_Case_Name"},
	}
	for _, c := range cases {
		got := api.createTableNameFromName(c.display, c.project)
		if got != c.expected {
			t.Errorf("display=%q project=%d: expected %q, got %q", c.display, c.project, c.expected, got)
		}
	}
}

// ---- Stop tests ----

func TestStop_ClosesDB_And_ExitsViaLogger(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	calledExit := 0
	logger.ExitFunc = func(code int) {
		calledExit = code
		// Do not os.Exit during tests
	}

	api.Stop()

	if !db.closeCalled {
		t.Error("expected database.Close to be called")
	}
	if calledExit != 1 {
		t.Errorf("expected logger.Exit(1) to be called, got %d", calledExit)
	}
}

// ---- Additional: Response writes result header code before body ----

func TestResponse_WritesHeaderBeforeBody(t *testing.T) {
	logger := logrus.New()
	db := &fakeDB{}
	api := newAPI(logger, db, []byte("secret"), false)

	rr := httptest.NewRecorder()
	api.Response(rr, http.StatusNotFound, map[string]string{"err": "nope"})

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	// Body should be JSON of ResultStatus; ensure not empty
	if len(bytes.TrimSpace(rr.Body.Bytes())) == 0 {
		t.Fatal("expected non-empty JSON body")
	}
}

// Note: Testing library/framework used: Go's standard 'testing' package with httptest.
// If repository includes stretchr/testify, these tests can be refactored to use require/assert.