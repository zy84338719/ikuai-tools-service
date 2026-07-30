package model

import "time"

// JobRun records one execution of a scheduled job (custom_isp / stream_domain
// sync). It replaces the in-memory status map so history survives restarts.
type JobRun struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RouterID  string    `gorm:"size:50;index" json:"router_id"`
	JobType   string    `gorm:"size:32;index" json:"job_type"`  // "custom-isp" / "stream-domain"
	Tag       string    `gorm:"size:64;index" json:"tag"`        // rule tag/name within the job
	Status    string    `gorm:"size:16;index" json:"status"`     // "success" / "failed" / "running"
	Changed   int       `json:"changed"`                          // rows/chunks added+edited
	Failed    int       `json:"failed"`                           // chunks that errored
	Duration  float64   `json:"duration_ms"`                      // wall-clock duration in ms
	Error     string    `gorm:"type:text" json:"error"`
	StartedAt time.Time `gorm:"index" json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

func (JobRun) TableName() string {
	return "job_runs"
}
