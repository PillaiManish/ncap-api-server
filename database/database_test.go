package database

import (
	"reflect"
	"errors"
	"strings"
	"testing"

	"ncap-api-server/common"
	"ncap-api-server/model"
)

func TestMapDataType_MySQLMappings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		inputType    string
		databaseType string
		want         string
	}{
		{"text maps to text", "text", common.MYSQL, "text"},
		{"boolean maps to tinyint(1)", "boolean", common.MYSQL, "tinyint(1)"},
		{"int maps to int", "int", common.MYSQL, "int"},
		{"unknown type passes through", "varchar(255)", common.MYSQL, "varchar(255)"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapDataType(tt.inputType, tt.databaseType)
			if got != tt.want {
				t.Fatalf("MapDataType(%q, %q) = %q; want %q", tt.inputType, tt.databaseType, got, tt.want)
			}
		})
	}
}

func TestMapDataType_DefaultPassthroughForUnknownDB(t *testing.T) {
	t.Parallel()
	// Using a database type that is not explicitly handled should return the original data type.
	inputs := []string{"json", "uuid", "customType"}
	for _, in := range inputs {
		in := in
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			got := MapDataType(in, "unknown-db")
			if got != in {
				t.Fatalf("MapDataType(%q, %q) = %q; want passthrough %q", in, "unknown-db", got, in)
			}
		})
	}
}

func TestNewDatabase_UnsupportedTypeReturnsError(t *testing.T) {
	t.Parallel()
	cfg := &model.DatabaseConfig{
		Type: "sqlite", // assumed unsupported per switch
	}
	db, err := NewDatabase(cfg)
	if err == nil {
		if db != nil {
			t.Fatalf("expected error for unsupported type; got non-nil db and nil error")
		}
		t.Fatalf("expected non-nil error for unsupported type")
	}
	// Verify error message contains the type and 'not supported' phrase to avoid brittle exact-match
	if !strings.Contains(err.Error(), cfg.Type) || !strings.Contains(strings.ToLower(err.Error()), "not supported") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewDatabase_MySQLPathDoesNotPanic(t *testing.T) {
	t.Parallel()
	// We do not assert success because NewMySQLConnection may require a live DB.
	// Instead, we ensure the factory does not panic and returns either a db or a well-formed error.
	cfg := &model.DatabaseConfig{
		Type: common.MYSQL,
		// Leave the remainder of config fields zero-valued to avoid external dependency.
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewDatabase panicked for MYSQL config: %v", r)
		}
	}()
	_, err := NewDatabase(cfg)
	// Accept either nil or non-nil error; just ensure the function executes safely.
	var _ error = err
}

func TestDatabaseInterface_SignatureExpectations(t *testing.T) {
	t.Parallel()
	// This test provides a compile-time check: any mismatch in method names/signatures will fail to compile
	// when we attempt to accept a type that claims to implement 'Database'.
	// We can't instantiate a concrete implementation here, so we only assert interface method count via reflection as a sanity check.
	const expectedMethodCount = 13 // Update this if the interface changes
	typ := (*Database)(nil)
	// Use reflection without importing reflect directly by relying on the method set of a nil interface.
	// If expectedMethodCount goes out of sync, this test reminds maintainers to update tests.
	_ = typ
	// We still import reflect to count the methods.
	type methodCounter interface{}
	var mc methodCounter = typ
	if mc == nil {
		// Just avoid lint warnings; the real count below:
	}
	// Now real count:
	// Import reflect here to avoid scope issues.
	// We avoid t.Helper() to keep standard libs minimal.
	if got := interfaceMethodCount((*Database)(nil)); got != expectedMethodCount {
		t.Fatalf("Database interface method count = %d; want %d", got, expectedMethodCount)
	}
}

// interfaceMethodCount returns the number of methods declared by the Database interface.
// Isolated as a helper so it is easy to test/update without exposing extra APIs.

// _databaseInterfaceMethodCount is separated so we can safely import reflect in one place without unused import issues.

// interfaceMethodCount returns the number of methods on the Database interface using reflect.
func interfaceMethodCount(_ Database) int {
	tp := reflect.TypeOf((*Database)(nil)).Elem()
	return tp.NumMethod()
}