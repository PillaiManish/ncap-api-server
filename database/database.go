package database

import (
	"context"
	"fmt"
	"ncap-api-server/model"
)

type Database interface {
	CreateNewProject(ctx context.Context, projectId, projectName, projectDesc string) error
	CreateTableForProject(ctx context.Context, model *model.CreateTableModel) error
	InsertIntoTableForProject(ctx context.Context, model *model.InsertIntoTableModel) error
	DropTableForProject(ctx context.Context, model *model.DropTableModel) error
	InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel) error
	SelectTableListFromMetaData(ctx context.Context, projectId string) ([]string, error)
	RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error)
	DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) error
}

func NewDatabase(typeS string) (Database, error) {
	switch typeS {
	case "MySQL":
		return NewMySQLConnection()
	default:
		return nil, fmt.Errorf("database type %s not supported", typeS)
	}

}
