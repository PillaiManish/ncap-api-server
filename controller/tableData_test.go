package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"ncap-api-server/testutils"
	"ncap-api-server/model"
	"ncap-api-server/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTableData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TableData Suite")
}

var _ = Describe("InsertIntoTable", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid insert", func() {
		It("should insert data successfully", func() {
			// Mock successful insert
			apiServer.database.(*testutils.MockDatabase).InsertIntoTableFunc = func(ctx context.Context, table *model.InsertIntoTableModel) error {
				return nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			insertData := map[string]interface{}{
				"table_name": "1_test_table",
				"columns": map[string]string{
					"name":  "John Doe",
					"age":   "25",
					"email": "john@example.com",
				},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/insert-into-table", insertData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			insertData := map[string]interface{}{
				"table_name": "1_test_table",
				"columns": map[string]string{
					"name": "John Doe",
				},
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/insert-into-table", insertData)
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/insert-into-table", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			insertData := map[string]interface{}{
				"table_name": "1_test_table",
				"columns": map[string]string{
					"name": "John Doe",
				},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/insert-into-table", insertData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("RetrieveDataTable", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid data retrieval", func() {
		It("should retrieve data successfully", func() {
			// Mock successful data retrieval
			apiServer.database.(*testutils.MockDatabase).RetrieveDataFromTableFunc = func(ctx context.Context, table *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error) {
				columns := []string{"id", "name", "age", "email"}
				data := []map[string]interface{}{
					{"id": 1, "name": "John Doe", "age": 25, "email": "john@example.com"},
					{"id": 2, "name": "Jane Doe", "age": 30, "email": "jane@example.com"},
				}
				return columns, data, nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			retrieveData := map[string]string{
				"project_id":   "1",
				"display_name": "test-table",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-data-from-table", retrieveData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveDataTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal("SUCCESS"))
			Expect(result.Data).NotTo(BeNil())
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			retrieveData := map[string]string{
				"project_id":   "1",
				"display_name": "test-table",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/retrieve-data-from-table", retrieveData)
			w := httptest.NewRecorder()
			
			apiServer.RetrieveDataTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/retrieve-data-from-table", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveDataTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			retrieveData := map[string]string{
				"project_id":   "1",
				"display_name": "test-table",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-data-from-table", retrieveData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveDataTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("DeleteEntryFromTable", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid delete", func() {
		It("should delete entry successfully", func() {
			// Mock successful delete
			apiServer.database.(*testutils.MockDatabase).DeleteEntryFromTableFunc = func(ctx context.Context, table *model.DeleteEntryFromTableModel) error {
				return nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			deleteData := map[string]interface{}{
				"project_id":   "1",
				"display_name": "test-table",
				"table_row_id": 1,
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/delete-from-table", deleteData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DeleteEntryFromTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			deleteData := map[string]interface{}{
				"project_id":   "1",
				"display_name": "test-table",
				"table_row_id": 1,
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/delete-from-table", deleteData)
			w := httptest.NewRecorder()
			
			apiServer.DeleteEntryFromTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/delete-from-table", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DeleteEntryFromTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			deleteData := map[string]interface{}{
				"project_id":   "1",
				"display_name": "test-table",
				"table_row_id": 1,
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/delete-from-table", deleteData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DeleteEntryFromTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("Table Data Validation", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("table name validation", func() {
		It("should handle empty table name", func() {
			insertData := map[string]interface{}{
				"table_name": "",
				"columns": map[string]string{
					"name": "John Doe",
				},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/insert-into-table", insertData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(200), Equal(400)))
		})
	})
	
	Context("columns validation", func() {
		It("should handle empty columns", func() {
			insertData := map[string]interface{}{
				"table_name": "1_test_table",
				"columns":    map[string]string{},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/insert-into-table", insertData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.InsertIntoTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(200), Equal(400)))
		})
	})
	
	Context("project ID validation", func() {
		It("should handle empty project ID", func() {
			retrieveData := map[string]string{
				"project_id":   "",
				"display_name": "test-table",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-data-from-table", retrieveData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveDataTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(200), Equal(400)))
		})
	})
	
	Context("table row ID validation", func() {
		It("should handle invalid table row ID", func() {
			deleteData := map[string]interface{}{
				"project_id":   "1",
				"display_name": "test-table",
				"table_row_id": 0, // Invalid ID
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/delete-from-table", deleteData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DeleteEntryFromTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(200), Equal(400)))
		})
	})
})
