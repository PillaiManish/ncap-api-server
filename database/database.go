package database

import (
	"context"
	"fmt"
)

type Database interface {
	CreateNewProject(ctx context.Context, projectId, projectName, projectDesc string) error
}

func NewDatabase(typeS string) (Database, error) {
	switch typeS {
	case "MySQL":
		return NewMySQLConnection()
	default:
		return nil, fmt.Errorf("database type %s not supported", typeS)
	}

}
