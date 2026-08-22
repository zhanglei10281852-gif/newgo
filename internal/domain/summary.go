package domain

type PlatformSummary struct {
	Fields         int `json:"fields"`
	Wells          int `json:"wells"`
	ActiveBatches  int `json:"active_batches"`
	OpenEvents     int `json:"open_events"`
	PendingPermits int `json:"pending_permits"`
	FailedJobs     int `json:"failed_jobs"`
}
type Page struct{ Page, Size, Total int }
