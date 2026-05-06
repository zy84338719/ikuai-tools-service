package job

import (
	"errors"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/zy84338719/ikuai-tools-service/internal/conf"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
)

var errNoClient = errors.New("ikuai client not initialized")

func itoa(n int) string {
	return strconv.Itoa(n)
}

type Scheduler struct {
	cron *gocron.Scheduler
}

var globalScheduler *Scheduler

func Start(cfg *conf.JobsConfig) error {
	tz, _ := time.LoadLocation("Local")
	s := gocron.NewScheduler(tz)
	s.SetMaxConcurrentJobs(1, gocron.WaitMode)

	for i, c := range cfg.CustomISP {
		tag := "custom-isp-" + strconv.Itoa(i+1)
		s = setCron(s, c.Cron, cfg.SkipStart).Name(tag).Tag(tag)
		if _, err := s.Do(func() {
			if err := SyncCustomISP(&c, cfg); err != nil {
				logger.Error(tag + ": " + err.Error())
			}
		}); err != nil {
			logger.Error("register job " + tag + ": " + err.Error())
		}
		logger.Info("registered job " + tag + " cron=" + c.Cron)
	}

	for i, c := range cfg.StreamDomain {
		tag := "stream-domain-" + strconv.Itoa(i+1)
		s = setCron(s, c.Cron, cfg.SkipStart).Name(tag).Tag(tag)
		if _, err := s.Do(func() {
			if err := SyncStreamDomain(&c, cfg); err != nil {
				logger.Error(tag + ": " + err.Error())
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
