package controllerservice

import (
	"context"
	"encoding/json"
	"log/slog"
	"sprinkler-controller-service/internal/config"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CommandMessage struct {
	CommandType string `json:"command_type"`
	Zone        string `json:"zone"`
	State       int    `json:"state"`
}

// {
// 	"messageType":"status",
// 	"zone1Active":false,
// 	"zone1LastOnTime":"2024-07-05 16:40:01 UTC",
// 	"zone2Active":false,
// 	"zone2LastOnTime":"2024-07-05 17:00:01 UTC",
// 	"zone3Active":true,
// 	"zone3LastOnTime":"2024-07-05 17:20:03 UTC",
// 	"zone4Active":false,
// 	"zone4LastOnTime":"never"
// }

type PicowStatusResponse struct {
	MessageType     string `json:"messageType"`
	Zone1Active     bool   `json:"zone1Active"`
	Zone1LastOnTime string `json:"zone1LastOnTime"`
	Zone2Active     bool   `json:"zone2Active"`
	Zone2LastOnTime string `json:"zone2LastOnTime"`
	Zone3Active     bool   `json:"zone3Active"`
	Zone3LastOnTime string `json:"zone3LastOnTime"`
	Zone4Active     bool   `json:"zone4Active"`
	Zone4LastOnTime string `json:"zone4LastOnTime"`
}

type ControllerService struct {
	Ctx           context.Context
	Wg            *sync.WaitGroup
	Logger        *slog.Logger
	Config        *config.Config
	ApiHandler    IApiHandler
	LastResetDate time.Time
	TaskQueue     chan CommandMessage
	MqttClient    *MqttClient
}

func NewControllerService(ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	cfg *config.Config,
	apiHdnlr IApiHandler,
) *ControllerService {
	cs := &ControllerService{
		Ctx:           ctx,
		Wg:            wg,
		Logger:        logger,
		Config:        cfg,
		ApiHandler:    apiHdnlr,
		LastResetDate: time.Now(),
		TaskQueue:     make(chan CommandMessage, 100),
	}

	client := NewMqttClient(
		cfg.AppConfig.MqttBroker,
		cfg.AppConfig.MqttPort,
		cs.OnMsgHndlrFactory(),
		cs.OnConnectHndlrFactory(),
		cs.OnConnectLostHndlrFactory(),
	)

	cs.MqttClient = client

	cs.MqttSubscribe("sprinkler_system_controller/picow/status")

	return cs
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
			c.MqttCleanup()
			c.Wg.Done()
			return
		default:
			for zoneName, zoneInfo := range c.Config.ZoneList {
				c.Logger.Debug("Checking zone schedule", "zone", zoneName)
				zoneInfo.Mutex.RLock()

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

					c.Logger.Debug("Time comparison",
						"zone", zoneName,
						"scheduleItemIndex", idx,
						"current", currentTime.Format(time.RFC3339),
						"start", startTime.Format(time.RFC3339),
						"end", endTime.Format(time.RFC3339),
						"zoneActive", zoneInfo.Active,
					)

					if currentTime.After(startTime) && currentTime.Before(endTime) && !zoneInfo.Active {
						c.Logger.Info("Starting sprinkler event",
							"zoneName", zoneName,
							"currentTime", currentTime.Format(time.RFC3339),
							"startTime", startTime.Format(time.RFC3339),
							"endTime", endTime.Format(time.RFC3339),
							"durationMinutes", scheduleItem.DurationMinutes,
						)

						newCmdMessage := CommandMessage{
							CommandType: "update_zone_state",
							Zone:        zoneName,
							State:       1,
						}
						c.TaskQueue <- newCmdMessage
					} else if currentTime.After(endTime) && zoneInfo.Active {
						c.Logger.Info("Stopping sprinkler event",
							"zoneName", zoneName,
							"currentTime", currentTime.Format(time.RFC3339),
							"startTime", startTime.Format(time.RFC3339),
							"endTime", endTime.Format(time.RFC3339),
							"durationMinutes", scheduleItem.DurationMinutes,
						)

						newCmdMessage := CommandMessage{
							CommandType: "update_zone_state",
							Zone:        zoneName,
							State:       0,
						}
						c.TaskQueue <- newCmdMessage
					}
				}
				zoneInfo.Mutex.RUnlock()
			}
		}

		timer := time.NewTimer(5 * time.Second)
		select {
		case <-c.Ctx.Done():
			c.Logger.Debug("Done ctx signal detected during contoller service wait loop - exiting")
			taskProcCancel()
			taskProcWg.Wait()
			c.MqttCleanup()
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
		case m := <-c.TaskQueue:
			jsonEncoded, err := json.Marshal(m)
			if err != nil {
				c.Logger.Error("Error marshalling command message", "error", err, "msg", m)
				continue
			}
			c.Logger.Info("Sending command message", "msg", m)
			token := c.MqttClient.Client.Publish("sprinkler_system_controller/picow/command", 1, false, jsonEncoded)

			// Wait for the publish to complete
			if token.Wait() && token.Error() != nil {
				c.Logger.Error("Failed to publish message", "error", token.Error(), "msg", m)
			} else {
				c.Logger.Info("Message published successfully", "msg", m)
			}
		}
	}
}

func (c *ControllerService) OnMsgHndlrFactory() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		c.Logger.Info("Received MQT message: ", "payload", msg.Payload(), "topic", msg.Topic())

		switch msg.Topic() {
		case "sprinkler_system_controller/picow/status":
			var statusMsg PicowStatusResponse
			err := json.Unmarshal(msg.Payload(), &statusMsg)
			if err != nil {
				c.Logger.Error("Error unmarshalling picow status message", "msg", msg.Payload(), "error", err)
			}
			// This is messy - need a better way to do this
			c.Config.ZoneList["zone1"].Mutex.Lock()
			c.Config.ZoneList["zone1"].Active = statusMsg.Zone1Active
			c.Config.ZoneList["zone1"].Mutex.Unlock()

			c.Config.ZoneList["zone2"].Mutex.Lock()
			c.Config.ZoneList["zone2"].Active = statusMsg.Zone2Active
			c.Config.ZoneList["zone2"].Mutex.Unlock()

			c.Config.ZoneList["zone3"].Mutex.Lock()
			c.Config.ZoneList["zone3"].Active = statusMsg.Zone3Active
			c.Config.ZoneList["zone3"].Mutex.Unlock()

			c.Config.ZoneList["zone4"].Mutex.Lock()
			c.Config.ZoneList["zone4"].Active = statusMsg.Zone4Active
			c.Config.ZoneList["zone4"].Mutex.Unlock()
		}
	}
}

func (c *ControllerService) OnConnectHndlrFactory() mqtt.OnConnectHandler {
	return func(client mqtt.Client) {
		c.Logger.Info("MQTT connected")
	}
}

func (c *ControllerService) OnConnectLostHndlrFactory() mqtt.ConnectionLostHandler {
	return func(client mqtt.Client, err error) {
		c.Logger.Error("MQTT connection lost", "error", err)
	}
}

func (c *ControllerService) MqttSubscribe(topic string) {
	token := c.MqttClient.Client.Subscribe(topic, 1, nil)
	token.Wait()
	c.Logger.Info("Subscribing to MQTT topic", "topic", topic)
}

func (c *ControllerService) MqttCleanup() {
	c.Logger.Info("MQTT cleanup called")

	c.MqttClient.Client.Unsubscribe("sprinkler_system_controller/picow/status")

	c.MqttClient.Client.Disconnect(100)

}
