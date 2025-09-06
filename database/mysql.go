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
	"ncap-api-server/common"
	"ncap-api-server/model"
	"net/http"
	"os"
	"strings"
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

func (mysql *MySql) CreateNewProject(ctx context.Context, model *model.ProjectModel, projectLimit int) (int, error) {
	tx, err := mysql.db.BeginTx(ctx, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	query := "SELECT COUNT(id) FROM project_details where user_id = ?"

	rowCount := 0
	row := tx.QueryRowContext(ctx, query, model.UserId)

	if err = row.Scan(&rowCount); err != nil {
		return http.StatusInternalServerError, err
	}

	if rowCount >= projectLimit {
		return http.StatusConflict, fmt.Errorf("project limit exceeded")
	}

	query = "INSERT INTO project_details (user_id, name, description) VALUES (?, ?, ?)"

	res, err := tx.ExecContext(ctx, query, model.UserId, model.Name, model.Desc)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	query = "INSERT INTO users_projects (user_id, project_id, role, status) VALUES (?, ?, ?, ?)"
	_, err = tx.ExecContext(ctx, query, model.UserId, lastInsertId, common.OWNER, common.APPROVED)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = tx.Commit()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}
func (mysql *MySql) CreateTable(ctx context.Context, model *model.CreateTableModel) (int, error) {
	// Base query
	query := fmt.Sprintf("CREATE TABLE %s", model.TableName)

	// Start with auto-incrementing ID column
	columnStmt := "`id` INT AUTO_INCREMENT"

	// Add additional columns
	for columnName, columnType := range model.Columns {
		columnStmt += fmt.Sprintf(", `%s` %s", columnName, MapDataType(columnType, common.MYSQL))
	}

	// Build primary key clause
	var primaryKeyStmt string
	if len(model.PrimaryKeys) == 0 {
		// No primary key specified, use the default `id`
		primaryKeyStmt = "`id`"
	} else {
		for index, primaryKey := range model.PrimaryKeys {
			if index > 0 {
				primaryKeyStmt += ", "
			}
			primaryKeyStmt += fmt.Sprintf("`%s`", primaryKey)
		}
	}

	// Finalize query
	query = fmt.Sprintf("%s (%s, PRIMARY KEY (%s))", query, columnStmt, primaryKeyStmt)

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

func (mysql *MySql) UpdateTableData(ctx context.Context, model *model.UpdateTableDataModel) error {
	query := fmt.Sprintf("UPDATE %s SET ", model.TableName)

	var values []interface{}
	var setStatements []string

	for columnName, columnValue := range model.Columns {
		setStatements = append(setStatements, fmt.Sprintf("%s = ?", columnName))
		values = append(values, columnValue)
	}

	query += strings.Join(setStatements, ", ")
	query += " WHERE id = ?"
	values = append(values, model.Id)

	_, err := mysql.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}
	return nil
}

func (mysql *MySql) DropTable(ctx context.Context, model *model.DropTableModel) error {
	tx, err := mysql.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", model.TableName)
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM project_metadata WHERE table_name = ?")
	_, err = tx.ExecContext(ctx, query, model.TableName)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (mysql *MySql) InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error) {
	query := "INSERT INTO project_metadata (project_id, display_name, table_name, metadata, description) values (?, ?, ?, ?, ?)"

	data, err := json.Marshal(model)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	_, err = mysql.db.ExecContext(ctx, query, projectId, model.DisplayName, model.TableName, data, model.Description)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (mysql *MySql) RetrieveTableListFromMetaData(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error) {
	query := "SELECT METADATA FROM project_metadata WHERE project_id = ?"
	tables := []*model.CreateTableModel{}

	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	defer rows.Close()

	for rows.Next() {
		metaData := ""
		if err := rows.Scan(&metaData); err != nil {
			return nil, http.StatusInternalServerError, err
		}

		retrieveTableList := &model.CreateTableModel{}
		err := json.Unmarshal([]byte(metaData), retrieveTableList)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		tables = append(tables, retrieveTableList)

	}
	return tables, http.StatusOK, nil
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

	defer rows.Close()

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

func (mysql *MySql) LoginUser(ctx context.Context, userModel *model.UserModel) (*model.UserModel, int, error) {
	query := "SELECT id, password, name FROM users where email = ?"

	row := mysql.db.QueryRowContext(ctx, query, userModel.Email)

	hashedPassword := ""
	if err := row.Scan(&userModel.Id, &hashedPassword, &userModel.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userModel, http.StatusNotFound, fmt.Errorf("user not registered")
		}
		return userModel, http.StatusInternalServerError, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(userModel.Password))
	if err != nil {
		return userModel, http.StatusUnauthorized, fmt.Errorf("invalid password")
	}

	return userModel, http.StatusAccepted, nil

}

func (mysql *MySql) RetrieveProjectList(ctx context.Context, userId int) ([]*model.ProjectModel, int, error) {
	query := "SELECT id, name, description FROM project_details where user_id = ?"

	rows, err := mysql.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	projectList := []*model.ProjectModel{}

	defer rows.Close()

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

func (mysql *MySql) Close() error {
	err := mysql.db.Close()
	if err != nil {
		return err
	}

	return nil
}

// v2
func (mysql *MySql) EditProject(ctx context.Context, model *model.ProjectModel) (int, error) {
	query := "UPDATE project_details SET name = ?, description = ? WHERE id = ? and user_id = ?"

	_, err := mysql.db.ExecContext(ctx, query, model.Name, model.Desc, model.Id, model.UserId)
	if err != nil {
		var mysqlErr *mysqlLib.MySQLError
		if errors.As(err, &mysqlErr) {
			return http.StatusInternalServerError, err
		}
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (mysql *MySql) CheckPermission(ctx context.Context, userId int, projectId string, minRole common.PERMISSION) (int, error) {
	query := "SELECT user_id, role FROM users_projects WHERE user_id = ? and project_id = ? and status = ?"

	row := mysql.db.QueryRowContext(ctx, query, userId, projectId, common.APPROVED)
	var id int
	var currentRole common.PERMISSION

	err := row.Scan(&id, &currentRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return http.StatusNotFound, fmt.Errorf("no permission")
		}
		return http.StatusInternalServerError, err
	}

	if currentRole < minRole {
		return http.StatusForbidden, fmt.Errorf("insufficient permission")
	}

	return http.StatusOK, nil
}

func (mysql *MySql) ProjectAccessRequest(ctx context.Context, userId int, projectId string, role common.PERMISSION) error {
	// Validate inputs
	if userId <= 0 || projectId == "" {
		return fmt.Errorf("invalid user ID or project ID")
	}
	query := "INSERT INTO users_projects (user_id, project_id, role) VALUES (?, ?, ?)"
	_, err := mysql.db.ExecContext(ctx, query, userId, projectId, role)
	if err != nil {
		var mysqlErr *mysqlLib.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			//TODO: return proper status code and not internal error
			return fmt.Errorf("access request already exists")
		}
		return err
	}
	return nil
}

func (mysql *MySql) ProjectAccessStatus(ctx context.Context, userId int, projectId string) (common.STATUS, error) {
	// Validate inputs
	if userId <= 0 || projectId == "" {
		return common.PENDING, fmt.Errorf("invalid user ID or project ID")
	}

	query := "SELECT status FROM users_projects WHERE user_id = ? and project_id = ?"
	row := mysql.db.QueryRowContext(ctx, query, userId, projectId)
	var status common.STATUS
	err := row.Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.PENDING, fmt.Errorf("no access request")
		}
		return common.PENDING, err
	}
	return status, nil
}

func (mysql *MySql) ProjectAccessList(ctx context.Context, userId int, projectId string) ([]*model.ProjectAccessRequestModel, error) {
	if userId <= 0 || projectId == "" {
		return nil, fmt.Errorf("invalid user ID or project ID")
	}

	query := "SELECT user_id, role, status FROM users_projects WHERE project_id = ?"
	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projectAccessList []*model.ProjectAccessRequestModel

	for rows.Next() {
		projectAccess := &model.ProjectAccessRequestModel{}
		var role common.PERMISSION
		var status common.STATUS

		// TODO: add something to get name from user id
		err = rows.Scan(&projectAccess.UserId, &role, &status)
		if err != nil {
			return nil, err
		}

		projectAccess.Status = status.String()
		projectAccess.Role = role.String()

		if projectAccess.UserId != userId {
			projectAccessList = append(projectAccessList, projectAccess)

		}

	}

	return projectAccessList, nil
}

func (mysql *MySql) ProjectAccessAction(ctx context.Context, projectAccessAction *model.ProjectAccessRequestModel, projectId string) error {
	if projectAccessAction.UserId <= 0 || projectId == "" {
		return fmt.Errorf("invalid user ID or project ID")
	}

	var action common.STATUS
	switch projectAccessAction.Action {
	case "approve":
		action = common.APPROVED
	case "reject":
		action = common.REJECTED
	default:
		return fmt.Errorf("invalid status")
	}

	query := "UPDATE users_projects SET status = ? WHERE user_id = ? AND project_id = ?"
	_, err := mysql.db.ExecContext(ctx, query, action, projectAccessAction.UserId, projectId)
	if err != nil {
		var mysqlErr *mysqlLib.MySQLError
		if errors.As(err, &mysqlErr) {
			return fmt.Errorf("database error: %w", err)
		}
		return err
	}

	return nil
}

func (mysql *MySql) CheckProjectExists(ctx context.Context, projectId string) (bool, error) {
	query := "SELECT id FROM project_details WHERE id = ?"
	row := mysql.db.QueryRowContext(ctx, query, projectId)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
