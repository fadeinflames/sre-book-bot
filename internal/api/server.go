package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"sre-learning-bot/internal/app"
)

type Server struct {
	logger       *slog.Logger
	store        *app.Store
	httpServer   *http.Server
	requestCount atomic.Uint64
}

func New(logger *slog.Logger, addr string, store *app.Store) *Server {
	s := &Server{
		logger: logger,
		store:  store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/metrics", s.metrics)
	mux.HandleFunc("/modules", s.modules)
	mux.HandleFunc("/users/progress", s.userProgress)
	mux.HandleFunc("/mentor/reports", s.mentorReports)
	mux.HandleFunc("/mentor/checklists/review", s.reviewChecklist)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.withMetrics(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Run() error {
	s.logger.Info("api server started", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requestCount.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("http_requests_total " + strconv.FormatUint(s.requestCount.Load(), 10) + "\n"))
}

func (s *Server) modules(w http.ResponseWriter, r *http.Request) {
	telegramID, err := parseInt64Required(r.URL.Query().Get("telegram_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user, err := s.store.GetUserByTelegramID(r.Context(), telegramID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	roadmap, err := s.store.GetRoadmap(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get roadmap"})
		return
	}
	writeJSON(w, http.StatusOK, roadmap)
}

func (s *Server) userProgress(w http.ResponseWriter, r *http.Request) {
	telegramID, err := parseInt64Required(r.URL.Query().Get("telegram_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	user, err := s.store.GetUserByTelegramID(r.Context(), telegramID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	roadmap, err := s.store.GetRoadmap(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get progress"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"telegram_id": user.TelegramID,
		"username":    user.Username,
		"roadmap":     roadmap,
	})
}

func (s *Server) mentorReports(w http.ResponseWriter, r *http.Request) {
	reports, err := s.store.ListJuniorReports(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to build report"})
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func (s *Server) reviewChecklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		ReviewerTelegramID int64  `json:"reviewer_telegram_id"`
		SubmissionID       int64  `json:"submission_id"`
		Action             string `json:"action"`
		Comment            string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	reviewer, err := s.store.GetUserByTelegramID(r.Context(), req.ReviewerTelegramID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "reviewer not found"})
		return
	}
	if reviewer.Role != app.RoleMentor && reviewer.Role != app.RoleAdmin {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "mentor/admin required"})
		return
	}
	approve := strings.EqualFold(req.Action, "approve")
	if !approve && !strings.EqualFold(req.Action, "rework") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be approve|rework"})
		return
	}
	if err := s.store.ReviewChecklist(r.Context(), reviewer.ID, req.SubmissionID, approve, req.Comment); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed review"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func parseInt64Required(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(value, 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
