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
	"golang.org/x/crypto/bcrypt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

var _ = Describe("Login", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid login", func() {
		It("should login successfully", func() {
			// Mock successful login
			apiServer.database.(*testutils.MockDatabase).LoginUserFunc = func(ctx context.Context, userModel *model.UserModel) (*model.UserModel, int, error) {
				userModel.Id = 1
				userModel.Name = "Test User"
				return userModel, 200, nil
			}
			
			loginData := map[string]string{
				"email":    "test@example.com",
				"password": "password",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/login", loginData)
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusOK))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal("SUCCESS"))
		})
	})
	
	Context("invalid credentials", func() {
		It("should return unauthorized", func() {
			// Mock failed login
			apiServer.database.(*testutils.MockDatabase).LoginUserFunc = func(ctx context.Context, userModel *model.UserModel) (*model.UserModel, int, error) {
				return userModel, 401, nil
			}
			
			loginData := map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/login", loginData)
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})
	
	Context("empty email password", func() {
		It("should return bad request", func() {
			loginData := map[string]string{
				"email":    "",
				"password": "",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/login", loginData)
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

var _ = Describe("Register", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("valid registration", func() {
		It("should register successfully", func() {
			// Mock successful registration
			apiServer.database.(*testutils.MockDatabase).RegisterUserFunc = func(ctx context.Context, userModel *model.UserModel) (int64, int, error) {
				return 1, 201, nil
			}
			
			registerData := map[string]string{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/register", registerData)
			w := httptest.NewRecorder()
			
			apiServer.Register(w, req)
			
			Expect(w.Code).To(Equal(http.StatusCreated))
			
			var result model.ResultStatus
			err := json.Unmarshal(w.Body.Bytes(), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal("SUCCESS"))
		})
	})
	
	Context("empty fields", func() {
		It("should return bad request", func() {
			registerData := map[string]string{
				"name":     "",
				"email":    "",
				"password": "",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/register", registerData)
			w := httptest.NewRecorder()
			
			apiServer.Register(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("user already exists", func() {
		It("should return conflict", func() {
			// Mock user already exists error
			apiServer.database.(*testutils.MockDatabase).RegisterUserFunc = func(ctx context.Context, userModel *model.UserModel) (int64, int, error) {
				return -1, 409, nil
			}
			
			registerData := map[string]string{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/register", registerData)
			w := httptest.NewRecorder()
			
			apiServer.Register(w, req)
			
			Expect(w.Code).To(Equal(http.StatusConflict))
		})
	})
	
	Context("invalid JSON", func() {
		It("should return bad request", func() {
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			apiServer.Register(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

var _ = Describe("Password Hashing", func() {
	Context("password hashing", func() {
		It("should hash password successfully", func() {
			password := "testpassword"
			
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			Expect(err).NotTo(HaveOccurred())
			
			// Verify the hashed password
			err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
			Expect(err).NotTo(HaveOccurred())
		})
	})
	
	Context("password verification", func() {
		It("should verify correct password", func() {
			password := "testpassword"
			wrongPassword := "wrongpassword"
			
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			Expect(err).NotTo(HaveOccurred())
			
			// Test correct password
			err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
			Expect(err).NotTo(HaveOccurred())
			
			// Test wrong password
			err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(wrongPassword))
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Auth Validation", func() {
	var apiServer *APIServerModel
	
	BeforeEach(func() {
		apiServer = createTestAPIServer()
	})
	
	Context("email validation", func() {
		It("should reject empty email", func() {
			loginData := map[string]string{
				"email":    "",
				"password": "password",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/login", loginData)
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
	
	Context("password validation", func() {
		It("should reject empty password", func() {
			loginData := map[string]string{
				"email":    "test@example.com",
				"password": "",
			}
			
			req := testutils.CreateHTTPRequest(http.MethodPost, "/login", loginData)
			w := httptest.NewRecorder()
			
			apiServer.Login(w, req)
			
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
