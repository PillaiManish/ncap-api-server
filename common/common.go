package common

const (
	JSONApplicationType = "application/json"
	FAILED              = "FAILED"
	SUCCESS             = "SUCCESS"
	ADMIN_PANEL_URL     = "ADMIN_PANEL_URL"
	MYSQL               = "MySQL"
)

type PERMISSION int

const (
	ADMIN PERMISSION = 8
	OWNER            = 10
)

type STATUS int

const (
	PENDING STATUS = iota
	APPROVED
	REJECTED
)

func (p PERMISSION) String() string {
	switch p {
	case OWNER:
		return "OWNER"
	case ADMIN:
		return "ADMIN"
	default:
		return "UNKNOWN"
	}
}

func (s STATUS) String() string {
	switch s {
	case PENDING:
		return "PENDING"
	case APPROVED:
		return "APPROVED"
	case REJECTED:
		return "REJECTED"
	default:
		return "UNKNOWN"
	}
}
