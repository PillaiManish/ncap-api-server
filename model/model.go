package model

type (
	UserModel struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	ProjectModel struct {
		UserId int    `json:"user_id,omitempty"`
		Id     string `json:"id,omitempty"`
		Name   string `json:"name,omitempty"`
		Desc   string `json:"desc,omitempty"`
	}

	ResultStatus struct {
		Status string      `json:"status"`
		Error  string      `json:"error,omitempty"`
		Data   interface{} `json:"data,omitempty"`
	}

	CreateTableModel struct {
		DisplayName string            `json:"name"`
		TableName   string            `json:"table_name"`
		Columns     map[string]string `json:"columns"`
		PrimaryKeys []string          `json:"primary_keys"`
	}

	InsertIntoTableModel struct {
		TableName string            `json:"table_name"`
		Columns   map[string]string `json:"columns"`
	}

	DropTableModel struct {
		TableName string `json:"table_name"`
	}

	RetrieveTableListModel struct {
		ProjectId string `json:"project_id"`
	}

	RetrieveDataFromTableModel struct {
		ProjectId   string `json:"project_id"`
		DisplayName string `json:"display_name"`
	}

	DeleteEntryFromTableModel struct {
		ProjectId   string `json:"project_id"`
		DisplayName string `json:"display_name"`
		TableRowId  int    `json:"table_row_id"`
	}
)
