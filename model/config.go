package model

type (
	Config struct {
		Database *DatabaseConfig `yaml:"database,omitempty" json:"database,omitempty"`
	}

	DatabaseConfig struct {
		Type       string            `yaml:"type" json:"type"`
		Connection *ConnectionConfig `yaml:"connection,omitempty" json:"connection,omitempty"`
	}

	ConnectionConfig struct {
		URI string `yaml:"uri" json:"uri"`
	}
)
