package database

type MySql struct {
}

func NewMySQLConnection() Database {
	return &MySql{}
}

func (mysql *MySql) CreateNewProject(projectName string) error {
	return nil
}
