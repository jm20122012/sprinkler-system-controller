package controllerservice

import (
	"context"
	"log/slog"
	"sprinkler-controller-service/internal/config"
)

type IApiHandler interface {
	SendSprinklerEventRequest(event *config.ScheduleItem) error
}

type ApiHandler struct {
	Ctx    context.Context
	Logger *slog.Logger
	ApiUrl string
}

func NewApiHandler(ctx context.Context, logger *slog.Logger, apiUrl string) IApiHandler {
	return &ApiHandler{
		Ctx:    ctx,
		Logger: logger,
		ApiUrl: apiUrl,
	}
}

func (a *ApiHandler) SendSprinklerEventRequest(event *config.ScheduleItem) error {
	a.Logger.Info("Sending event to API", "event", event)

	return nil
}
