package database

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlLib "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"ncap-api-server/common"
	"ncap-api-server/model"
)

func newMock(t *testing.T) (*MySql, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	mysql := &MySql{db: db}
	cleanup := func() {
		_ = db.Close()
	}
	return mysql, mock, cleanup
}

// --- registerTLSConfig tests ---

func writeTempCert(t *testing.T, dir string) string {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	path := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatalf("write pem: %v", err)
	}
	return path
}

func Test_registerTLSConfig_success(t *testing.T) {
	dir := t.TempDir()
	caPath := writeTempCert(t, dir)
	if err := registerTLSConfig(caPath); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func Test_registerTLSConfig_missingFile(t *testing.T) {
	err := registerTLSConfig("database/testdata/missing-ca.pem")
	if err == nil || !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), "failed to read CA cert") {
		t.Fatalf("expected file read error, got: %v", err)
	}
}

func Test_registerTLSConfig_invalidPEM(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(p, []byte("not a cert"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	err := registerTLSConfig(p)
	if err == nil {
		t.Fatalf("expected append certs error, got nil")
	}
}

// --- CreateNewProject tests ---

func Test_CreateNewProject_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	pm := &model.ProjectModel{UserId: 42, Name: "proj", Desc: "desc"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(id\) FROM project_details where user_id = \?`).
		WithArgs(pm.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectExec(`INSERT INTO project_details`).
		WithArgs(pm.UserId, pm.Name, pm.Desc).
		WillReturnResult(sqlmock.NewResult(101, 1))

	mock.ExpectExec(`INSERT INTO users_projects`).
		WithArgs(pm.UserId, int(common.OWNER), int(common.APPROVED)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO users_projects`).
		WithArgs(pm.UserId, int(common.ADMIN), int(common.APPROVED)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	status, err := mysql.CreateNewProject(ctx, pm, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func Test_CreateNewProject_limitExceeded(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	pm := &model.ProjectModel{UserId: 7, Name: "p", Desc: "d"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(id\) FROM project_details where user_id = \?`).
		WithArgs(pm.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectRollback()

	status, err := mysql.CreateNewProject(ctx, pm, 3)
	if status != http.StatusConflict {
		t.Fatalf("expected conflict, got %d err=%v", status, err)
	}
	if err == nil || !strings.Contains(err.Error(), "project limit exceeded") {
		t.Fatalf("expected project limit exceeded error, got %v", err)
	}
	_ = mock.ExpectationsWereMet()
}

func Test_CreateNewProject_beginTxError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	pm := &model.ProjectModel{UserId: 1}

	mock.ExpectBegin().WillReturnError(errors.New("boom"))
	status, err := mysql.CreateNewProject(ctx, pm, 10)
	if status != http.StatusInternalServerError || err == nil {
		t.Fatalf("expected 500 with error, got %d %v", status, err)
	}
}

func Test_CreateNewProject_insertUserProjectError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	pm := &model.ProjectModel{UserId: 9, Name: "p", Desc: "d"}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COUNT\(id\) FROM project_details where user_id = \?`).
		WithArgs(pm.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO project_details`).
		WithArgs(pm.UserId, pm.Name, pm.Desc).
		WillReturnResult(sqlmock.NewResult(55, 1))
	mock.ExpectExec(`INSERT INTO users_projects`).
		WithArgs(pm.UserId, int(common.OWNER), int(common.APPROVED)).
		WillReturnError(errors.New("insert owner err"))
	mock.ExpectRollback()

	status, err := mysql.CreateNewProject(ctx, pm, 10)
	if status != http.StatusInternalServerError || err == nil {
		t.Fatalf("expected 500 with error, got %d %v", status, err)
	}
}

// --- CreateTable tests ---

func Test_CreateTable_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	ct := &model.CreateTableModel{
		DisplayName: "Items",
		TableName:   "items",
		Columns: map[string]string{
			"name": "STRING",
			"age":  "NUMBER",
		},
		PrimaryKeys: []string{},
	}

	mock.ExpectExec(`CREATE TABLE items`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	status, err := mysql.CreateTable(ctx, ct)
	if err != nil || status != http.StatusCreated {
		t.Fatalf("expected created, got %d err=%v", status, err)
	}
}

// --- InsertIntoTable tests ---

func Test_InsertIntoTable_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	im := &model.InsertIntoTableModel{
		TableName: "items",
		Columns: map[string]string{
			"name": "Alice",
			"age":  "30",
		},
	}

	// Columns order is map-based (non-deterministic). Match loosely.
	mock.ExpectExec(`INSERT INTO items`).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := mysql.InsertIntoTable(ctx, im); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_InsertIntoTable_execError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	im := &model.InsertIntoTableModel{
		TableName: "items",
		Columns:   map[string]string{"x": "y"},
	}
	mock.ExpectExec(`INSERT INTO items`).WillReturnError(errors.New("exec fail"))

	err := mysql.InsertIntoTable(ctx, im)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- DropTable tests ---

func Test_DropTable_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	dm := &model.DropTableModel{TableName: "items"}

	mock.ExpectBegin()
	mock.ExpectExec(`DROP TABLE IF EXISTS items`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM project_metadata WHERE table_name = \?`).WithArgs(dm.TableName).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := mysql.DropTable(ctx, dm); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_DropTable_deleteMetaError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	dm := &model.DropTableModel{TableName: "items"}

	mock.ExpectBegin()
	mock.ExpectExec(`DROP TABLE IF EXISTS items`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM project_metadata WHERE table_name = \?`).WithArgs(dm.TableName).
		WillReturnError(errors.New("meta delete err"))
	mock.ExpectRollback()

	if err := mysql.DropTable(ctx, dm); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- InsertIntoMetaData tests ---

func Test_InsertIntoMetaData_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()

	ctx := context.Background()
	ct := &model.CreateTableModel{
		DisplayName: "Items",
		TableName:   "items",
		Description: "desc",
	}
	mock.ExpectExec(`INSERT INTO project_metadata`).
		WithArgs(10, ct.DisplayName, ct.TableName, sqlmock.AnyArg(), ct.Description).
		WillReturnResult(sqlmock.NewResult(1, 1))

	status, err := mysql.InsertIntoMetaData(ctx, ct, 10)
	if err != nil || status != http.StatusOK {
		t.Fatalf("expected 200, got %d err=%v", status, err)
	}
}

func Test_InsertIntoMetaData_execError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	ct := &model.CreateTableModel{DisplayName: "x", TableName: "y"}
	mock.ExpectExec(`INSERT INTO project_metadata`).WillReturnError(errors.New("boom"))

	_, err := mysql.InsertIntoMetaData(ctx, ct, 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- RetrieveTableListFromMetaData tests ---

func Test_RetrieveTableListFromMetaData_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	meta1 := `{"display_name":"A","table_name":"a"}`
	meta2 := `{"display_name":"B","table_name":"b"}`

	rows := sqlmock.NewRows([]string{"METADATA"}).AddRow(meta1).AddRow(meta2)
	mock.ExpectQuery(`SELECT METADATA FROM project_metadata WHERE project_id = \?`).
		WithArgs("12").WillReturnRows(rows)

	list, status, err := mysql.RetrieveTableListFromMetaData(ctx, "12")
	if err != nil || status != http.StatusOK {
		t.Fatalf("expected ok, got %d err=%v", status, err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list))
	}
}

func Test_RetrieveTableListFromMetaData_unmarshalError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"METADATA"}).AddRow("not-json")
	mock.ExpectQuery(`SELECT METADATA FROM project_metadata WHERE project_id = \?`).
		WithArgs("12").WillReturnRows(rows)

	_, _, err := mysql.RetrieveTableListFromMetaData(ctx, "12")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- RetrieveDataFromTable tests ---

func Test_RetrieveDataFromTable_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	// Lookup real table name
	mock.ExpectQuery(`SELECT TABLE_NAME FROM project_metadata where project_id = \? and display_name = \?`).
		WithArgs("12", "Items").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("items"))

	// Data query
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, []byte("alice")). // []byte should be converted to string
		AddRow(2, "bob")
	mock.ExpectQuery(`SELECT \* FROM items`).WillReturnRows(rows)

	cols, data, err := mysql.RetrieveDataFromTable(ctx, &model.RetrieveDataFromTableModel{
		ProjectId:   "12",
		DisplayName: "Items",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 2 || cols[0] != "id" || cols[1] != "name" {
		t.Fatalf("unexpected columns: %v", cols)
	}
	if len(data) != 2 || data[0]["name"] != "alice" || data[1]["name"] != "bob" {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func Test_RetrieveDataFromTable_tableLookupNoRows(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT TABLE_NAME FROM project_metadata where project_id = \? and display_name = \?`).
		WithArgs("12", "Items").
		WillReturnError(sql.ErrNoRows)

	_, _, err := mysql.RetrieveDataFromTable(ctx, &model.RetrieveDataFromTableModel{
		ProjectId:   "12",
		DisplayName: "Items",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- DeleteEntryFromTable tests ---

func Test_DeleteEntryFromTable_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT TABLE_NAME FROM project_metadata where project_id = \? and display_name = \?`).
		WithArgs("12", "Items").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("items"))
	mock.ExpectExec(`DELETE FROM items WHERE id = \?`).WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := mysql.DeleteEntryFromTable(ctx, &model.DeleteEntryFromTableModel{
		ProjectId:   "12",
		DisplayName: "Items",
		TableRowId:  99,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_DeleteEntryFromTable_deleteError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT TABLE_NAME FROM project_metadata where project_id = \? and display_name = \?`).
		WithArgs("12", "Items").
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}).AddRow("items"))
	mock.ExpectExec(`DELETE FROM items WHERE id = \?`).WithArgs(99).
		WillReturnError(errors.New("delete err"))

	err := mysql.DeleteEntryFromTable(ctx, &model.DeleteEntryFromTableModel{
		ProjectId:   "12",
		DisplayName: "Items",
		TableRowId:  99,
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// --- RegisterUser tests ---

func Test_RegisterUser_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	um := &model.UserModel{Name: "n", Email: "e@x.y", Password: "p"}
	mock.ExpectExec(`INSERT INTO users \(name, email, password\) value \(?, ?, \?\)`).
		WithArgs(um.Name, um.Email, um.Password).
		WillReturnResult(sqlmock.NewResult(77, 1))

	id, status, err := mysql.RegisterUser(ctx, um)
	if err != nil || status != http.StatusCreated || id != 77 {
		t.Fatalf("expected created id=77, got id=%d status=%d err=%v", id, status, err)
	}
}

func Test_RegisterUser_duplicate(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	um := &model.UserModel{Name: "n", Email: "e@x.y", Password: "p"}
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(um.Name, um.Email, um.Password).
		WillReturnError(&mysqlLib.MySQLError{Number: 1062, Message: "dup"})

	_, status, err := mysql.RegisterUser(ctx, um)
	if status != http.StatusConflict || err == nil {
		t.Fatalf("expected conflict, got %d err=%v", status, err)
	}
}

// --- LoginUser tests ---

func Test_LoginUser_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	pw := "secret"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	u := &model.UserModel{Email: "e@x.y", Password: pw}

	mock.ExpectQuery(`SELECT id, password, name FROM users where email = \?`).
		WithArgs(u.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "name"}).AddRow(11, string(hash), "Alice"))

	user, status, err := mysql.LoginUser(ctx, u)
	if err != nil || status != http.StatusAccepted {
		t.Fatalf("expected accepted, got %d err=%v", status, err)
	}
	if user.Id != 11 || user.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func Test_LoginUser_notFound(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	u := &model.UserModel{Email: "x@y.z"}

	mock.ExpectQuery(`SELECT id, password, name FROM users where email = \?`).
		WithArgs(u.Email).WillReturnError(sql.ErrNoRows)

	_, status, err := mysql.LoginUser(ctx, u)
	if status != http.StatusNotFound || err == nil {
		t.Fatalf("expected 404, got %d err=%v", status, err)
	}
}

func Test_LoginUser_wrongPassword(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	rightHash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	u := &model.UserModel{Email: "e@x.y", Password: "wrong"}

	mock.ExpectQuery(`SELECT id, password, name FROM users where email = \?`).
		WithArgs(u.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "password", "name"}).AddRow(1, string(rightHash), "Name"))

	_, status, err := mysql.LoginUser(ctx, u)
	if status != http.StatusUnauthorized || err == nil {
		t.Fatalf("expected 401, got %d err=%v", status, err)
	}
}

// --- RetrieveProjectList tests ---

func Test_RetrieveProjectList_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "name", "description"}).
		AddRow(1, "A", "a").AddRow(2, "B", "b")
	mock.ExpectQuery(`SELECT id, name, description FROM project_details where user_id = \?`).
		WithArgs(9).WillReturnRows(rows)

	list, status, err := mysql.RetrieveProjectList(ctx, 9)
	if err != nil || status != http.StatusOK || len(list) != 2 {
		t.Fatalf("unexpected: status=%d len=%d err=%v", status, len(list), err)
	}
}

func Test_RetrieveProjectList_queryError(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	mock.ExpectQuery(`SELECT id, name, description FROM project_details where user_id = \?`).
		WithArgs(9).WillReturnError(errors.New("q err"))

	_, status, err := mysql.RetrieveProjectList(ctx, 9)
	if status != http.StatusInternalServerError || err == nil {
		t.Fatalf("expected 500, got %d err=%v", status, err)
	}
}

// --- Close tests ---

func Test_Close_ok(t *testing.T) {
	mysql, _, cleanup := newMock(t)
	defer cleanup()
	if err := mysql.Close(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// --- EditProject tests ---

func Test_EditProject_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	pm := &model.ProjectModel{Id: 3, UserId: 9, Name: "n", Desc: "d"}

	mock.ExpectExec(`UPDATE project_details SET name = \?, description = \? WHERE id = \? and user_id = \?`).
		WithArgs(pm.Name, pm.Desc, pm.Id, pm.UserId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	status, err := mysql.EditProject(ctx, pm)
	if err != nil || status != http.StatusOK {
		t.Fatalf("expected ok, got %d err=%v", status, err)
	}
}

func Test_EditProject_error(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()
	pm := &model.ProjectModel{Id: 3, UserId: 9, Name: "n", Desc: "d"}

	mock.ExpectExec(`UPDATE project_details`).WillReturnError(&mysqlLib.MySQLError{Number: 1064, Message: "syntax"})

	status, err := mysql.EditProject(ctx, pm)
	if status != http.StatusInternalServerError || err == nil {
		t.Fatalf("expected 500, got %d err=%v", status, err)
	}
}

// --- CheckPermission tests ---

func Test_CheckPermission_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT user_id FROM users_projects WHERE user_id = \? and project_id = \? and status = \? and role = \?`).
		WithArgs(9, "12", int(common.APPROVED), int(common.OWNER)).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(9))

	status, err := mysql.CheckPermission(ctx, 9, "12", common.OWNER)
	if err != nil || status != http.StatusOK {
		t.Fatalf("expected ok, got %d err=%v", status, err)
	}
}

func Test_CheckPermission_notFound(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectQuery(`SELECT user_id FROM users_projects WHERE user_id = \? and project_id = \? and status = \? and role = \?`).
		WithArgs(9, "12", int(common.APPROVED), int(common.OWNER)).
		WillReturnError(sql.ErrNoRows)

	status, err := mysql.CheckPermission(ctx, 9, "12", common.OWNER)
	if status != http.StatusNotFound || err == nil {
		t.Fatalf("expected 404, got %d err=%v", status, err)
	}
}

// --- ProjectAccessRequest tests ---

func Test_ProjectAccessRequest_success(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO users_projects \(user_id, project_id\) VALUES \(\?, \?\)`).
		WithArgs(9, "12").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := mysql.ProjectAccessRequest(ctx, 9, "12"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ProjectAccessRequest_error(t *testing.T) {
	mysql, mock, cleanup := newMock(t)
	defer cleanup()
	ctx := context.Background()

	mock.ExpectExec(`INSERT INTO users_projects`).WithArgs(9, "12").WillReturnError(errors.New("boom"))

	if err := mysql.ProjectAccessRequest(ctx, 9, "12"); err == nil {
		t.Fatalf("expected error, got nil")
	}
}