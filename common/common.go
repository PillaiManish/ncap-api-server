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
	OWNER PERMISSION = iota
	ADMIN
)

type STATUS int

const (
	PENDING STATUS = iota
	APPROVED
	REJECTED
)
