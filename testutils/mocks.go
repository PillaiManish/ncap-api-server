package testutils

import (
	"context"
	"ncap-api-server/common"
	"ncap-api-server/model"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	// User operations
	RegisterUserFunc           func(ctx context.Context, model *model.UserModel) (int64, int, error)
	LoginUserFunc              func(ctx context.Context, model *model.UserModel) (*model.UserModel, int, error)
	
	// Project operations
	CreateNewProjectFunc       func(ctx context.Context, model *model.ProjectModel, projectLimit int) (int, error)
	EditProjectFunc            func(ctx context.Context, model *model.ProjectModel) (int, error)
	RetrieveProjectListFunc    func(ctx context.Context, userId int) ([]*model.ProjectModel, int, error)
	CheckPermissionFunc        func(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error)
	ProjectAccessRequestFunc   func(ctx context.Context, userId int, projectId string) error
	
	// Table operations
	CreateTableFunc                    func(ctx context.Context, model *model.CreateTableModel) (int, error)
	InsertIntoTableFunc                func(ctx context.Context, model *model.InsertIntoTableModel) error
	DropTableFunc                      func(ctx context.Context, model *model.DropTableModel) error
	InsertIntoMetaDataFunc             func(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error)
	RetrieveTableListFromMetaDataFunc  func(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error)
	RetrieveDataFromTableFunc          func(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error)
	DeleteEntryFromTableFunc           func(ctx context.Context, model *model.DeleteEntryFromTableModel) error
	
	CloseFunc                  func() error
}

// User operations
func (m *MockDatabase) RegisterUser(ctx context.Context, model *model.UserModel) (int64, int, error) {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, model)
	}
	return 0, 0, nil
}

func (m *MockDatabase) LoginUser(ctx context.Context, model *model.UserModel) (*model.UserModel, int, error) {
	if m.LoginUserFunc != nil {
		return m.LoginUserFunc(ctx, model)
	}
	return model, 200, nil
}

// Project operations
func (m *MockDatabase) CreateNewProject(ctx context.Context, model *model.ProjectModel, projectLimit int) (int, error) {
	if m.CreateNewProjectFunc != nil {
		return m.CreateNewProjectFunc(ctx, model, projectLimit)
	}
	return 201, nil
}

func (m *MockDatabase) EditProject(ctx context.Context, model *model.ProjectModel) (int, error) {
	if m.EditProjectFunc != nil {
		return m.EditProjectFunc(ctx, model)
	}
	return 200, nil
}

func (m *MockDatabase) RetrieveProjectList(ctx context.Context, userId int) ([]*model.ProjectModel, int, error) {
	if m.RetrieveProjectListFunc != nil {
		return m.RetrieveProjectListFunc(ctx, userId)
	}
	return []*model.ProjectModel{}, 200, nil
}

func (m *MockDatabase) CheckPermission(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(ctx, userId, projectId, role)
	}
	return 200, nil
}

func (m *MockDatabase) ProjectAccessRequest(ctx context.Context, userId int, projectId string) error {
	if m.ProjectAccessRequestFunc != nil {
		return m.ProjectAccessRequestFunc(ctx, userId, projectId)
	}
	return nil
}

// Table operations
func (m *MockDatabase) CreateTable(ctx context.Context, model *model.CreateTableModel) (int, error) {
	if m.CreateTableFunc != nil {
		return m.CreateTableFunc(ctx, model)
	}
	return 201, nil
}

func (m *MockDatabase) InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) error {
	if m.InsertIntoTableFunc != nil {
		return m.InsertIntoTableFunc(ctx, model)
	}
	return nil
}

func (m *MockDatabase) DropTable(ctx context.Context, model *model.DropTableModel) error {
	if m.DropTableFunc != nil {
		return m.DropTableFunc(ctx, model)
	}
	return nil
}

func (m *MockDatabase) InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error) {
	if m.InsertIntoMetaDataFunc != nil {
		return m.InsertIntoMetaDataFunc(ctx, model, projectId)
	}
	return 200, nil
}

func (m *MockDatabase) RetrieveTableListFromMetaData(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error) {
	if m.RetrieveTableListFromMetaDataFunc != nil {
		return m.RetrieveTableListFromMetaDataFunc(ctx, projectId)
	}
	return []*model.CreateTableModel{}, 200, nil
}

func (m *MockDatabase) RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error) {
	if m.RetrieveDataFromTableFunc != nil {
		return m.RetrieveDataFromTableFunc(ctx, model)
	}
	return []string{}, []map[string]interface{}{}, nil
}

func (m *MockDatabase) DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) error {
	if m.DeleteEntryFromTableFunc != nil {
		return m.DeleteEntryFromTableFunc(ctx, model)
	}
	return nil
}

func (m *MockDatabase) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
