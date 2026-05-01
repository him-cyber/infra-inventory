package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"

	"github.com/him-cyber/infra-inventory-stream/internal/adapters/auth"
	"github.com/him-cyber/infra-inventory-stream/internal/core/domain"
	"github.com/him-cyber/infra-inventory-stream/internal/core/service"
)

type Server struct {
	app  *service.Service
	log  *slog.Logger
	auth *auth.Manager
}

func New(app *service.Service, log *slog.Logger, authManager *auth.Manager) http.Handler {
	s := &Server{app: app, log: log, auth: authManager}
	r := chi.NewRouter()
	r.Use(securityHeaders)
	r.Get("/healthz", s.health)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", s.auth.Login)
		r.Get("/callback", s.auth.Callback)
		r.Get("/logout", s.auth.Logout)
	})
	r.Group(func(r chi.Router) {
		r.Use(s.auth.Middleware)
		r.Route("/api", func(r chi.Router) {
			r.Get("/auth/session", s.authSession)
			r.Post("/org/import", s.importOrg)
			r.Get("/org/imports", s.orgImports)
			r.Post("/assets", s.upsertAsset)
			r.Get("/assets/search", s.searchAssets)
			r.Get("/assets/{id}", s.getAsset)
			r.Get("/topology", s.topology)
			r.Get("/analytics", s.analytics)
			r.Post("/config/recommendations", s.configRecommendations)
			r.Post("/security/analyze", s.startSecurityAnalysis)
			r.Post("/servicenow/tickets", s.createServiceNowTicket)
			r.Get("/stream/recent", s.recentEvents)
			r.Post("/stream/replay", s.replayEvents)
			r.Get("/config/version", s.configVersion)
		})
	})
	return r
}

func (s *Server) authSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.auth.Config(r))
}

func (s *Server) importOrg(w http.ResponseWriter, r *http.Request) {
	var req domain.OrgImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.app.ImportOrg(r.Context(), tenant(r), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (s *Server) orgImports(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"imports": s.app.OrgImports()})
}

func (s *Server) configRecommendations(w http.ResponseWriter, r *http.Request) {
	recommendations, err := s.app.ConfigIntelligence(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, recommendations)
}

func (s *Server) topology(w http.ResponseWriter, r *http.Request) {
	topology, err := s.app.Topology(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, topology)
}

func (s *Server) analytics(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.app.Analytics(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) startSecurityAnalysis(w http.ResponseWriter, r *http.Request) {
	report, err := s.app.StartCVEAnalysis(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, report)
}

func (s *Server) createServiceNowTicket(w http.ResponseWriter, r *http.Request) {
	ticket, err := s.app.CreateServiceNowTicket(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, ticket)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) upsertAsset(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("api").Start(r.Context(), "upsert_asset")
	defer span.End()
	var asset domain.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	event, err := s.app.UpsertAsset(ctx, tenant(r), asset)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, event)
}

func (s *Server) searchAssets(w http.ResponseWriter, r *http.Request) {
	q := domain.SearchQuery{
		Text:        r.URL.Query().Get("q"),
		Type:        r.URL.Query().Get("type"),
		Environment: r.URL.Query().Get("env"),
		Owner:       r.URL.Query().Get("owner"),
	}
	result, err := s.app.Search(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getAsset(w http.ResponseWriter, r *http.Request) {
	asset, ok, err := s.app.GetAsset(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
		return
	}
	writeJSON(w, http.StatusOK, asset)
}

func (s *Server) recentEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"events": s.app.RecentEvents()})
}

func (s *Server) replayEvents(w http.ResponseWriter, r *http.Request) {
	count, err := s.app.Replay(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]int{"replayed": count})
}

func (s *Server) configVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.app.ConfigVersion())
}

func tenant(r *http.Request) string {
	if value := r.Header.Get("X-Tenant"); value != "" {
		return value
	}
	return "demo"
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
