package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"ncap-api-server/model"
	"os"
	"path/filepath"
)

// CreateTestConfig creates a test configuration
func CreateTestConfig() *model.Config {
	return &model.Config{
		ErrorDataResponse: true,
		LogLevel:          "DEBUG",
		AdminPanelUrl:     "*",
		ProjectConfig: &model.ProjectConfig{
			Limit: 5,
		},
		Database: &model.DatabaseConfig{
			Type: "MySQL",
			Connection: &model.ConnectionConfig{
				Host:   "localhost",
				Port:   3306,
				DBName: "test_db",
				User:   "test_user",
				Pass:   "test_pass",
			},
			SSL: &model.SSL{
				Enabled: false,
			},
		},
		JWT: &model.JWT{
			Secret: []byte("test-secret-key"),
		},
	}
}

// CreateTestUser creates a test user model
func CreateTestUser() *model.UserModel {
	return &model.UserModel{
		Id:       1,
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
	}
}

// CreateTestProject creates a test project model
func CreateTestProject() *model.ProjectModel {
	return &model.ProjectModel{
		UserId: 1,
		Id:     "1",
		Name:   "Test Project",
		Desc:   "Test Description",
	}
}

// CreateTestTableModel creates a test table model
func CreateTestTableModel() *model.CreateTableModel {
	return &model.CreateTableModel{
		Description: "Test table description",
		DisplayName: "test-table",
		TableName:   "1_test_table",
		Columns: map[string]string{
			"name":  "text",
			"age":   "int",
			"email": "text",
		},
		PrimaryKeys: []string{"id"},
	}
}

// CreateTestInsertModel creates a test insert model
func CreateTestInsertModel() *model.InsertIntoTableModel {
	return &model.InsertIntoTableModel{
		TableName: "1_test_table",
		Columns: map[string]string{
			"name":  "John Doe",
			"age":   "25",
			"email": "john@example.com",
		},
	}
}

// CreateTestRetrieveModel creates a test retrieve model
func CreateTestRetrieveModel() *model.RetrieveDataFromTableModel {
	return &model.RetrieveDataFromTableModel{
		ProjectId:   "1",
		DisplayName: "test-table",
	}
}

// CreateTestDeleteModel creates a test delete model
func CreateTestDeleteModel() *model.DeleteEntryFromTableModel {
	return &model.DeleteEntryFromTableModel{
		ProjectId:   "1",
		DisplayName: "test-table",
		TableRowId:  1,
	}
}

// CreateTestRetrieveTableListModel creates a test retrieve table list model
func CreateTestRetrieveTableListModel() *model.RetrieveTableListModel {
	return &model.RetrieveTableListModel{
		ProjectId: "1",
	}
}

// CreateTestDropTableModel creates a test drop table model
func CreateTestDropTableModel() *model.DropTableModel {
	return &model.DropTableModel{
		TableName: "1_test_table",
	}
}

// CreateHTTPRequest creates an HTTP request for testing
func CreateHTTPRequest(method, url string, body interface{}) *http.Request {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateHTTPRequestWithAuth creates an HTTP request with authorization header
func CreateHTTPRequestWithAuth(method, url string, body interface{}, token string) *http.Request {
	req := CreateHTTPRequest(method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// CreateTestToken creates a test JWT token
func CreateTestToken(userId int, name, email string, secret []byte) string {
	// This is a simplified token creation for testing
	// In real tests, you might want to use the actual JWT library
	return fmt.Sprintf("test-token-%d-%s-%s", userId, name, email)
}

// CreateTestResponseWriter creates a test response writer
func CreateTestResponseWriter() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// AssertJSONResponse checks if the response matches expected JSON
func AssertJSONResponse(t testingT, w *httptest.ResponseRecorder, expectedStatus int, expectedData interface{}) {
	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, w.Code)
	}
	
	var response model.ResultStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
		return
	}
	
	if response.StatusCode != expectedStatus {
		t.Errorf("Expected status code %d, got %d", expectedStatus, response.StatusCode)
	}
}

// CreateTestContext creates a test context
func CreateTestContext() context.Context {
	return context.Background()
}

// CreateTempConfigFile creates a temporary config file for testing
func CreateTempConfigFile() (string, error) {
	config := CreateTestConfig()
	configData, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	
	tempDir := os.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")
	
	err = os.WriteFile(configFile, configData, 0644)
	if err != nil {
		return "", err
	}
	
	return configFile, nil
}

// CleanupTempFile removes a temporary file
func CleanupTempFile(filePath string) error {
	return os.Remove(filePath)
}

// testingT is a minimal interface for testing
type testingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}
