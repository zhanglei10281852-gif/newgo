package domain

type Action string

const (
	ActionReadField     Action = "read_field"
	ActionManageField   Action = "manage_field"
	ActionRecordEvent   Action = "record_event"
	ActionRequestPermit Action = "request_permit"
	ActionReviewPermit  Action = "review_permit"
	ActionReadAudit     Action = "read_audit"
)

func Allowed(role Role, action Action) bool {
	switch action {
	case ActionReadField:
		return role != ""
	case ActionManageField, ActionRecordEvent:
		return role == RoleFieldEngineer || role == RoleGeophysicist || role == RoleAuditor
	case ActionRequestPermit:
		return role == RoleFieldEngineer || role == RoleGeophysicist
	case ActionReviewPermit:
		return role == RoleSafetyReviewer || role == RoleAuditor
	case ActionReadAudit:
		return role == RoleAuditor
	default:
		return false
	}
}
