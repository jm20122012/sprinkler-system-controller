package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type StatusResponse struct {
	MicrocontrollerStatus string            `json:"microcontrollerStatus"`
	SprinklerStatus       []SprinklerStatus `json:"sprinklerStatus"`
}
type SprinklerStatus struct {
	ZoneName string `json:"zoneName"`
	IsActive bool   `json:"isActive"`
}
type Api struct {
	Ctx    context.Context
	Logger *slog.Logger
	Wg     *sync.WaitGroup
	Mux    *http.ServeMux
}

func NewApi(ctx context.Context, logger *slog.Logger, wg *sync.WaitGroup) *Api {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", pingHandler)
	mux.HandleFunc("GET /system-status", getSystemStatus)

	return &Api{
		Ctx:    ctx,
		Mux:    mux,
		Logger: logger,
		Wg:     wg,
	}
}

func (a *Api) Run() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.Mux,
	}

	a.Logger.Info("Starting API...")
	err := server.ListenAndServe()
	if err != nil {
		a.Logger.Error("Error starting API", "error", err)
	}

	<-a.Ctx.Done()

	a.Logger.Info("Done context signal received in API - exiting")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		a.Logger.Error("Error shutting down http server", "error", err)
	}

	a.Wg.Done()
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"ping": "OK",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func getSystemStatus(w http.ResponseWriter, r *http.Request) {

	resp := StatusResponse{
		MicrocontrollerStatus: "OKAY",
	}

	for i := 0; i < 5; i++ {
		s := SprinklerStatus{
			ZoneName: fmt.Sprintf("zone%d", i),
			IsActive: false,
		}
		resp.SprinklerStatus = append(resp.SprinklerStatus, s)
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
