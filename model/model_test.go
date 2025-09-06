package model

import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Suite")
}

var _ = Describe("Config", func() {
	It("should create valid config", func() {
		config := &Config{
			ErrorDataResponse: true,
			LogLevel:          "DEBUG",
			AdminPanelUrl:     "*",
			Database: &DatabaseConfig{
				Type: "MySQL",
			},
			ProjectConfig: &ProjectConfig{
				Limit: 5,
			},
			JWT: &JWT{
				Secret: []byte("test-secret"),
			},
		}
		
		Expect(config.ErrorDataResponse).To(BeTrue())
		Expect(config.LogLevel).To(Equal("DEBUG"))
		Expect(config.AdminPanelUrl).To(Equal("*"))
		Expect(config.Database).NotTo(BeNil())
		Expect(config.ProjectConfig).NotTo(BeNil())
		Expect(config.JWT).NotTo(BeNil())
	})
})

var _ = Describe("ProjectConfig", func() {
	It("should create valid project config", func() {
		projectConfig := &ProjectConfig{
			Limit: 10,
		}
		
		Expect(projectConfig.Limit).To(Equal(10))
	})
})

var _ = Describe("DatabaseConfig", func() {
	It("should create valid database config", func() {
		dbConfig := &DatabaseConfig{
			Type: "MySQL",
			Connection: &ConnectionConfig{
				Host:   "localhost",
				Port:   3306,
				DBName: "testdb",
				User:   "user",
				Pass:   "pass",
			},
			SSL: &SSL{
				Enabled: true,
				CAPath:  "/path/to/ca.pem",
			},
		}
		
		Expect(dbConfig.Type).To(Equal("MySQL"))
		Expect(dbConfig.Connection).NotTo(BeNil())
		Expect(dbConfig.SSL).NotTo(BeNil())
	})
})

var _ = Describe("ConnectionConfig", func() {
	It("should create valid connection config", func() {
		connConfig := &ConnectionConfig{
			DBName: "testdb",
			Host:   "localhost",
			Port:   3306,
			User:   "user",
			Pass:   "pass",
		}
		
		Expect(connConfig.DBName).To(Equal("testdb"))
		Expect(connConfig.Host).To(Equal("localhost"))
		Expect(connConfig.Port).To(Equal(3306))
		Expect(connConfig.User).To(Equal("user"))
		Expect(connConfig.Pass).To(Equal("pass"))
	})
	
	It("should handle URI configuration", func() {
		uri := "user:pass@tcp(localhost:3306)/testdb"
		connConfig := &ConnectionConfig{
			URI:    &uri,
			DBName: "testdb",
			Host:   "localhost",
			Port:   3306,
			User:   "user",
			Pass:   "pass",
		}
		
		Expect(connConfig.URI).NotTo(BeNil())
		Expect(*connConfig.URI).To(Equal(uri))
	})
})

var _ = Describe("SSL", func() {
	It("should create valid SSL config", func() {
		ssl := &SSL{
			Enabled: true,
			CAPath:  "/path/to/ca.pem",
		}
		
		Expect(ssl.Enabled).To(BeTrue())
		Expect(ssl.CAPath).To(Equal("/path/to/ca.pem"))
	})
})

var _ = Describe("JWT", func() {
	It("should create valid JWT config", func() {
		secret := []byte("test-secret-key")
		jwt := &JWT{
			Secret: secret,
		}
		
		Expect(len(jwt.Secret)).To(Equal(len(secret)))
		Expect(string(jwt.Secret)).To(Equal(string(secret)))
	})
})

var _ = Describe("UserModel", func() {
	It("should create valid user model", func() {
		user := &UserModel{
			Id:       1,
			Email:    "test@example.com",
			Password: "hashedpassword",
			Name:     "Test User",
		}
		
		Expect(user.Id).To(Equal(1))
		Expect(user.Email).To(Equal("test@example.com"))
		Expect(user.Password).To(Equal("hashedpassword"))
		Expect(user.Name).To(Equal("Test User"))
	})
})

var _ = Describe("ProjectModel", func() {
	It("should create valid project model", func() {
		project := &ProjectModel{
			UserId: 1,
			Id:     "project-123",
			Name:   "Test Project",
			Desc:   "Test Description",
		}
		
		Expect(project.UserId).To(Equal(1))
		Expect(project.Id).To(Equal("project-123"))
		Expect(project.Name).To(Equal("Test Project"))
		Expect(project.Desc).To(Equal("Test Description"))
	})
})

var _ = Describe("ResultStatus", func() {
	It("should create valid result status", func() {
		result := &ResultStatus{
			Status:     "SUCCESS",
			StatusCode: 200,
			Data:       "test data",
			Message:    "Operation successful",
		}
		
		Expect(result.Status).To(Equal("SUCCESS"))
		Expect(result.StatusCode).To(Equal(200))
		Expect(result.Data).To(Equal("test data"))
		Expect(result.Message).To(Equal("Operation successful"))
	})
})

var _ = Describe("CreateTableModel", func() {
	It("should create valid table model", func() {
		table := &CreateTableModel{
			Description: "Test table description",
			DisplayName: "test-table",
			TableName:   "test_table",
			Columns: map[string]string{
				"name":  "text",
				"age":   "int",
				"email": "text",
			},
			PrimaryKeys: []string{"id", "email"},
		}
		
		Expect(table.Description).To(Equal("Test table description"))
		Expect(table.DisplayName).To(Equal("test-table"))
		Expect(table.TableName).To(Equal("test_table"))
		Expect(len(table.Columns)).To(Equal(3))
		Expect(len(table.PrimaryKeys)).To(Equal(2))
	})
	
	It("should handle columns correctly", func() {
		table := &CreateTableModel{
			Columns: map[string]string{
				"name":  "text",
				"age":   "int",
				"email": "text",
			},
		}
		
		Expect(table.Columns["name"]).To(Equal("text"))
		Expect(table.Columns["age"]).To(Equal("int"))
		Expect(table.Columns["email"]).To(Equal("text"))
	})
})

var _ = Describe("InsertIntoTableModel", func() {
	It("should create valid insert model", func() {
		insert := &InsertIntoTableModel{
			TableName: "test_table",
			Columns: map[string]string{
				"name":  "John Doe",
				"age":   "25",
				"email": "john@example.com",
			},
		}
		
		Expect(insert.TableName).To(Equal("test_table"))
		Expect(len(insert.Columns)).To(Equal(3))
		Expect(insert.Columns["name"]).To(Equal("John Doe"))
	})
})

var _ = Describe("DropTableModel", func() {
	It("should create valid drop table model", func() {
		drop := &DropTableModel{
			TableName: "test_table",
		}
		
		Expect(drop.TableName).To(Equal("test_table"))
	})
})

var _ = Describe("RetrieveTableListModel", func() {
	It("should create valid retrieve table list model", func() {
		retrieve := &RetrieveTableListModel{
			ProjectId: "project-123",
		}
		
		Expect(retrieve.ProjectId).To(Equal("project-123"))
	})
})

var _ = Describe("RetrieveDataFromTableModel", func() {
	It("should create valid retrieve data model", func() {
		retrieve := &RetrieveDataFromTableModel{
			ProjectId:   "project-123",
			DisplayName: "test-table",
		}
		
		Expect(retrieve.ProjectId).To(Equal("project-123"))
		Expect(retrieve.DisplayName).To(Equal("test-table"))
	})
})

var _ = Describe("DeleteEntryFromTableModel", func() {
	It("should create valid delete entry model", func() {
		delete := &DeleteEntryFromTableModel{
			ProjectId:   "project-123",
			DisplayName: "test-table",
			TableRowId:  1,
		}
		
		Expect(delete.ProjectId).To(Equal("project-123"))
		Expect(delete.DisplayName).To(Equal("test-table"))
		Expect(delete.TableRowId).To(Equal(1))
	})
})
