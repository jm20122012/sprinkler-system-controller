package controllerservice

import (
	"context"
	"log/slog"
	"sprinkler-controller-service/internal/config"
	"sync"
	"time"
)

type ControllerService struct {
	Ctx           context.Context
	Wg            *sync.WaitGroup
	Logger        *slog.Logger
	Config        *config.Config
	LastResetDate time.Time
	Mutex         sync.Mutex
}

func NewControllerService(ctx context.Context, wg *sync.WaitGroup, logger *slog.Logger, cfg *config.Config) *ControllerService {
	return &ControllerService{
		Ctx:           ctx,
		Wg:            wg,
		Logger:        logger,
		Config:        cfg,
		LastResetDate: time.Now(),
		Mutex:         sync.Mutex{},
	}
}

func (c *ControllerService) ResetIfNewDay() {
	currentDate := time.Now().Truncate(24 * time.Hour)
	if currentDate.After(c.LastResetDate) {
		for _, zoneItem := range c.Config.ZoneList {
			for idx := range zoneItem.Schedule {
				c.Mutex.Lock()
				zoneItem.Schedule[idx].Completed = false
				c.Mutex.Unlock()
			}
		}
		c.LastResetDate = currentDate
	}
}

func (c *ControllerService) Run() {
	for {
		select {
		case <-c.Ctx.Done():
			c.Logger.Info("Done context signal detected in controller service - cleaning up")
			c.Wg.Done()
		default:

			c.ResetIfNewDay()

			c.Logger.Info("Checking zone schedule...")
			for zoneName, zoneInfo := range c.Config.ZoneList {
				c.Logger.Debug("Checking zone schedule", "zone", zoneName)
				for idx := range zoneInfo.Schedule {
					scheduleItem := zoneInfo.Schedule[idx]
					currentTime := time.Now()
					c.Logger.Debug("Comparing current and start times", "zone", zoneName, "current", currentTime, "startTime", scheduleItem.StartTime)

					t, err := time.Parse(time.TimeOnly, scheduleItem.StartTime)
					if err != nil {
						c.Logger.Error("Error parsing start time", "startTime", scheduleItem.StartTime, "error", err)
					}

					if time.Now().After(t) {
						c.Logger.Info("Current time exceeds start time for zone schedule item", "zone", zoneName, "currentTime", time.Now(), "startTime", t)
					}
				}
			}
		}

		time.Sleep(1 * time.Second)
	}
}
