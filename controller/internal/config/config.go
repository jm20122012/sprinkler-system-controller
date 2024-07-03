package config

import (
	"encoding/json"
	"os"
	"sync"
)

type Config struct {
	AppConfig AppConfig        `json:"appConfig"`
	ZoneList  map[string]*Zone `json:"zoneList"`
}

type AppConfig struct {
	TZ         string `json:"tz"`
	ApiUrl     string `json:"apiUrl"`
	DebugLevel string `json:"debugLevel"`
	DryRun     bool   `json:"dryRun"`
}

type Zone struct {
	FriendlyName string         `json:"friendlyName"`
	Location     string         `json:"location"`
	Schedule     []ScheduleItem `json:"schedule"`
}

// Weekdays is a uint8 where each bit represents a day of the week
// starting with Sunday at the LSB.  The MSB is not used.  For example,
// 0 0 1 0 1 0 1 0 means that Monday, Wednesday, and Friday are enabled
type ScheduleItem struct {
	StartTime       string        `json:"startTime"`
	DurationMinutes int           `json:"durationMinutes"`
	Weekdays        uint8         `json:"weekdays"`
	Active          bool          `json:"active"`
	Mutex           *sync.RWMutex `json:"-"`
}

func LoadConfig(file string) (*Config, error) {
	var cfg Config

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	for _, item := range cfg.ZoneList {
		for idx := range item.Schedule {
			item.Schedule[idx].Mutex = &sync.RWMutex{}
		}
	}
	return &cfg, nil
}
