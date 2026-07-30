package job

import (
	"context"
	"sync"
	"time"

	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db"
	"github.com/zy84338719/ikuai-tools-service/internal/repo/db/model"
	"go.uber.org/zap"
)

// JobStatus is the in-memory, real-time view of a job (used for "running now"
// indicators). Historical runs live in the job_runs table.
type JobStatus struct {
	Name      string     `json:"name"`
	RouterID  string     `json:"router_id"`
	LastRunAt *time.Time `json:"last_run_at"`
	LastError string     `json:"last_error"`
	Success   bool       `json:"success"`
	Running   bool       `json:"running"`
	RunCount  int        `json:"run_count"`
}

var (
	statusMu sync.RWMutex
	statuses = map[string]*JobStatus{}
)

func markRunning(routerID, name string) {
	statusMu.Lock()
	defer statusMu.Unlock()
	key := routerID + "/" + name
	if _, ok := statuses[key]; !ok {
		statuses[key] = &JobStatus{Name: name, RouterID: routerID}
	}
	statuses[key].Running = true
}

func markDone(routerID, name string, err error) {
	statusMu.Lock()
	defer statusMu.Unlock()
	key := routerID + "/" + name
	s := statuses[key]
	if s == nil {
		s = &JobStatus{Name: name, RouterID: routerID}
		statuses[key] = s
	}
	now := time.Now()
	s.LastRunAt = &now
	s.Running = false
	s.RunCount++
	if err != nil {
		s.LastError = err.Error()
		s.Success = false
	} else {
		s.LastError = ""
		s.Success = true
	}
}

// recordRun persists a finished job execution to the job_runs table. Failures
// are logged but never propagate (history is best-effort).
func recordRun(routerID, jobType, tag, status string, changed, failed int, dur time.Duration, errMsg string) {
	if db.DB == nil {
		return
	}
	now := time.Now()
	r := &model.JobRun{
		RouterID:  routerID,
		JobType:   jobType,
		Tag:       tag,
		Status:    status,
		Changed:   changed,
		Failed:    failed,
		Duration:  float64(dur.Milliseconds()),
		Error:     errMsg,
		StartedAt: now.Add(-dur),
		EndedAt:   now,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.DB.WithContext(ctx).Create(r).Error; err != nil {
		logger.Error("job_run write failed", zap.Error(err))
	}
}

// ListHistory returns recent job executions from the DB.
func ListHistory(limit int) ([]model.JobRun, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var runs []model.JobRun
	if db.DB == nil {
		return runs, nil
	}
	err := db.DB.Order("id DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

func GetAllStatuses() []JobStatus {
	statusMu.RLock()
	defer statusMu.RUnlock()
	out := make([]JobStatus, 0, len(statuses))
	for _, s := range statuses {
		cp := *s
		out = append(out, cp)
	}
	return out
}
