package database

import (
	"context"
	"fmt"
	"ncap-api-server/model"
)

type Database interface {
	RegisterUser(ctx context.Context, model *model.UserModel) error
	LoginUser(ctx context.Context, model *model.UserModel) error

	CreateNewProject(ctx context.Context, projectId, projectName, projectDesc string) error
	CreateTable(ctx context.Context, model *model.CreateTableModel) error
	InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) error
	DropTable(ctx context.Context, model *model.DropTableModel) error
	InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel) error
	SelectTableListFromMetaData(ctx context.Context, projectId string) ([]string, error)
	RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error)
	DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) error
}

func NewDatabase(config *model.DatabaseConfig) (Database, error) {
	switch config.Type {
	case "MySQL":
		return NewMySQLConnection(config)
	default:
		return nil, fmt.Errorf("database type %s not supported", config.Type)
	}

}
