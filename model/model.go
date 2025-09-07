package model

type (
	UserModel struct {
		Id       int    `json:"id,omitempty"`
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
		Status     string `json:"status"`
		StatusCode int    `json:"statusCode"`
		//Error      string      `json:"error,omitempty"`
		Data    interface{} `json:"data,omitempty"`
		Message string      `json:"message,omitempty"`
	}

	CreateTableModel struct {
		Description string            `json:"description,omitempty"`
		DisplayName string            `json:"name"`
		TableName   string            `json:"table_name"`
		Columns     map[string]string `json:"columns"`
		PrimaryKeys []string          `json:"primary_keys"`
	}

	InsertIntoTableModel struct {
		TableName string            `json:"table_name"`
		Columns   map[string]string `json:"columns"`
	}

	UpdateTableDataModel struct {
		TableName  string            `json:"table_name"`
		NewColumns map[string]string `json:"newColumns"`
		OldColumns map[string]string `json:"oldColumns"`
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
		ProjectId   string            `json:"project_id"`
		DisplayName string            `json:"display_name"`
		Columns     map[string]string `json:"columns"`
	}

	ProjectAccessRequestModel struct {
		UserId int    `json:"user_id"` //TODO: replace it with email or name
		Role   string `json:"role,omitempty"`
		Status string `json:"status,omitempty"`
		Action string `json:"action,omitempty"`
	}
)
