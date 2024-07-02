package api

import (
	"context"
	"log/slog"
	"sprinkler-controller-service/internal/config"
)

type DryRunApiHandler struct {
	Ctx    context.Context
	Logger *slog.Logger
	ApiUrl string
}

func NewDryRunApiHandler(ctx context.Context, logger *slog.Logger, apiUrl string) IApiHandler {
	return &DryRunApiHandler{
		Ctx:    ctx,
		Logger: logger,
		ApiUrl: apiUrl,
	}
}

func (a *DryRunApiHandler) SendSprinklerEventRequest(event *config.ScheduleItem) error {
	a.Logger.Info("Sending event to API", "event", event)

	return nil
}
