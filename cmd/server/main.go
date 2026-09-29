package main

import (
	"context"
	"log"
	"net/http"

	"iot-backend/internal/api"
	"iot-backend/internal/config"
	"iot-backend/internal/db"
	"iot-backend/internal/ingest"
	"iot-backend/internal/repository"
	"iot-backend/internal/status"
	"iot-backend/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	deviceRepo := repository.NewDeviceRepo(pool)
	heartbeatRepo := repository.NewHeartbeatRepo(pool)
	detectionRepo := repository.NewDetectionRepo(pool)
	sensorRepo := repository.NewSensorRepo(pool)

	resolver := status.NewResolver(heartbeatRepo, cfg.OnlineThresholdSeconds)
	hub := ws.NewHub()

	router := &api.Router{
		Dashboard:  api.NewDashboardHandler(deviceRepo, sensorRepo, detectionRepo, resolver),
		Devices:    api.NewDeviceHandler(deviceRepo, sensorRepo, resolver),
		Stats:      api.NewStatsHandler(detectionRepo),
		Detections: api.NewDetectionsHandler(detectionRepo),
		Ingest:     ingest.NewHandlers(deviceRepo, heartbeatRepo, detectionRepo, sensorRepo, hub),
		Hub:        hub,
	}

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router.Build()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
