package domain

import "time"

type JobState string

const (
	JobPending JobState = "pending"
	JobRunning JobState = "running"
	JobFailed  JobState = "failed"
	JobDone    JobState = "done"
	JobDead    JobState = "dead"
)

type DeliveryJob struct {
	ID, Kind, EntityID, Payload string
	State                       JobState
	Attempts, MaxAttempts       int
	AvailableAt, LockedAt       *time.Time
	LastError                   string
	CreatedAt, UpdatedAt        time.Time
}

func (j DeliveryJob) Claim(now time.Time) (DeliveryJob, error) {
	if j.State != JobPending && j.State != JobFailed {
		return j, ConflictError{"job", "not claimable"}
	}
	j.State = JobRunning
	j.Attempts++
	t := now.UTC()
	j.LockedAt = &t
	j.UpdatedAt = t
	return j, nil
}
func (j DeliveryJob) Retry(err error, now time.Time) (DeliveryJob, error) {
	if j.State != JobRunning {
		return j, ConflictError{"job", "must be running"}
	}
	j.LastError = err.Error()
	t := now.UTC().Add(time.Duration(j.Attempts) * time.Second)
	j.AvailableAt = &t
	j.UpdatedAt = now.UTC()
	if j.Attempts >= j.MaxAttempts {
		j.State = JobDead
	} else {
		j.State = JobFailed
	}
	return j, nil
}
func (j DeliveryJob) Complete(now time.Time) (DeliveryJob, error) {
	if j.State != JobRunning {
		return j, ConflictError{"job", "must be running"}
	}
	j.State = JobDone
	j.UpdatedAt = now.UTC()
	return j, nil
}
