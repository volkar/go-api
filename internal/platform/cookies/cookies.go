package cookies

import (
	"net/http"
	"time"
)

type Manager struct {
	secure     bool
	sameSite   http.SameSite
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(secure bool, sameSite http.SameSite, accessTTL time.Duration, refreshTTL time.Duration) *Manager {
	return &Manager{
		secure:     secure,
		sameSite:   sameSite,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

/* Set access token cookie */
func (m *Manager) SetAccessCookie(w http.ResponseWriter, access string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: m.sameSite,
		MaxAge:   int(m.accessTTL.Seconds()),
	})
}

/* Set refresh token cookie */
func (m *Manager) SetRefreshCookie(w http.ResponseWriter, refresh string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: m.sameSite,
		MaxAge:   int(m.refreshTTL.Seconds()),
	})
}

/* Unset access token cookie */
func (m *Manager) UnsetAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

/* Unset refresh token cookie */
func (m *Manager) UnsetRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
