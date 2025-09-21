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

// DatabaseError represents a structured database error
type DatabaseError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *DatabaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// handleDatabaseError processes MySQL errors and returns appropriate status codes and messages
func handleDatabaseError(err error) *DatabaseError {
	if err == nil {
		return nil
	}

	var mysqlErr *mysqlLib.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062: // Duplicate entry
			return &DatabaseError{
				StatusCode: http.StatusConflict,
				Message:    "Duplicate entry: " + mysqlErr.Message,
				Err:        err,
			}
		case 1451: // Cannot delete or update a parent row
			return &DatabaseError{
				StatusCode: http.StatusConflict,
				Message:    "Cannot delete or update record due to foreign key constraints",
				Err:        err,
			}
		case 1452: // Cannot add or update a child row
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Foreign key constraint violation",
				Err:        err,
			}
		case 1146: // Table doesn't exist
			return &DatabaseError{
				StatusCode: http.StatusNotFound,
				Message:    "Table not found",
				Err:        err,
			}
		case 1054: // Unknown column
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Column not found",
				Err:        err,
			}
		case 1050: // Table already exists
			return &DatabaseError{
				StatusCode: http.StatusConflict,
				Message:    "Table already exists",
				Err:        err,
			}
		case 1051: // Unknown table
			return &DatabaseError{
				StatusCode: http.StatusNotFound,
				Message:    "Table not found",
				Err:        err,
			}
		case 1048: // Column cannot be null
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Required field cannot be null",
				Err:        err,
			}
		case 1406: // Data too long
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Data too long for column",
				Err:        err,
			}
		case 1264: // Out of range value
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Value out of range",
				Err:        err,
			}
		case 1292: // Incorrect datetime value
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid datetime format",
				Err:        err,
			}
		case 1364: // Field doesn't have a default value
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Field requires a value",
				Err:        err,
			}
		case 1065: // Query was empty
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "Empty query",
				Err:        err,
			}
		case 1064: // Syntax error
			return &DatabaseError{
				StatusCode: http.StatusBadRequest,
				Message:    "SQL syntax error",
				Err:        err,
			}
		case 1142: // Access denied
			return &DatabaseError{
				StatusCode: http.StatusForbidden,
				Message:    "Access denied",
				Err:        err,
			}
		case 2003: // Can't connect to MySQL server
			return &DatabaseError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Database connection failed",
				Err:        err,
			}
		case 2006: // MySQL server has gone away
			return &DatabaseError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Database connection lost",
				Err:        err,
			}
		default:
			return &DatabaseError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Database error: " + mysqlErr.Message,
				Err:        err,
			}
		}
	}

	// Handle other types of errors
	if errors.Is(err, sql.ErrNoRows) {
		return &DatabaseError{
			StatusCode: http.StatusNotFound,
			Message:    "Record not found",
			Err:        err,
		}
	}

	if errors.Is(err, sql.ErrConnDone) {
		return &DatabaseError{
			StatusCode: http.StatusServiceUnavailable,
			Message:    "Database connection closed",
			Err:        err,
		}
	}

	// Default error
	return &DatabaseError{
		StatusCode: http.StatusInternalServerError,
		Message:    "Internal database error",
		Err:        err,
	}
}

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
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	defer tx.Rollback()

	query := "SELECT COUNT(id) FROM project_details where user_id = ?"

	rowCount := 0
	row := tx.QueryRowContext(ctx, query, model.UserId)

	if err = row.Scan(&rowCount); err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if rowCount >= projectLimit {
		return http.StatusConflict, &DatabaseError{
			StatusCode: http.StatusConflict,
			Message:    "Project limit exceeded",
			Err:        fmt.Errorf("user has reached maximum project limit of %d", projectLimit),
		}
	}

	query = "INSERT INTO project_details (user_id, name, description) VALUES (?, ?, ?)"

	res, err := tx.ExecContext(ctx, query, model.UserId, model.Name, model.Desc)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	query = "INSERT INTO users_projects (user_id, project_id, role, status) VALUES (?, ?, ?, ?)"
	_, err = tx.ExecContext(ctx, query, model.UserId, lastInsertId, common.OWNER, common.APPROVED)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	err = tx.Commit()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	return http.StatusCreated, nil
}
func (mysql *MySql) CreateTable(ctx context.Context, projectId int, createTableModel *model.CreateTableModel) (int, error) {
	// Base query
	query := fmt.Sprintf("CREATE TABLE %s", createTableModel.TableName)

	// Start with auto-incrementing ID column
	columnStmt := "`id` INT AUTO_INCREMENT"
	if len(createTableModel.PrimaryKeys) > 0 {
		columnStmt = ""
	}

	// Add additional columns
	for columnName, columnType := range createTableModel.Columns {
		dataTypeMapping, err := MapDataType(columnType, common.MYSQL)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if columnStmt == "" {
			columnStmt = fmt.Sprintf("`%s` %s", columnName, dataTypeMapping)
		} else {
			columnStmt += fmt.Sprintf(", `%s` %s", columnName, dataTypeMapping)
		}
	}

	// Build primary key clause
	var primaryKeyStmt string
	if len(createTableModel.PrimaryKeys) == 0 {
		// No primary key specified, use the default `id`
		primaryKeyStmt = "`id`"
		createTableModel.PrimaryKeys = []string{"id"}
	} else {
		for index, primaryKey := range createTableModel.PrimaryKeys {
			if index > 0 {
				primaryKeyStmt += ", "
			}
			primaryKeyStmt += fmt.Sprintf("`%s`", primaryKey)
		}
	}

	// Process foreign keys
	var foreignKeyStmt string
	// Group foreign keys by referenced table
	tableWiseForeignKey := make(map[string][]*model.ForeignKey)
	for _, fk := range createTableModel.ForeignKeys {
		tableWiseForeignKey[fk.ReferencedTable] = append(tableWiseForeignKey[fk.ReferencedTable], fk)
	}

	for tableName, foreignKeys := range tableWiseForeignKey {
		if len(foreignKeys) > 1 {
			// Multiple foreign keys to same table - create composite key
			if foreignKeyStmt != "" {
				foreignKeyStmt += ", "
			}

			localColumns := ""
			referencedColumns := ""
			for i, fk := range foreignKeys {
				if i > 0 {
					localColumns += ", "
					referencedColumns += ", "
				}
				localColumns += fmt.Sprintf("`%s`", fk.LocalColumn)
				referencedColumns += fmt.Sprintf("`%s`", fk.ReferencedColumn)
			}

			foreignKeyStmt += fmt.Sprintf("FOREIGN KEY (%s) REFERENCES `%s`(%s)",
				localColumns,
				fmt.Sprintf("%d_%s", projectId, tableName),
				referencedColumns)

			if foreignKeys[0].CascadeDelete {
				foreignKeyStmt += " ON DELETE CASCADE"
			}
			if foreignKeys[0].CascadeUpdate {
				foreignKeyStmt += " ON UPDATE CASCADE"
			}
		} else {
			// Single foreign key
			if foreignKeyStmt != "" {
				foreignKeyStmt += ", "
			}
			fk := foreignKeys[0]
			foreignKeyStmt += fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES `%s`(`%s`)",
				fk.LocalColumn,
				fmt.Sprintf("%d_%s", projectId, tableName),
				fk.ReferencedColumn)

			if fk.CascadeDelete {
				foreignKeyStmt += " ON DELETE CASCADE"
			}
			if fk.CascadeUpdate {
				foreignKeyStmt += " ON UPDATE CASCADE"
			}
		}
	}

	// Combine all parts of the query
	query += fmt.Sprintf(" (%s", columnStmt)
	if primaryKeyStmt != "" {
		query += fmt.Sprintf(", PRIMARY KEY (%s)", primaryKeyStmt)
	}
	if foreignKeyStmt != "" {
		query += fmt.Sprintf(", %s", foreignKeyStmt)
	}
	query += ")"

	// Execute the query
	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	return http.StatusCreated, nil
}
func (mysql *MySql) InsertIntoTable(ctx context.Context, model *model.InsertIntoTableModel) (int, error) {
	// Query Generation

	query := fmt.Sprintf("INSERT INTO %s", model.TableName)

	// For NewColumns
	var columnStmt string
	var columnValueStmt string

	i := 0
	for columnName, columnValue := range model.Columns {

		// TODO: not a good way, find other way to handle it
		if columnName == "id" {
			continue
		}
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

	fmt.Println(query)
	_, err := mysql.db.ExecContext(ctx, query)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	return http.StatusCreated, nil
}

func (mysql *MySql) UpdateTableData(ctx context.Context, model *model.UpdateTableDataModel) (int, error) {
	query := fmt.Sprintf("UPDATE %s SET ", model.TableName)

	var values []interface{}
	var setStatements, whereStatements []string

	for columnName, columnValue := range model.NewColumns {
		setStatements = append(setStatements, fmt.Sprintf("%s = ?", columnName))
		values = append(values, columnValue)
	}

	query += strings.Join(setStatements, ", ")
	query += " WHERE "

	for columnName, columnValue := range model.OldColumns {
		whereStatements = append(whereStatements, fmt.Sprintf("%s = ?", columnName))
		values = append(values, columnValue)
	}
	query += strings.Join(whereStatements, " and ")

	result, err := mysql.db.ExecContext(ctx, query, values...)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if rowsAffected == 0 {
		return http.StatusNotFound, &DatabaseError{
			StatusCode: http.StatusNotFound,
			Message:    "No records found to update",
			Err:        fmt.Errorf("no rows matched the update criteria"),
		}
	}

	return http.StatusOK, nil
}

func (mysql *MySql) DropTable(ctx context.Context, model *model.DropTableModel) (int, error) {
	tx, err := mysql.db.BeginTx(ctx, nil)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	defer tx.Rollback()

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", model.TableName)
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	query = fmt.Sprintf("DELETE FROM project_metadata WHERE table_name = ?")
	_, err = tx.ExecContext(ctx, query, model.TableName)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	err = tx.Commit()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	return http.StatusOK, nil
}

func (mysql *MySql) InsertIntoMetaData(ctx context.Context, model *model.CreateTableModel, projectId int) (int, error) {
	query := "INSERT INTO project_metadata (project_id, display_name, table_name, metadata, description) values (?, ?, ?, ?, ?)"

	data, err := json.Marshal(model)
	if err != nil {
		return http.StatusInternalServerError, &DatabaseError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal metadata",
			Err:        err,
		}
	}

	_, err = mysql.db.ExecContext(ctx, query, projectId, model.DisplayName, model.TableName, data, model.Description)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	return http.StatusOK, nil
}

func (mysql *MySql) RetrieveTableListFromMetaData(ctx context.Context, projectId string) ([]*model.CreateTableModel, int, error) {
	query := "SELECT METADATA FROM project_metadata WHERE project_id = ?"
	tables := []*model.CreateTableModel{}

	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, dbErr.StatusCode, dbErr
	}

	defer rows.Close()

	for rows.Next() {
		metaData := ""
		if err := rows.Scan(&metaData); err != nil {
			dbErr := handleDatabaseError(err)
			return nil, dbErr.StatusCode, dbErr
		}

		retrieveTableList := &model.CreateTableModel{}
		err := json.Unmarshal([]byte(metaData), retrieveTableList)
		if err != nil {
			return nil, http.StatusInternalServerError, &DatabaseError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to unmarshal table metadata",
				Err:        err,
			}
		}

		tables = append(tables, retrieveTableList)

	}
	return tables, http.StatusOK, nil
}

func (mysql *MySql) RetrieveDataFromTable(ctx context.Context, retrieveDataFromTableModel *model.RetrieveDataFromTableModel) ([]string, []map[string]interface{}, []string, []*model.ForeignKey, int, error) {
	// Retrieve the table name
	query := "SELECT TABLE_NAME, METADATA FROM project_metadata where project_id = ? and display_name = ?"

	var tableName, metaDataInfo string
	var metaData *model.CreateTableModel

	row := mysql.db.QueryRowContext(ctx, query, retrieveDataFromTableModel.ProjectId, retrieveDataFromTableModel.DisplayName)

	if err := row.Scan(&tableName, &metaDataInfo); err != nil {
		dbErr := handleDatabaseError(err)
		return nil, nil, nil, nil, dbErr.StatusCode, dbErr
	}

	err := json.Unmarshal([]byte(metaDataInfo), &metaData)
	if err != nil {
		return nil, nil, nil, nil, http.StatusInternalServerError, &DatabaseError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to unmarshal table metadata",
			Err:        err,
		}
	}

	var primaryKeys []string
	if len(metaData.PrimaryKeys) == 0 {
		primaryKeys = []string{"id"}
	} else {
		primaryKeys = metaData.PrimaryKeys
	}

	query = "SELECT * FROM " + tableName
	if len(metaData.ForeignKeys) > 0 {
		// Group foreign keys by referenced table
		tableWiseForeignKey := make(map[string][]*model.ForeignKey)
		for _, fk := range metaData.ForeignKeys {
			tableWiseForeignKey[fk.ReferencedTable] = append(tableWiseForeignKey[fk.ReferencedTable], fk)
		}

		joins := []string{}
		for referencedTable, foreignKeys := range tableWiseForeignKey {
			if len(foreignKeys) > 1 {
				// Multiple foreign keys to same table - handle as composite
				onConditions := []string{}
				for _, fk := range foreignKeys {
					onConditions = append(onConditions, fmt.Sprintf("%s.%s = %s.%s",
						tableName,
						fk.LocalColumn,
						fmt.Sprintf("%s_%s", retrieveDataFromTableModel.ProjectId, fk.ReferencedTable),
						fk.ReferencedColumn))
				}
				join := fmt.Sprintf(" LEFT JOIN %s ON %s",
					fmt.Sprintf("%s_%s", retrieveDataFromTableModel.ProjectId, referencedTable),
					strings.Join(onConditions, " AND "))
				joins = append(joins, join)
			} else {
				// Single foreign key
				fk := foreignKeys[0]
				join := fmt.Sprintf(" LEFT JOIN %s ON %s.%s = %s.%s",
					fmt.Sprintf("%s_%s", retrieveDataFromTableModel.ProjectId, fk.ReferencedTable),
					tableName,
					fk.LocalColumn,
					fmt.Sprintf("%s_%s", retrieveDataFromTableModel.ProjectId, fk.ReferencedTable),
					fk.ReferencedColumn)
				joins = append(joins, join)
			}
		}
		query += strings.Join(joins, "")
	}

	fmt.Println(query)

	rows, err := mysql.db.QueryContext(ctx, query)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, nil, nil, nil, dbErr.StatusCode, dbErr
	}

	columns, err := rows.Columns()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, nil, nil, nil, dbErr.StatusCode, dbErr
	}

	results := []map[string]interface{}{}

	defer rows.Close()
	if err = rows.Err(); err != nil {
		dbErr := handleDatabaseError(err)
		return nil, nil, nil, nil, dbErr.StatusCode, dbErr
	}

	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			dbErr := handleDatabaseError(err)
			return nil, nil, nil, nil, dbErr.StatusCode, dbErr
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

	return columns, results, primaryKeys, metaData.ForeignKeys, http.StatusOK, nil
}
func (mysql *MySql) DeleteEntryFromTable(ctx context.Context, model *model.DeleteEntryFromTableModel) (int, error) {
	query := "SELECT TABLE_NAME FROM project_metadata where project_id = ? and display_name = ?"

	tableName := ""
	row := mysql.db.QueryRowContext(ctx, query, model.ProjectId, model.DisplayName)

	if err := row.Scan(&tableName); err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	query = fmt.Sprintf("DELETE FROM %s WHERE ", tableName)

	var whereStatements []string
	var values []interface{}

	for columnName, columnValue := range model.Columns {
		whereStatements = append(whereStatements, fmt.Sprintf("%s = ?", columnName))
		values = append(values, columnValue)
	}
	query += strings.Join(whereStatements, " and ")

	result, err := mysql.db.ExecContext(ctx, query, values...)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if rowsAffected == 0 {
		return http.StatusNotFound, &DatabaseError{
			StatusCode: http.StatusNotFound,
			Message:    "No records found to delete",
			Err:        fmt.Errorf("no rows matched the delete criteria"),
		}
	}

	return http.StatusOK, nil
}

func (mysql *MySql) RegisterUser(ctx context.Context, model *model.UserModel) (int64, int, error) {
	query := "INSERT INTO users (name, email, password) value (?, ?, ?)"

	res, err := mysql.db.ExecContext(ctx, query, model.Name, model.Email, model.Password)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return -1, dbErr.StatusCode, dbErr
	}
	id, err := res.LastInsertId()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return -1, dbErr.StatusCode, dbErr
	}
	return id, http.StatusCreated, nil
}

func (mysql *MySql) LoginUser(ctx context.Context, userModel *model.UserModel) (*model.UserModel, int, error) {
	query := "SELECT id, password, name FROM users where email = ?"

	row := mysql.db.QueryRowContext(ctx, query, userModel.Email)

	hashedPassword := ""
	if err := row.Scan(&userModel.Id, &hashedPassword, &userModel.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userModel, http.StatusNotFound, &DatabaseError{
				StatusCode: http.StatusNotFound,
				Message:    "User not found",
				Err:        fmt.Errorf("user with email %s not registered", userModel.Email),
			}
		}
		dbErr := handleDatabaseError(err)
		return userModel, dbErr.StatusCode, dbErr
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(userModel.Password))
	if err != nil {
		return userModel, http.StatusUnauthorized, &DatabaseError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid password",
			Err:        fmt.Errorf("password does not match"),
		}
	}

	return userModel, http.StatusAccepted, nil

}

func (mysql *MySql) RetrieveProjectList(ctx context.Context, userId int) (map[string][]*model.ProjectModel, int, error) {
	query := "SELECT id, name, description FROM project_details where user_id = ?"

	rows, err := mysql.db.QueryContext(ctx, query, userId)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, dbErr.StatusCode, dbErr
	}

	projectList := []*model.ProjectModel{}

	defer rows.Close()

	var projectIds []string

	for rows.Next() {
		project := &model.ProjectModel{}
		err := rows.Scan(&project.Id, &project.Name, &project.Desc)
		if err != nil {
			dbErr := handleDatabaseError(err)
			return nil, dbErr.StatusCode, dbErr
		}
		projectList = append(projectList, project)
		projectIds = append(projectIds, project.Id)
	}

	if err := rows.Err(); err != nil {
		dbErr := handleDatabaseError(err)
		return nil, dbErr.StatusCode, dbErr
	}

	accessibleProjectList := []*model.ProjectModel{}
	query = "SELECT pd.id, pd.name, pd.description FROM project_details pd JOIN users_projects up ON pd.id = up.project_id WHERE up.user_id = ? AND up.status = ?"
	if len(projectIds) > 0 {
		query += " AND pd.id NOT IN (" + strings.Join(projectIds, ",") + ")"
	}
	rows, err = mysql.db.QueryContext(ctx, query, userId, common.APPROVED)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, dbErr.StatusCode, dbErr
	}
	defer rows.Close()

	for rows.Next() {
		project := &model.ProjectModel{}
		err := rows.Scan(&project.Id, &project.Name, &project.Desc)
		if err != nil {
			dbErr := handleDatabaseError(err)
			return nil, dbErr.StatusCode, dbErr
		}
		accessibleProjectList = append(accessibleProjectList, project)
	}

	return map[string][]*model.ProjectModel{
		"ownerProject":      projectList,
		"accessibleProject": accessibleProjectList,
	}, http.StatusOK, nil

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

	result, err := mysql.db.ExecContext(ctx, query, model.Name, model.Desc, model.Id, model.UserId)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if rowsAffected == 0 {
		return http.StatusNotFound, &DatabaseError{
			StatusCode: http.StatusNotFound,
			Message:    "Project not found or you don't have permission to edit it",
			Err:        fmt.Errorf("no project found with id %s for user %d", model.Id, model.UserId),
		}
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
			return http.StatusNotFound, &DatabaseError{
				StatusCode: http.StatusNotFound,
				Message:    "No access permission found",
				Err:        fmt.Errorf("user %d has no access to project %s", userId, projectId),
			}
		}
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if currentRole < minRole {
		return http.StatusForbidden, &DatabaseError{
			StatusCode: http.StatusForbidden,
			Message:    "Insufficient permissions",
			Err:        fmt.Errorf("user has role %s but requires %s or higher", currentRole.String(), minRole.String()),
		}
	}

	return http.StatusOK, nil
}

func (mysql *MySql) ProjectAccessRequest(ctx context.Context, userId int, projectId string, role common.PERMISSION) (int, error) {
	// Validate inputs
	if userId <= 0 || projectId == "" {
		return http.StatusBadRequest, &DatabaseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user ID or project ID",
			Err:        fmt.Errorf("user ID must be positive and project ID cannot be empty"),
		}
	}
	query := "INSERT INTO users_projects (user_id, project_id, role) VALUES (?, ?, ?)"
	_, err := mysql.db.ExecContext(ctx, query, userId, projectId, role)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}
	return http.StatusCreated, nil
}

func (mysql *MySql) ProjectAccessStatus(ctx context.Context, userId int, projectId string) (common.STATUS, int, error) {
	// Validate inputs
	if userId <= 0 || projectId == "" {
		return common.PENDING, http.StatusBadRequest, &DatabaseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user ID or project ID",
			Err:        fmt.Errorf("user ID must be positive and project ID cannot be empty"),
		}
	}

	query := "SELECT status FROM users_projects WHERE user_id = ? and project_id = ?"
	row := mysql.db.QueryRowContext(ctx, query, userId, projectId)
	var status common.STATUS
	err := row.Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return common.PENDING, http.StatusNotFound, &DatabaseError{
				StatusCode: http.StatusNotFound,
				Message:    "No access request found",
				Err:        fmt.Errorf("no access request found for user %d to project %s", userId, projectId),
			}
		}
		dbErr := handleDatabaseError(err)
		return common.PENDING, dbErr.StatusCode, dbErr
	}
	return status, http.StatusOK, nil
}

func (mysql *MySql) ProjectAccessList(ctx context.Context, userId int, projectId string) ([]*model.ProjectAccessRequestModel, int, error) {
	if userId <= 0 || projectId == "" {
		return nil, http.StatusBadRequest, &DatabaseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user ID or project ID",
			Err:        fmt.Errorf("user ID must be positive and project ID cannot be empty"),
		}
	}

	query := "SELECT user_id, role, status FROM users_projects WHERE project_id = ?"
	rows, err := mysql.db.QueryContext(ctx, query, projectId)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return nil, dbErr.StatusCode, dbErr
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
			dbErr := handleDatabaseError(err)
			return nil, dbErr.StatusCode, dbErr
		}

		projectAccess.Status = status.String()
		projectAccess.Role = role.String()

		if projectAccess.UserId != userId {
			projectAccessList = append(projectAccessList, projectAccess)

		}

	}

	return projectAccessList, http.StatusOK, nil
}

func (mysql *MySql) ProjectAccessAction(ctx context.Context, projectAccessAction *model.ProjectAccessRequestModel, projectId string) (int, error) {
	if projectAccessAction.UserId <= 0 || projectId == "" {
		return http.StatusBadRequest, &DatabaseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid user ID or project ID",
			Err:        fmt.Errorf("user ID must be positive and project ID cannot be empty"),
		}
	}

	var action common.STATUS
	switch projectAccessAction.Action {
	case "approve":
		action = common.APPROVED
	case "reject":
		action = common.REJECTED
	default:
		return http.StatusBadRequest, &DatabaseError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid action",
			Err:        fmt.Errorf("action must be 'approve' or 'reject', got: %s", projectAccessAction.Action),
		}
	}

	query := "UPDATE users_projects SET status = ? WHERE user_id = ? AND project_id = ?"
	result, err := mysql.db.ExecContext(ctx, query, action, projectAccessAction.UserId, projectId)
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbErr := handleDatabaseError(err)
		return dbErr.StatusCode, dbErr
	}

	if rowsAffected == 0 {
		return http.StatusNotFound, &DatabaseError{
			StatusCode: http.StatusNotFound,
			Message:    "Access request not found",
			Err:        fmt.Errorf("no access request found for user %d to project %s", projectAccessAction.UserId, projectId),
		}
	}

	return http.StatusOK, nil
}

func (mysql *MySql) CheckProjectExists(ctx context.Context, projectId string) (bool, int, error) {
	query := "SELECT id FROM project_details WHERE id = ?"
	row := mysql.db.QueryRowContext(ctx, query, projectId)
	var id int
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, http.StatusNotFound, nil
		}
		dbErr := handleDatabaseError(err)
		return false, dbErr.StatusCode, dbErr
	}

	return true, http.StatusOK, nil
}
