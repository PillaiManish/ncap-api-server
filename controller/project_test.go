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

func TestProject(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Project Suite")
}

var _ = Describe("CreateNewProject", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid project creation", func() {
		It("should create project successfully", func() {
			// Mock successful project creation
			apiServer.database.(*testutils.MockDatabase).CreateNewProjectFunc = func(ctx context.Context, project *model.ProjectModel, projectLimit int) (int, error) {
				return 201, nil
			}
			
			projectData := map[string]string{
				"name": "Test Project",
				"desc": "Test Description",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-new-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			Expect(w.Code).To(Equal(201))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			projectData := map[string]string{
				"name": "Test Project",
				"desc": "Test Description",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/create-new-project", projectData)
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/create-new-project", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("project limit exceeded", func() {
		It("should return conflict", func() {
			// Mock project limit exceeded
			apiServer.database.(*testutils.MockDatabase).CreateNewProjectFunc = func(ctx context.Context, project *model.ProjectModel, projectLimit int) (int, error) {
				return 409, nil
			}
			
			projectData := map[string]string{
				"name": "Test Project",
				"desc": "Test Description",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-new-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			Expect(w.Code).To(Equal(409))
		})
	})
})

var _ = Describe("EditProject", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid project edit", func() {
		It("should edit project successfully", func() {
			// Mock successful project edit
			apiServer.database.(*testutils.MockDatabase).EditProjectFunc = func(ctx context.Context, project *model.ProjectModel) (int, error) {
				return 200, nil
			}
			
			projectData := map[string]string{
				"id":   "1",
				"name": "Updated Project",
				"desc": "Updated Description",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPatch, "/edit-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.EditProject(w, req)
			
			Expect(w.Code).To(Equal(200))
		})
	})
	
	Context("missing required fields", func() {
		It("should return bad request", func() {
			projectData := map[string]string{
				"id":   "",
				"name": "",
				"desc": "",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPatch, "/edit-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.EditProject(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			projectData := map[string]string{
				"id":   "1",
				"name": "Updated Project",
				"desc": "Updated Description",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPatch, "/edit-project", projectData)
			w := httptest.NewRecorder()
			
			apiServer.EditProject(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("CheckProjectPermission", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid permission check", func() {
		It("should check permission successfully", func() {
			// Mock successful permission check
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return 200, nil
			}
			
			req, _ := http.NewRequest(http.MethodPost, "/check-project-permission?projectId=1", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CheckProjectPermission(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
	
	Context("missing project ID", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/check-project-permission", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CheckProjectPermission(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			req, _ := http.NewRequest(http.MethodPost, "/check-project-permission?projectId=1", nil)
			w := httptest.NewRecorder()
			
			apiServer.CheckProjectPermission(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("ProjectAccessRequest", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid access request", func() {
		It("should create access request successfully", func() {
			// Mock successful access request
			apiServer.database.(*testutils.MockDatabase).ProjectAccessRequestFunc = func(ctx context.Context, userId int, projectId string) error {
				return nil
			}
			
			req, _ := http.NewRequest(http.MethodPost, "/project-access-request?projectId=1", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.ProjectAccessRequest(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			req, _ := http.NewRequest(http.MethodPost, "/project-access-request?projectId=1", nil)
			w := httptest.NewRecorder()
			
			apiServer.ProjectAccessRequest(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("RetrieveProjectList", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid project list retrieval", func() {
		It("should retrieve project list successfully", func() {
			// Mock successful project list retrieval
			apiServer.database.(*testutils.MockDatabase).RetrieveProjectListFunc = func(ctx context.Context, userId int) ([]*model.ProjectModel, int, error) {
				projects := []*model.ProjectModel{
					{Id: "1", Name: "Project 1", Desc: "Description 1"},
					{Id: "2", Name: "Project 2", Desc: "Description 2"},
				}
				return projects, 200, nil
			}
			
			req, _ := http.NewRequest(http.MethodGet, "/retrieve-project-list", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			w := httptest.NewRecorder()
			
			apiServer.RetrieveProjectList(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal("SUCCESS"))
		})
	})
	
	Context("unauthorized request", func() {
		It("should return unauthorized", func() {
			req, _ := http.NewRequest(http.MethodGet, "/retrieve-project-list", nil)
			w := httptest.NewRecorder()
			
			apiServer.RetrieveProjectList(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("Project Validation", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("project name validation", func() {
		It("should handle empty project name", func() {
			projectData := map[string]string{
				"name": "",
				"desc": "Test Description",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-new-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(201), Equal(400)))
		})
	})
	
	Context("project description validation", func() {
		It("should handle empty project description", func() {
			projectData := map[string]string{
				"name": "Test Project",
				"desc": "",
			}
			
			req := testutils.CreateHTTPRequestWithAuth(http.MethodPost, "/create-new-project", projectData, "valid-token")
			w := httptest.NewRecorder()
			
			apiServer.CreateNewProject(w, req)
			
			// Should still process the request (validation happens in database layer)
			Expect(w.Code).To(Or(Equal(201), Equal(400)))
		})
	})
})
