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
}

func NewDatabase(typeS string) (Database, error) {
	switch typeS {
	case "MySQL":
		return NewMySQLConnection()
	default:
		return nil, fmt.Errorf("database type %s not supported", typeS)
	}

}
