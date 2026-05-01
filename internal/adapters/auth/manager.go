package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Session struct {
	Subject   string    `json:"sub"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	TenantID  string    `json:"tid,omitempty"`
	Provider  string    `json:"provider"`
	ExpiresAt time.Time `json:"exp"`
}

type statePayload struct {
	State        string    `json:"state"`
	Nonce        string    `json:"nonce"`
	CodeVerifier string    `json:"code_verifier"`
	ExpiresAt    time.Time `json:"exp"`
}

type Manager struct {
	mode        string
	required    bool
	callbackURL string
	oauth       *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	session     *SecureCookie
	state       *SecureCookie
}

func NewFromEnv(ctx context.Context) (*Manager, error) {
	mode := strings.ToLower(valueOrDefault(os.Getenv("AUTH_MODE"), "dev"))
	secureCookie := !strings.EqualFold(os.Getenv("COOKIE_SECURE"), "false") && mode != "dev"
	session, err := NewSecureCookie("iis_session", os.Getenv("AUTH_SESSION_KEY"), secureCookie)
	if err != nil {
		return nil, err
	}
	state, err := NewSecureCookie("iis_oauth_state", os.Getenv("AUTH_SESSION_KEY"), secureCookie)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		mode:        mode,
		required:    strings.EqualFold(os.Getenv("AUTH_REQUIRED"), "true"),
		callbackURL: valueOrDefault(os.Getenv("AUTH_REDIRECT_URL"), "http://localhost:8080/auth/callback"),
		session:     session,
		state:       state,
	}
	if mode != "azure" {
		return m, nil
	}
	tenant := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	if tenant == "" || clientID == "" || clientSecret == "" {
		return nil, errors.New("AZURE_TENANT_ID, AZURE_CLIENT_ID, and AZURE_CLIENT_SECRET are required when AUTH_MODE=azure")
	}
	provider, err := oidc.NewProvider(ctx, "https://login.microsoftonline.com/"+tenant+"/v2.0")
	if err != nil {
		return nil, err
	}
	m.oauth = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  m.callbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	m.verifier = provider.Verifier(&oidc.Config{ClientID: clientID})
	return m, nil
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secureHeaders(w)
		if !m.required {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := m.Current(r); !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Manager) Config(r *http.Request) map[string]any {
	session, ok := m.Current(r)
	cfg := map[string]any{
		"mode":          m.mode,
		"required":      m.required,
		"authenticated": ok,
		"login_url":     "/auth/login",
		"logout_url":    "/auth/logout",
		"storage":       "encrypted HttpOnly SameSite cookie",
		"transport":     "TLS required in deployed environments; local demo uses localhost HTTP only",
	}
	if ok {
		cfg["user"] = session
	}
	return cfg
}

func (m *Manager) Current(r *http.Request) (Session, bool) {
	var session Session
	if err := m.session.Read(r, &session); err != nil {
		return Session{}, false
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return Session{}, false
	}
	return session, true
}

func (m *Manager) Login(w http.ResponseWriter, r *http.Request) {
	if m.mode != "azure" {
		_ = m.session.Set(w, Session{
			Subject:   "local-dev-user",
			Email:     "operator@local.dev",
			Name:      "Local Operator",
			Provider:  "dev",
			ExpiresAt: time.Now().UTC().Add(8 * time.Hour),
		}, time.Now().UTC().Add(8*time.Hour))
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	state, nonce, verifier := randomToken(), randomToken(), oauth2.GenerateVerifier()
	payload := statePayload{State: state, Nonce: nonce, CodeVerifier: verifier, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	if err := m.state.Set(w, payload, payload.ExpiresAt); err != nil {
		http.Error(w, "could not create auth state", http.StatusInternalServerError)
		return
	}
	url := m.oauth.AuthCodeURL(state, oidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier))
	http.Redirect(w, r, url, http.StatusFound)
}

func (m *Manager) Callback(w http.ResponseWriter, r *http.Request) {
	if m.mode != "azure" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var expected statePayload
	if err := m.state.Read(r, &expected); err != nil || expected.State != r.URL.Query().Get("state") || time.Now().UTC().After(expected.ExpiresAt) {
		http.Error(w, "invalid auth state", http.StatusBadRequest)
		return
	}
	token, err := m.oauth.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(expected.CodeVerifier))
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusBadGateway)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "id_token missing", http.StatusBadGateway)
		return
	}
	idToken, err := m.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token verification failed", http.StatusUnauthorized)
		return
	}
	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
		TenantID          string `json:"tid"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "claims parse failed", http.StatusBadGateway)
		return
	}
	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}
	expires := time.Now().UTC().Add(time.Hour)
	if idToken.Expiry.After(time.Now()) {
		expires = idToken.Expiry
	}
	_ = m.session.Set(w, Session{
		Subject:   idToken.Subject,
		Email:     email,
		Name:      claims.Name,
		TenantID:  claims.TenantID,
		Provider:  "azure",
		ExpiresAt: expires,
	}, expires)
	m.state.Clear(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	m.session.Clear(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func randomToken() string {
	data := make([]byte, 32)
	_, _ = rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}

func secureHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
