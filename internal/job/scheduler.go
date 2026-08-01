package job

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/ikuai"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

var errNoClient = errors.New("ikuai client not initialized")

// errV4ActionCallGone signals that a legacy /Action/call feature is no longer
// reachable on iKuai v4 (the RPC was removed). Jobs that depend on it abort
// with this so job_runs records a clear cause instead of an opaque 404.
var errV4ActionCallGone = errors.New("not supported on iKuai v4: the v3 /Action/call RPC was removed and this feature has no v4 equivalent yet")

func itoa(n int) string { return strconv.Itoa(n) }

type Scheduler struct {
	cron *gocron.Scheduler
}

var globalScheduler *Scheduler

// Start registers the cron jobs. Each job runs once per registered router, so
// adding a router live (without restart) makes the next tick cover it too.
func Start(cfg *conf.JobsConfig) error {
	tz, _ := time.LoadLocation("Local")
	s := gocron.NewScheduler(tz)
	s.SetMaxConcurrentJobs(1, gocron.WaitMode)

	for i, c := range cfg.CustomISP {
		tag := "custom-isp-" + strconv.Itoa(i+1)
		c := c // capture
		s = setCron(s, c.Cron, cfg.SkipStart).Name(tag).Tag(tag)
		if _, err := s.Do(func() {
			for _, routerID := range ikuai.GetRegistry().Names() {
				if err := SyncCustomISP(routerID, &c, cfg); err != nil {
					logger.Error(tag + "[" + routerID + "]: " + err.Error())
				}
			}
		}); err != nil {
			logger.Error("register job " + tag + ": " + err.Error())
		}
		logger.Info("registered job " + tag + " cron=" + c.Cron)
	}

	for i, c := range cfg.StreamDomain {
		tag := "stream-domain-" + strconv.Itoa(i+1)
		c := c // capture
		s = setCron(s, c.Cron, cfg.SkipStart).Name(tag).Tag(tag)
		if _, err := s.Do(func() {
			for _, routerID := range ikuai.GetRegistry().Names() {
				if err := SyncStreamDomain(routerID, &c, cfg); err != nil {
					logger.Error(tag + "[" + routerID + "]: " + err.Error())
				}
			}
		}); err != nil {
			logger.Error("register job " + tag + ": " + err.Error())
		}
		logger.Info("registered job " + tag + " cron=" + c.Cron)
	}

	globalScheduler = &Scheduler{cron: s}
	s.StartAsync()
	logger.Info("job scheduler started, total jobs=" + itoa(s.Len()))
	return nil
}

func Stop() {
	if globalScheduler != nil {
		globalScheduler.cron.Stop()
	}
}

func setCron(s *gocron.Scheduler, cronStr string, skipStart bool) *gocron.Scheduler {
	d, err := time.ParseDuration(cronStr)
	if err != nil {
		return s.Cron(cronStr)
	}
	s = s.Every(d)
	if skipStart {
		s = s.StartAt(time.Now().Add(d))
	}
	return s
}
