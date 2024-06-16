package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"ncap-api-server/model"
)

type MySql struct {
	db *sql.DB
}

func NewMySQLConnection(config *model.DatabaseConfig) (Database, error) {
	db, err := sql.Open("mysql", config.Connection.URI)

	//db, err := sql.Open("mysql", "newuser:password@tcp(localhost:3306)/ncap")
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

func (mysql *MySql) CreateTable(ctx context.Context, model *model.CreateTableModel) error {
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

func (mysql *MySql) InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) error {
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

func (mysql *MySql) DropTable(ctx context.Context, model *model.DropTableModel) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", model.TableName)
	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// TODO: drop list from meta-data
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

func (mysql *MySql) SelectTableListFromMetaData(ctx context.Context, projectId string) ([]string, error) {
	query := "SELECT DISPLAY_NAME FROM project_metadata WHERE project_id = ?"
	tableNameList := []string{}

	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return tableNameList, err
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return tableNameList, err
		}

		tableNameList = append(tableNameList, tableName)

	}
	return tableNameList, nil
}

func (mysql *MySql) RetrieveDataFromTable(ctx context.Context, model *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, error) {
	// Retrieve the table name

	query := "SELECT TABLE_NAME FROM project_metadata where project_id = ? and display_name = ?"

	tableName := ""
	row := mysql.db.QueryRowContext(ctx, query, model.ProjectId, model.DisplayName)

	if err := row.Scan(&tableName); err != nil {
		return nil, nil, err
	}

	columns := []string{}
	rowData := []map[string]interface{}{}

	query = fmt.Sprintf("SELECT * FROM %s", tableName)

	rows, err := mysql.db.QueryContext(ctx, query)
	if err != nil {
		return columns, rowData, err
	}

	columns, err = rows.Columns()
	if err != nil {
		return columns, rowData, err
	}
	results := []map[string]interface{}{}
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]

			byteVal, ok := val.([]byte)
			if ok {
				rowMap[colName] = string(byteVal)
			} else {
				rowMap[colName] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, results, nil
}

func (mysql *MySql) DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) error {
	query := "SELECT TABLE_NAME FROM project_metadata where project_id = ? and display_name = ?"

	tableName := ""
	row := mysql.db.QueryRowContext(ctx, query, model.ProjectId, model.DisplayName)

	if err := row.Scan(&tableName); err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)

	_, err := mysql.db.ExecContext(ctx, query, model.TableRowId)
	if err != nil {
		return err
	}

	return nil
}

func (mysql *MySql) RegisterUser(ctx context.Context, model *model.UserModel) error {
	query := "INSERT INTO users (name, email, password) value (?, ?, ?)"

	_, err := mysql.db.ExecContext(ctx, query, model.Name, model.Email, model.Password)
	if err != nil {
		return err
	}
	return nil
}

func (mysql *MySql) LoginUser(ctx context.Context, model *model.UserModel) (int, error) {
	query := "SELECT id, email, password FROM users where email = ?"

	userId, email, password := -1, "", ""
	row := mysql.db.QueryRowContext(ctx, query, model.Email)

	if err := row.Scan(&userId, &email, &password); err != nil {
		return -1, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(model.Password))
	if err != nil {
		return -1, err
	}

	return userId, nil

}
