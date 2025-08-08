package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/services"
	"github.com/MoeMahhouk/go-tcb-notify/pkg/models"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config          *config.Config
	db              *sql.DB
	tcbFetcher      *services.TCBFetcher
	quoteChecker    *services.QuoteChecker
	alertPublisher  *services.AlertPublisher
	fmspcService    *services.FMSPCService
	registryService *services.RegistryService
	router          *mux.Router
}

func New(cfg *config.Config, db *sql.DB, tcbFetcher *services.TCBFetcher, quoteChecker *services.QuoteChecker, alertPublisher *services.AlertPublisher, fmspcService *services.FMSPCService, registryService *services.RegistryService) *http.Server {
	s := &Server{
		config:          cfg,
		db:              db,
		tcbFetcher:      tcbFetcher,
		quoteChecker:    quoteChecker,
		alertPublisher:  alertPublisher,
		fmspcService:    fmspcService,
		registryService: registryService,
		router:          mux.NewRouter(),
	}

	s.setupRoutes()

	return &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: s.router,
	}
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/ready", s.readyHandler).Methods("GET")

	if s.config.MetricsEnabled {
		s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	}

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/tcb/{fmspc}", s.getTCBInfoHandler).Methods("GET")
	api.HandleFunc("/quotes", s.getQuotesHandler).Methods("GET")
	api.HandleFunc("/quotes/{address}/alerts", s.getQuoteAlertsHandler).Methods("GET")
	api.HandleFunc("/alerts/{id}/acknowledge", s.acknowledgeAlertHandler).Methods("POST")
	api.HandleFunc("/fmspcs", s.getFMSPCsHandler).Methods("GET")
	api.HandleFunc("/fmspcs/refresh", s.refreshFMSPCsHandler).Methods("POST")
	api.HandleFunc("/registry/refresh", s.refreshRegistryHandler).Methods("POST")
}

func (s *Server) getFMSPCsHandler(w http.ResponseWriter, r *http.Request) {
	fmspcs, err := s.fmspcService.GetAllFMSPCs()
	if err != nil {
		http.Error(w, "Failed to get FMSPCs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fmspcs)
}

func (s *Server) refreshFMSPCsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.fmspcService.FetchAndStoreAllFMSPCs(ctx); err != nil {
		http.Error(w, "Failed to refresh FMSPCs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "FMSPCs refreshed successfully"}`))
}

func (s *Server) refreshRegistryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.registryService.FetchQuotesFromRegistry(ctx); err != nil {
		http.Error(w, "Failed to refresh registry quotes", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Registry quotes refreshed successfully"}`))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database not ready"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

func (s *Server) getTCBInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmspc := vars["fmspc"]

	if fmspc == "" {
		http.Error(w, "FMSPC parameter required", http.StatusBadRequest)
		return
	}

	tcbInfo, err := s.getCurrentTCBInfo(fmspc)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "TCB info not found for FMSPC", http.StatusNotFound)
		} else {
			logrus.WithError(err).Error("Failed to get TCB info")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tcbInfo)
}

func (s *Server) getQuotesHandler(w http.ResponseWriter, r *http.Request) {
	quotes, err := s.registryService.GetMonitoredQuotes()
	if err != nil {
		logrus.WithError(err).Error("Failed to get monitored quotes")
		http.Error(w, "Failed to get quotes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotes)
}

func (s *Server) getQuoteAlertsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	if address == "" {
		http.Error(w, "Address parameter required", http.StatusBadRequest)
		return
	}

	// Get limit from query parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alerts, err := s.alertPublisher.GetAlertHistory(address, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get alert history")
		http.Error(w, "Failed to get alerts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func (s *Server) acknowledgeAlertHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	if err := s.alertPublisher.AcknowledgeAlert(id); err != nil {
		logrus.WithError(err).Error("Failed to acknowledge alert")
		http.Error(w, "Failed to acknowledge alert", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Alert acknowledged"}`))
}

func (s *Server) getCurrentTCBInfo(fmspc string) (interface{}, error) {
	query := `
		SELECT fmspc, version, issue_date, next_update, tcb_type, 
		       tcb_evaluation_data_number, tcb_levels, raw_response, created_at
		FROM tdx_tcb_info 
		WHERE fmspc = $1 
		ORDER BY tcb_evaluation_data_number DESC 
		LIMIT 1`

	var result models.TCBInfo

	var tcbLevelsBytes, rawResponseBytes []byte
	err := s.db.QueryRow(query, fmspc).Scan(
		&result.FMSPC,
		&result.Version,
		&result.IssueDate,
		&result.NextUpdate,
		&result.TCBType,
		&result.TCBEvaluationDataNumber,
		&tcbLevelsBytes,
		&rawResponseBytes,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if err := json.Unmarshal(tcbLevelsBytes, &result.TCBLevels); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(rawResponseBytes, &result.RawResponse); err != nil {
		return nil, err
	}

	return result, nil
}
