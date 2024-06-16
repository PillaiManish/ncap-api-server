package model

type (
	Config struct {
		Database *DatabaseConfig `yaml:"database,omitempty" json:"database,omitempty"`
		JWT      *JWT            `yaml:"jwt,omitempty" json:"jwt,omitempty"`
	}

	DatabaseConfig struct {
		Type       string            `yaml:"type" json:"type"`
		Connection *ConnectionConfig `yaml:"connection,omitempty" json:"connection,omitempty"`
	}

	ConnectionConfig struct {
		URI string `yaml:"uri" json:"uri"`
	}

	JWT struct {
		Secret []byte `yaml:"secret" json:"secret"`
	}
)
