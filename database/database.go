package database

import (
	"context"
	"fmt"
	"ncap-api-server/model"
)

type Database interface {
	RegisterUser(ctx context.Context, model *model.UserModel) (int64, int, error)
	LoginUser(ctx context.Context, model *model.UserModel) (int, int, error)

	CreateNewProject(ctx context.Context, model *model.ProjectModel, projectLimit int) (int, error)
	CreateTable(ctx context.Context, model *model.CreateTableModel) (int, error)
	InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) error
	DropTable(ctx context.Context, model *model.DropTableModel) error
	InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error)
	SelectTableListFromMetaData(ctx context.Context, projectId string) ([]string, int, error)
	RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error)
	DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) error
	RetrieveProjectList(ctx context.Context, userId int) ([]*model.ProjectModel, int, error)
}

func NewDatabase(config *model.DatabaseConfig) (Database, error) {
	switch config.Type {
	case "MySQL":
		return NewMySQLConnection(config)
	default:
		return nil, fmt.Errorf("database type %s not supported", config.Type)
	}

}
