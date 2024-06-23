package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	mysqlLib "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"ncap-api-server/model"
	"net/http"
	"os"
)

type MySql struct {
	db *sql.DB
}

func registerTLSConfig(caPath string) error {
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA cert: %w", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("failed to append CA certs")
	}

	err = mysqlLib.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewMySQLConnection(config *model.DatabaseConfig) (Database, error) {
	uri := ""
	connectionDetails := config.Connection

	if config.Connection.URI == nil {
		uri = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?parseTime=true", connectionDetails.User, connectionDetails.Pass, connectionDetails.Host, connectionDetails.Port, connectionDetails.DBName)
	} else {
		uri = *config.Connection.URI
	}

	if config.SSL != nil && config.SSL.Enabled {
		err := registerTLSConfig(config.SSL.CAPath)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("mysql", uri)
	//db.
	//db, err := sql.Open("mysql", "newuser:password@tcp(localhost:3306)/ncap")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &MySql{db: db}, nil
}

func (mysql *MySql) CreateNewProject(ctx context.Context, model *model.ProjectModel, projectLimit int) error {
	query := "SELECT COUNT(id) FROM project_details where user_id = ?"

	rowCount := 0
	row := mysql.db.QueryRowContext(ctx, query, model.UserId)

	if err := row.Scan(&rowCount); err != nil {
		return err
	}

	if rowCount >= projectLimit {
		return fmt.Errorf("project limit exceeded")
	}

	query = "INSERT INTO project_details (user_id, name, description) VALUES (?, ?, ?)"

	_, err := mysql.db.ExecContext(ctx, query, model.UserId, model.Name, model.Desc)
	if err != nil {
		return err
	}
	return nil
}

func (mysql *MySql) CreateTable(ctx context.Context, model *model.CreateTableModel) (int, error) {
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
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
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

func (mysql *MySql) InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error) {
	query := "INSERT INTO project_metadata (project_id, display_name, table_name, metadata) values (?, ?, ?, ?)"

	data, err := json.Marshal(model)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	_, err = mysql.db.ExecContext(ctx, query, projectId, model.DisplayName, model.TableName, data)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (mysql *MySql) SelectTableListFromMetaData(ctx context.Context, projectId string) ([]string, int, error) {
	query := "SELECT DISPLAY_NAME FROM project_metadata WHERE project_id = ?"
	tableNameList := []string{}

	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return tableNameList, http.StatusInternalServerError, err
	}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return tableNameList, http.StatusInternalServerError, err
		}

		tableNameList = append(tableNameList, tableName)

	}
	return tableNameList, http.StatusOK, nil
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

func (mysql *MySql) RegisterUser(ctx context.Context, model *model.UserModel) (int64, int, error) {
	query := "INSERT INTO users (name, email, password) value (?, ?, ?)"

	res, err := mysql.db.ExecContext(ctx, query, model.Name, model.Email, model.Password)
	if err != nil {
		var mysqlErr *mysqlLib.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return -1, http.StatusConflict, fmt.Errorf("user already exists, please log in")
		}
		return -1, http.StatusInternalServerError, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, http.StatusInternalServerError, fmt.Errorf("error registration user")
	}
	return id, http.StatusCreated, nil
}

func (mysql *MySql) LoginUser(ctx context.Context, model *model.UserModel) (int, int, error) {
	query := "SELECT id, email, password FROM users where email = ?"

	userId, email, password := -1, "", ""
	row := mysql.db.QueryRowContext(ctx, query, model.Email)

	if err := row.Scan(&userId, &email, &password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, http.StatusNotFound, fmt.Errorf("user not registered")
		}
		return -1, http.StatusInternalServerError, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(model.Password))
	if err != nil {
		return -1, http.StatusUnauthorized, fmt.Errorf("invalid password")
	}

	return userId, http.StatusAccepted, nil

}

func (mysql *MySql) RetrieveProjectList(ctx context.Context, userId int) ([]*model.ProjectModel, int, error) {
	query := "SELECT id, name, description FROM project_details where user_id = ?"

	rows, err := mysql.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	projectList := []*model.ProjectModel{}

	for rows.Next() {
		project := &model.ProjectModel{}
		err := rows.Scan(&project.Id, &project.Name, &project.Desc)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		projectList = append(projectList, project)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return projectList, http.StatusOK, nil
}
