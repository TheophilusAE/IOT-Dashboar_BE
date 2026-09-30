package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"iot-backend/internal/ingest"
	"iot-backend/internal/ws"
)

type Router struct {
	Dashboard  *DashboardHandler
	Devices    *DeviceHandler
	Stats      *StatsHandler
	Detections *DetectionsHandler
	Ingest     *ingest.Handlers
	Hub        *ws.Hub
}

func (rt *Router) Build() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// The Flutter app's web build runs on its own origin (e.g. the flutter
	// run/dev server) and calls this API cross-origin, so the browser
	// requires CORS headers here. There's no browser-facing auth to protect,
	// so any origin is allowed.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-Device-UID"},
		AllowCredentials: false,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/dashboard/summary", rt.Dashboard.Summary)
		r.Get("/devices", rt.Devices.List)
		r.Get("/devices/{id}/status", rt.Devices.Status)
		r.Get("/stats", rt.Stats.Stats)
		r.Get("/stats/daily", rt.Stats.DailyStats)
		r.Get("/detections", rt.Detections.List)

		r.Post("/ingest/heartbeat", rt.Ingest.Heartbeat)
		r.Post("/ingest/detection", rt.Ingest.Detection)
		r.Post("/ingest/sensor", rt.Ingest.Sensor)

		r.Get("/ws", ws.Handler(rt.Hub))
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return r
}
