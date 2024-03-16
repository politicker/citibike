package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/politicker/betterbike-api/internal/api"
	"github.com/politicker/betterbike-api/internal/db"
	"github.com/politicker/betterbike-api/internal/domain"
	"go.uber.org/zap"
)

type Server struct {
	logger    *zap.Logger
	queries   *db.Queries
	port      string
	bikesRepo *domain.BikesRepo
}

func NewServer(port string, queries *db.Queries, logger *zap.Logger) Server {
	return Server{
		queries:   queries,
		port:      port,
		logger:    logger,
		bikesRepo: domain.NewBikesRepo(queries, logger),
	}
}

func (s *Server) Start() {
	http.HandleFunc("/", s.GetBikes)

	s.logger.Info("listening", zap.String("port", s.port))
	http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) renderError(w http.ResponseWriter, message string, errorCode string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.logger.Error(message)

	json.NewEncoder(w).Encode(map[string]string{
		"error":     message,
		"errorCode": errorCode,
	})
}

func (s *Server) GetBikes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	start := time.Now()

	w.Header().Set("Content-Type", "application/json")

	var stationParams db.GetStationsParams
	err := json.NewDecoder(r.Body).Decode(&stationParams)
	if err != nil {
		s.renderError(w, "lat and lon are required", "invalid-json-payload")
		return
	}

	if stationParams.Lat == 0 || stationParams.Lon == 0 {
		s.renderError(w, fmt.Sprintf("invalid lat or lon: %f, %f", stationParams.Lat, stationParams.Lon), "missing-coords")
		return
	}

	stations, err := s.bikesRepo.GetNearbyStationEbikes(ctx, stationParams)
	if err != nil {
		s.renderError(w, "error fetching stations", "internal-error")
		return
	}
	if len(stations) == 0 {
		s.renderError(w, "No ebikes nearby. Are you in New York City?", "too-far-away")
		return
	}

	json.NewEncoder(w).Encode(api.Home{
		LastUpdated: stations[0].CreatedAt,
		Stations:    stations,
	})

	s.logger.Info(
		fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, time.Since(start)),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Duration("duration", time.Since(start)),
	)
}
