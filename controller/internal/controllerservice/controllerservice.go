package controllerservice

import (
	"context"
	"log/slog"
	"sprinkler-controller-service/internal/config"
	"sync"
	"time"
)

type Task struct {
	ZoneName string
	Action   bool
}

type ControllerService struct {
	Ctx           context.Context
	Wg            *sync.WaitGroup
	Logger        *slog.Logger
	Config        *config.Config
	ApiHandler    IApiHandler
	LastResetDate time.Time
	TaskQueue     chan *config.ScheduleItem
}

func NewControllerService(ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	cfg *config.Config,
	apiHdnlr IApiHandler,
) *ControllerService {
	return &ControllerService{
		Ctx:           ctx,
		Wg:            wg,
		Logger:        logger,
		Config:        cfg,
		ApiHandler:    apiHdnlr,
		LastResetDate: time.Now(),
		TaskQueue:     make(chan *config.ScheduleItem, 100),
	}
}

func isDayEnabled(weekdays uint8, day time.Weekday) bool {
	// Convert time.Weekday to our custom representation
	// time.Weekday is 0-6 where 0 is Sunday, 1 is Monday, etc.
	// We need to shift 1 by this amount, but handle Sunday specially
	var dayBit uint8
	if day == time.Sunday {
		dayBit = 1 // Sunday is already the LSB
	} else {
		dayBit = 1 << day
	}

	return weekdays&dayBit != 0
}

func (c *ControllerService) Run() {
	taskProcCtx, taskProcCancel := context.WithCancel(context.Background())
	taskProcWg := sync.WaitGroup{}
	taskProcWg.Add(1)

	go c.TaskProcessor(&taskProcWg, taskProcCtx)

	for {
		select {
		case <-c.Ctx.Done():
			c.Logger.Info("Done context signal detected in controller service - cleaning up")
			taskProcCancel()
			taskProcWg.Wait()
			c.Wg.Done()
			return
		default:

			for zoneName, zoneInfo := range c.Config.ZoneList {
				c.Logger.Debug("Checking zone schedule", "zone", zoneName)
				for idx := range zoneInfo.Schedule {
					scheduleItem := &zoneInfo.Schedule[idx]

					// Parse the start time
					currentTime := time.Now()

					// Check if today is an enabled day
					if !isDayEnabled(scheduleItem.Weekdays, currentTime.Weekday()) {
						c.Logger.Debug("Skipping schedule, day not enabled",
							"zone", zoneName,
							"weekdays", scheduleItem.Weekdays,
							"today", currentTime.Weekday())
						continue
					}

					startTime, err := time.ParseInLocation("15:04:05", scheduleItem.StartTime, currentTime.Location())
					if err != nil {
						c.Logger.Error("Error parsing start time", "startTime", scheduleItem.StartTime, "error", err)
						continue
					}

					// Adjust startTime to today's date
					startTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
						startTime.Hour(), startTime.Minute(), startTime.Second(),
						0, currentTime.Location())

					// Calculate end time
					duration := time.Duration(scheduleItem.DurationMinutes) * time.Minute
					endTime := startTime.Add(duration)

					scheduleItem.Mutex.RLock()
					c.Logger.Debug("Time comparison",
						"zone", zoneName,
						"scheduleItemIndex", idx,
						"current", currentTime.Format(time.RFC3339),
						"start", startTime.Format(time.RFC3339),
						"end", endTime.Format(time.RFC3339),
						"zoneActive", scheduleItem.Active,
					)

					if currentTime.After(startTime) && currentTime.Before(endTime) && !scheduleItem.Active {
						c.Logger.Info("Starting sprinkler event",
							"zoneName", zoneName,
							"currentTime", currentTime.Format(time.RFC3339),
							"startTime", startTime.Format(time.RFC3339),
							"endTime", endTime.Format(time.RFC3339),
							"durationMinutes", scheduleItem.DurationMinutes)
						c.TaskQueue <- scheduleItem
					} else if currentTime.After(endTime) && scheduleItem.Active {
						c.Logger.Info("Stopping sprinkler event",
							"zoneName", zoneName,
							"currentTime", currentTime.Format(time.RFC3339),
							"startTime", startTime.Format(time.RFC3339),
							"endTime", endTime.Format(time.RFC3339),
							"durationMinutes", scheduleItem.DurationMinutes)
						c.TaskQueue <- scheduleItem
					}
					scheduleItem.Mutex.RUnlock()
				}
			}
		}

		timer := time.NewTimer(5 * time.Second)
		select {
		case <-c.Ctx.Done():
			c.Logger.Debug("Done ctx signal detected during contoller service wait loop - exiting")
			taskProcCancel()
			taskProcWg.Wait()
			c.Wg.Done()
			return
		case <-timer.C:
		}
	}
}

func (c *ControllerService) TaskProcessor(wg *sync.WaitGroup, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.Logger.Info("Done signal detected in TaskProcessor - exiting")
			wg.Done()
			return
		case t := <-c.TaskQueue:
			// Send POST request to API to start the event
			err := c.ApiHandler.SendSprinklerEventRequest(t)
			if err != nil {
				c.Logger.Error("API request error", "event", t)
				continue
			}
			c.Logger.Info("Setting zone active flag to true", "schduleItem", t)
			t.Mutex.Lock()
			t.Active = true
			t.Mutex.Unlock()
		}
	}
}

// for zoneName, zoneInfo := range c.Config.ZoneList {
// 	c.Logger.Debug("Checking zone schedule", "zone", zoneName)
// 	for idx := range zoneInfo.Schedule {
// 		currentTime := time.Now()

// 		scheduleItem := &zoneInfo.Schedule[idx]

// 		c.Logger.Debug("Comparing current and start times", "zone", zoneName, "current", currentTime, "startTime", scheduleItem.StartTime)

// 		startTime, err := time.Parse(time.TimeOnly, scheduleItem.StartTime)
// 		if err != nil {
// 			c.Logger.Error("Error parsing start time", "startTime", scheduleItem.StartTime, "error", err)
// 		}

// 		duration := time.Duration(scheduleItem.DurationMinutes)
// 		endTime := startTime.Add(duration * time.Minute)

// 		scheduleItem.Mutex.RLock()
// 		if currentTime.After(startTime) && !scheduleItem.Active {
// 			c.Logger.Debug("Zone is not active and current time exceeds start time for zone schedule item", "zone", zoneName, "currentTime", time.Now(), "startTime", startTime)
// 			c.Logger.Info("Starting sprinkler event", "zoneName", zoneName, "currentTime", currentTime, "startTime", scheduleItem.StartTime, "endTime", endTime, "durationMinutes", scheduleItem.DurationMinutes)
// 			c.TaskQueue <- scheduleItem
// 		}

// 		if currentTime.After(endTime) && scheduleItem.Active {
// 			c.Logger.Debug("Zone is active and current time exceeds end time for zone schedule item", "zone", zoneName, "currentTime", time.Now(), "startTime", startTime)
// 			c.Logger.Info("Stopping sprinkler event", "zoneName", zoneName, "currentTime", currentTime, "startTime", scheduleItem.StartTime, "endTime", endTime, "durationMinutes", scheduleItem.DurationMinutes)
// 			c.TaskQueue <- scheduleItem
// 		}
// 		scheduleItem.Mutex.RUnlock()
