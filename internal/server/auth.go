package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	localauth "aiport/internal/auth"
	"aiport/internal/config"
	"aiport/internal/tasks"
	"aiport/internal/telemetry"
)

const sessionCookie = "aiport_session"

func NewConfigured(c config.Config, s config.Secrets, collector *telemetry.Collector, path string) http.Handler {
	h := NewWithTelemetry(c, s, collector).(*Handler)
	h.live = config.NewLive(path, c)
	h.live.SetClock(h.clock)
	if path != "" {
		h.worktrees = path + ".worktrees" // before keepChanges: its startup cleanup removes worktrees here
	}
	h.keepChanges(path)
	h.live.OnChange(h.applyLive)
	h.keys = config.EnvKeys(path)
	h.auth = localauth.New(c.Auth, func(updated config.Auth) error {
		_, _, err := h.live.Change("sign-in settings", config.SetAuth(updated))
		return err
	})
	h.refreshKeys(c)
	if store, err := tasks.Open(tasks.Path(path)); err == nil {
		h.tasks = store
	}
	return h
}
func (h *Handler) sessionValid(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	return err == nil && h.auth.Valid(cookie.Value)
}
func (h *Handler) setSession(w http.ResponseWriter, r *http.Request, value string) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: value, Path: "/api/v1/", HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil, MaxAge: int(localauth.SessionDuration.Seconds())})
}
func (h *Handler) status(r *http.Request) map[string]any {
	s := h.auth.Settings()
	return map[string]any{"configured": s.PasswordHash != "", "disabled": s.Disabled, "authenticated": s.Disabled || h.sessionValid(r), "username_required": s.Username != "", "token_env": s.TokenEnv}
}
func decodeAuth(w http.ResponseWriter, r *http.Request, value any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("use application/json")
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return errors.New("invalid authentication request")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("invalid authentication request")
	}
	return nil
}
func (h *Handler) authRoutes(w http.ResponseWriter, r *http.Request, management bool, fail func(int, string, string)) bool {
	path := r.URL.Path
	if path == "/api/v1/auth/status" {
		if r.Method != "GET" {
			fail(405, "method_not_allowed", "Use GET")
		} else {
			writeJSON(w, 200, h.status(r))
		}
		return true
	}
	if !strings.HasPrefix(path, "/api/v1/auth/") {
		return false
	}
	if r.Method != "POST" {
		fail(405, "method_not_allowed", "Use POST")
		return true
	}
	if r.Header.Get("X-Aiport-CSRF") != "1" {
		fail(403, "csrf_required", "Use the same-origin dashboard to change authentication")
		return true
	}
	switch path {
	case "/api/v1/auth/login":
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := decodeAuth(w, r, &input); err != nil {
			fail(400, "invalid_request", err.Error())
			return true
		}
		session, err := h.auth.Authenticate(input.Username, input.Password)
		if err != nil {
			fail(401, "login_failed", err.Error())
			return true
		}
		h.setSession(w, r, session)
		writeJSON(w, 200, map[string]bool{"ok": true})
	case "/api/v1/auth/recover":
		if !management {
			fail(401, "invalid_recovery_token", "A valid management token is required; agent tokens cannot recover access")
			return true
		}
		session, err := h.auth.NewSession()
		if err != nil {
			fail(429, "session_unavailable", err.Error())
			return true
		}
		h.setSession(w, r, session)
		writeJSON(w, 200, map[string]bool{"ok": true})
	case "/api/v1/auth/logout":
		if cookie, err := r.Cookie(sessionCookie); err == nil {
			h.auth.Revoke(cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{Name: sessionCookie, Path: "/api/v1/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: -1})
		writeJSON(w, 200, map[string]bool{"ok": true})
	case "/api/v1/auth/settings":
		if !management && !h.sessionValid(r) && !h.auth.Disabled() {
			fail(401, "unauthorized", "Sign in before changing authentication")
			return true
		}
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Disabled bool   `json:"disabled"`
		}
		if err := decodeAuth(w, r, &input); err != nil {
			fail(400, "invalid_request", err.Error())
			return true
		}
		session, err := h.auth.Update(input.Username, input.Password, input.Disabled)
		if err != nil {
			fail(400, "auth_update_failed", err.Error())
			return true
		}
		h.setSession(w, r, session)
		writeJSON(w, 200, map[string]bool{"ok": true})
	default:
		fail(404, "unsupported_path", "Authentication path is unavailable")
	}
	return true
}
