package database

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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
