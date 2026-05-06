package job

import (
	"sync"
	"time"
)

type JobStatus struct {
	Name      string     `json:"name"`
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

func markRunning(name string) {
	statusMu.Lock()
	defer statusMu.Unlock()
	if _, ok := statuses[name]; !ok {
		statuses[name] = &JobStatus{Name: name}
	}
	statuses[name].Running = true
}

func markDone(name string, err error) {
	statusMu.Lock()
	defer statusMu.Unlock()
	s := statuses[name]
	if s == nil {
		return
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
