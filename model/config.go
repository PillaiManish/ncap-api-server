package model

type (
	Config struct {
		LogLevel      string          `yaml:"logLevel,omitempty" json:"logLevel,omitempty"`
		AdminPanelUrl string          `yaml:"adminPanelUrl,omitempty" json:"adminPanelUrl,omitempty"`
		Database      *DatabaseConfig `yaml:"database,omitempty" json:"database,omitempty"`
		ProjectConfig *ProjectConfig  `yaml:"projectConfig,omitempty" json:"projectConfig,omitempty"`
		JWT           *JWT            `yaml:"jwt,omitempty" json:"jwt,omitempty"`
	}

	ProjectConfig struct {
		Limit int `yaml:"limit,omitempty" json:"limit,omitempty"`
	}

	DatabaseConfig struct {
		Type       string            `yaml:"type" json:"type"`
		Connection *ConnectionConfig `yaml:"connection,omitempty" json:"connection,omitempty"`
		SSL        *SSL              `yaml:"ssl,omitempty" json:"ssl,omitempty"`
	}

	SSL struct {
		Enabled bool   `json:"enabled" yaml:"enabled"`
		CAPath  string `json:"caPath" yaml:"caPath"`
	}

	ConnectionConfig struct {
		URI    *string `yaml:"uri" json:"uri"`
		DBName string  `yaml:"dbName" json:"dbName"`
		Host   string  `yaml:"host" json:"host"`
		Port   int     `yaml:"port" json:"port"`
		User   string  `yaml:"user" json:"user"`
		Pass   string  `yaml:"pass" json:"pass"`
	}

	JWT struct {
		Secret []byte `yaml:"secret" json:"secret"`
	}
)
