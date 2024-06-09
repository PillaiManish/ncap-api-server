package model

type (
	ProjectModel struct {
		Id   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Desc string `json:"desc,omitempty"`
	}

	ResultStatus struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}

	CreateTableModel struct {
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
)
