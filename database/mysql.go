package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"ncap-api-server/model"
)

type MySql struct {
	db *sql.DB
}

func NewMySQLConnection() (Database, error) {
	db, err := sql.Open("mysql", "pillaimanish:password@tcp(localhost:3306)/ncap")
	if err != nil {
		return nil, err
	}

	return &MySql{db: db}, nil
}

func (mysql *MySql) CreateNewProject(ctx context.Context, projectId, projectName, projectDesc string) error {
	query := "INSERT INTO project_details VALUES (?, ?, ?)"

	_, err := mysql.db.ExecContext(ctx, query, projectId, projectName, projectDesc)
	if err != nil {
		return err
	}
	return nil
}

func (mysql *MySql) CreateTableForProject(ctx context.Context, model *model.CreateTableModel) error {
	// Query Generation

	query := fmt.Sprintf("CREATE TABLE %s", model.TableName)

	// For Columns
	var columnStmt string
	i := 0
	for columnName, columnType := range model.Columns {
		if i == 0 {
			columnStmt = fmt.Sprintf("%s %s", columnName, columnType)
		} else {
			columnStmt += "," + fmt.Sprintf("%s %s", columnName, columnType)
		}

		i += 1
	}

	// For Primary Key
	var primaryKeyStmt string
	for index, primaryKey := range model.PrimaryKeys {
		if index == 0 {
			primaryKeyStmt = primaryKey
		} else {
			primaryKeyStmt += ", " + primaryKey
		}
	}

	if query != "" {
		query = fmt.Sprintf("%s (%s, PRIMARY KEY (%s))", query, columnStmt, primaryKeyStmt)
	}

	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (mysql *MySql) InsertIntoTableForProject(ctx context.Context, model *model.InsertIntoTableModel) error {
	// Query Generation

	query := fmt.Sprintf("INSERT INTO %s", model.TableName)

	// For Columns
	var columnStmt string
	var columnValueStmt string

	i := 0
	for columnName, columnValue := range model.Columns {
		if i == 0 {
			columnStmt = fmt.Sprintf("%s", columnName)
			columnValueStmt = fmt.Sprintf(`"%s"`, columnValue)
		} else {

			columnStmt += "," + fmt.Sprintf("%s", columnName)
			columnValueStmt += "," + fmt.Sprintf(`"%s"`, columnValue)
		}

		i += 1
	}

	query += fmt.Sprintf(" (%s)", columnStmt)
	query += fmt.Sprintf(" VALUES (%s)", columnValueStmt)

	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (mysql *MySql) DropTableForProject(ctx context.Context, model *model.DropTableModel) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", model.TableName)
	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (mysql *MySql) InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel) error {
	query := "INSERT INTO project_metadata (project_id, display_name, table_name, metadata) values (?, ?, ?, ?)"

	data, err := json.Marshal(model)
	if err != nil {
		return err
	}

	_, err = mysql.db.ExecContext(ctx, query, 1, model.TableName, model.TableName, data)
	if err != nil {
		return err
	}
	return nil
}
