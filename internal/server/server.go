package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/MoeMahhouk/go-tcb-notify/internal/config"
	"github.com/MoeMahhouk/go-tcb-notify/internal/services"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config         *config.Config
	db             *sql.DB
	tcbFetcher     *services.TCBFetcher
	quoteChecker   *services.QuoteChecker
	alertPublisher *services.AlertPublisher
	fmspcService   *services.FMSPCService
	router         *mux.Router
}

func New(cfg *config.Config, db *sql.DB, tcbFetcher *services.TCBFetcher, quoteChecker *services.QuoteChecker, alertPublisher *services.AlertPublisher, fmspcService *services.FMSPCService) *http.Server {
	s := &Server{
		config:         cfg,
		db:             db,
		tcbFetcher:     tcbFetcher,
		quoteChecker:   quoteChecker,
		alertPublisher: alertPublisher,
		fmspcService:   fmspcService,
		router:         mux.NewRouter(),
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
	api.HandleFunc("/fmspcs", s.getFMSPCsHandler).Methods("GET")
	api.HandleFunc("/fmspcs/refresh", s.refreshFMSPCsHandler).Methods("POST")
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

	// Implementation would query database for TCB info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "TCB info for ` + fmspc + `"}`))
}

func (s *Server) getQuotesHandler(w http.ResponseWriter, r *http.Request) {
	// Implementation would return monitored quotes
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"quotes": []}`))
}
