package database

import (
	"context"
	"fmt"
	"ncap-api-server/common"
	"ncap-api-server/model"
)

type Database interface {
	RegisterUser(ctx context.Context, model *model.UserModel) (int64, int, error)
	LoginUser(ctx context.Context, model *model.UserModel) (*model.UserModel, int, error)

	CreateNewProject(ctx context.Context, model *model.ProjectModel, projectLimit int) (int, error)
	CreateTable(ctx context.Context, projectId int, model *model.CreateTableModel) (int, error)
	InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) (int, error)
	UpdateTableData(ctx context.Context, model *model.UpdateTableDataModel) (int, error)
	DropTable(ctx context.Context, model *model.DropTableModel) (int, error)
	InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error)
	RetrieveTableListFromMetaData(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error)
	RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, []string, []*model.ForeignKey, int, error)
	DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) (int, error)
	RetrieveProjectList(ctx context.Context, userId int) (map[string][]*model.ProjectModel, int, error)
	Close() error

	// v2
	EditProject(ctx context.Context, model *model.ProjectModel) (int, error)
	CheckPermission(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error)
	ProjectAccessRequest(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error)
	ProjectAccessStatus(ctx context.Context, userId int, projectId string) (common.STATUS, int, error)
	ProjectAccessList(ctx context.Context, userId int, projectId string) ([]*model.ProjectAccessRequestModel, int, error)
	ProjectAccessAction(ctx context.Context, projectAccessAction *model.ProjectAccessRequestModel, projectId string) (int, error)
	CheckProjectExists(ctx context.Context, projectId string) (bool, int, error)
}

func NewDatabase(config *model.DatabaseConfig) (Database, error) {
	switch config.Type {
	case common.MYSQL:
		return NewMySQLConnection(config)
	default:
		return nil, fmt.Errorf("database type %s not supported", config.Type)
	}
}

func MapDataType(dataType, databaseType string) (string, error) {
	switch databaseType {
	case common.MYSQL:
		switch dataType {
		case "text":
			return "varchar(250)", nil
		case "boolean":
			return "tinyint(1)", nil
		case "int":
			return "int", nil
		case "datetime":
			return "DATETIME", nil
		default:
			return "", fmt.Errorf("this datatype is not supported")
		}
	default:
		return "", fmt.Errorf("database type %s not supported", databaseType)
	}

}
