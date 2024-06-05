package database

import "fmt"

type Database interface {
	CreateNewProject(projectName string) error
}

func NewDatabase(typeS string) (Database, error) {
	switch typeS {
	case "MySQL":
		return NewMySQLConnection(), nil
	default:
		return nil, fmt.Errorf("database type %s not supported", typeS)
	}

}
