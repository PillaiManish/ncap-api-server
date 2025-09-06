package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"ncap-api-server/testutils"
	"ncap-api-server/model"
	"ncap-api-server/common"
	"github.com/sirupsen/logrus"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("NewApiServerModel", func() {
	It("should create API server model", func() {
		ctx := context.Background()
		logger := logrus.New()
		mockDB := &testutils.MockDatabase{}
		config := testutils.CreateTestConfig()
		
		apiServer := NewApiServerModel(ctx, logger, mockDB, config)
		Expect(apiServer).NotTo(BeNil())
	})
})

var _ = Describe("JwtClaims", func() {
	It("should create valid JWT claims", func() {
		claims := &JwtClaims{
			UserID:   1,
			Email:    "test@example.com",
			Password: "password",
			Name:     "Test User",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}
		
		Expect(claims.UserID).To(Equal(1))
		Expect(claims.Email).To(Equal("test@example.com"))
		Expect(claims.Password).To(Equal("password"))
		Expect(claims.Name).To(Equal("Test User"))
	})
})

var _ = Describe("APIServerModel", func() {
	It("should initialize all components", func() {
		ctx := context.Background()
		logger := logrus.New()
		mockDB := &testutils.MockDatabase{}
		config := testutils.CreateTestConfig()
		
		apiServer := &APIServerModel{
			logger:   logger,
			database: mockDB,
			config:   config,
			ctx:      ctx,
		}
		
		Expect(apiServer.logger).NotTo(BeNil())
		Expect(apiServer.database).NotTo(BeNil())
		Expect(apiServer.config).NotTo(BeNil())
		Expect(apiServer.ctx).NotTo(BeNil())
	})
})

var _ = Describe("Response", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("success response", func() {
		It("should return success response", func() {
			w := httptest.NewRecorder()
			payload := map[string]string{"message": "success"}
			
			apiServer.Response(w, http.StatusOK, payload)
			
			Expect(w.Code).To(Equal(http.StatusOK))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(common.SUCCESS))
			Expect(result.StatusCode).To(Equal(http.StatusOK))
		})
	})
	
	Context("bad request response", func() {
		It("should return bad request response", func() {
			w := httptest.NewRecorder()
			payload := "error message"
			
			apiServer.Response(w, http.StatusBadRequest, payload)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(common.FAILED))
		})
	})
	
	Context("unauthorized response", func() {
		It("should return unauthorized response", func() {
			w := httptest.NewRecorder()
			payload := "unauthorized"
			
			apiServer.Response(w, http.StatusUnauthorized, payload)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(common.FAILED))
		})
	})
	
	Context("internal server error response", func() {
		It("should return internal server error response", func() {
			w := httptest.NewRecorder()
			payload := "internal error"
			
			apiServer.Response(w, http.StatusInternalServerError, payload)
			
			Expect(w.Code).To(Equal(http.StatusInternalServerError))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(common.FAILED))
		})
	})
})

var _ = Describe("Authorization", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("missing authorization header", func() {
		It("should return error", func() {
			w := httptest.NewRecorder()
			
			userId, err := apiServer.authorization(w, "", "test")
			
			Expect(err).To(HaveOccurred())
			Expect(userId).To(Equal(-1))
		})
	})
	
	Context("invalid authorization header", func() {
		It("should return error", func() {
			w := httptest.NewRecorder()
			
			userId, err := apiServer.authorization(w, "InvalidToken", "test")
			
			Expect(err).To(HaveOccurred())
			Expect(userId).To(Equal(-1))
		})
	})
})

var _ = Describe("VerifyToken", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("invalid token", func() {
		It("should return error", func() {
			userId, err := apiServer.verifyToken("invalid-token")
			
			Expect(err).To(HaveOccurred())
			Expect(userId).To(Equal(-1))
		})
	})
})

var _ = Describe("CreateToken", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid token creation", func() {
		It("should create token successfully", func() {
			token, err := apiServer.createToken(1, "Test User", "test@example.com")
			
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("CheckPermission", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("successful permission check", func() {
		It("should grant permission", func() {
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return http.StatusOK, nil
			}
			
			hasPermission, code, err := apiServer.checkPermission(1, "project-1", common.ADMIN)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(hasPermission).To(BeTrue())
			Expect(code).To(Equal(http.StatusOK))
		})
	})
	
	Context("failed permission check", func() {
		It("should deny permission", func() {
			apiServer.database.(*testutils.MockDatabase).CheckPermissionFunc = func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
				return http.StatusNotFound, nil
			}
			
			hasPermission, code, err := apiServer.checkPermission(1, "project-1", common.ADMIN)
			
			Expect(err).NotTo(HaveOccurred())
			Expect(hasPermission).To(BeFalse())
			Expect(code).To(Equal(http.StatusNotFound))
		})
	})
})

var _ = Describe("CreateTableNameFromName", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("table name generation", func() {
		It("should create correct table name", func() {
			tableName := apiServer.createTableNameFromName("test-table", 1)
			expected := "1_test_table"
			
			Expect(tableName).To(Equal(expected))
		})
	})
	
	Context("table name with spaces", func() {
		It("should replace spaces with underscores", func() {
			tableName := apiServer.createTableNameFromName("test table", 1)
			expected := "1_test_table"
			
			Expect(tableName).To(Equal(expected))
		})
	})
})

var _ = Describe("Stop", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
		apiServer.database.(*testutils.MockDatabase).CloseFunc = func() error {
			return nil
		}
	})
	
	It("should stop API server without panic", func() {
		Expect(func() { apiServer.Stop() }).NotTo(Panic())
	})
})

// Helper function to create a test APIServer
func createTestAPIServer() *APIServerModel {
	ctx := context.Background()
	logger := logrus.New()
	mockDB := &testutils.MockDatabase{}
	config := testutils.CreateTestConfig()
	
	return &APIServerModel{
		logger:   logger,
		database: mockDB,
		config:   config,
		ctx:      ctx,
	}
}
