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

func TestTable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Table Suite")
}

var _ = Describe("CreateNewTable", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid table creation", func() {
		It("should create table successfully", func() {
			// Mock successful table creation
			apiServer.database.(*testutils.MockDatabase).CreateTableFunc = func(ctx context.Context, table *model.CreateTableModel) (int, error) {
				return 201, nil
			}
			apiServer.database.(*testutils.MockDatabase).InsertIntoMetaDataFunc = func(ctx context.Context, table *model.CreateTableModel, projectId int) (int, error) {
				return 200, nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			tableData := map[string]interface{}{
				"name":        "test-table",
				"description": "Test table description",
				"columns": map[string]string{
					"name":  "text",
					"age":   "int",
					"email": "text",
				},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-table?project_id=1", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			Expect(w.Code).To(Equal(201))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			tableData := map[string]interface{}{
				"name":        "test-table",
				"description": "Test table description",
				"columns": map[string]string{
					"name": "text",
				},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/create-table?project_id=1", tableData)
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid project ID", func() {
		It("should return internal server error", func() {
			tableData := map[string]interface{}{
				"name":        "test-table",
				"description": "Test table description",
				"columns": map[string]string{
					"name": "text",
				},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-table?project_id=invalid", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/create-table?project_id=1", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			tableData := map[string]interface{}{
				"name":        "test-table",
				"description": "Test table description",
				"columns": map[string]string{
					"name": "text",
				},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-table?project_id=1", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("DropTable", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid table drop", func() {
		It("should drop table successfully", func() {
			// Mock successful table drop
			apiServer.database.(*testutils.MockDatabase).DropTableFunc = func(ctx context.Context, table *model.DropTableModel) error {
				return nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			tableData := map[string]string{
				"table_name": "1_test_table",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/drop-table", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DropTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			tableData := map[string]string{
				"table_name": "1_test_table",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/drop-table", tableData)
			w := httptest.NewRecorder()
			
			apiServer.DropTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/drop-table", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DropTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			tableData := map[string]string{
				"table_name": "1_test_table",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/drop-table", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.DropTable(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("RetrieveTableListFromMetaData", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid table list retrieval", func() {
		It("should retrieve table list successfully", func() {
			// Mock successful table list retrieval
			apiServer.database.(*testutils.MockDatabase).RetrieveTableListFromMetaDataFunc = func(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error) {
				tables := []*model.CreateTableModel{
					{DisplayName: "table1", TableName: "1_table1"},
					{DisplayName: "table2", TableName: "1_table2"},
				}
				return tables, 200, nil
			}
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			tableData := map[string]string{
				"project_id": "1",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-table-list", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveTableListFromMetaData(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal("SUCCESS"))
		})
	})
	
	Context("empty project ID", func() {
		It("should return bad request", func() {
			tableData := map[string]string{
				"project_id": "",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-table-list", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveTableListFromMetaData(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			tableData := map[string]string{
				"project_id": "1",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/retrieve-table-list", tableData)
			w := httptest.NewRecorder()
			
			apiServer.RetrieveTableListFromMetaData(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("no permission", func() {
		It("should return unauthorized", func() {
			// Mock no permission
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 404, nil
			}
			
			tableData := map[string]string{
				"project_id": "1",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/retrieve-table-list", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveTableListFromMetaData(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("Table Validation", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("table name validation", func() {
		It("should handle empty table name", func() {
			tableData := map[string]interface{}{
				"name":        "",
				"description": "Test table description",
				"columns": map[string]string{
					"name": "text",
				},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-table?project_id=1", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(201), Equal(400)))
		})
	})
	
	Context("columns validation", func() {
		It("should handle empty columns", func() {
			tableData := map[string]interface{}{
				"name":        "test-table",
				"description": "Test table description",
				"columns":     map[string]string{},
				"primary_keys": []string{"id"},
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-table?project_id=1", tableData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewTable(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(201), Equal(400)))
		})
	})
})
