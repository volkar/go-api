package main

import (
	"api/internal/domain/albums/albumtypes"
	"api/internal/platform/request"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

/* Routes for development, must be deleted in production */
func (app *app) RoutesPlayground(r *chi.Mux) {
	r.Route("/playground", func(r chi.Router) {
		// Create admin with albums
		r.Get("/create_admin", app.PlaygroundCreateAdmin)
		// Create user with albums
		r.Get("/create_user", app.PlaygroundCreateUser)
		// Create new refresh token and set both access and refresh tokens as cookie for user
		r.Get("/get_user_cookies", app.PlaygroundGetUserCookies)
		// Create new refresh token and set both access and refresh tokens as cookie for admin
		r.Get("/get_admin_cookies", app.PlaygroundGetAdminCookies)
		// Clear token cookies
		r.Get("/clear_cookies", app.PlaygroundClearCookies)
		// Clear redis cache
		r.Get("/clear_cache", app.PlaygroundClearCache)
	})
}

/* Creates admin with 3 albums */
func (app *app) PlaygroundCreateAdmin(w http.ResponseWriter, r *http.Request) {
	admin, err := app.usersService.Create(r.Context(), "admin@test.test", "Almighty Admin", "admin", "admin")
	if err != nil {
		// User may be already created. Ignore
		app.response.SuccessDataOnly(w, r, map[string]string{
			"playground": "already created",
		})
		return
	}

	// Create admin albums
	var adminAtlas = []albumtypes.AtlasItem{
		{
			Type: "title",
			Src:  "Album title",
		},
		{
			Type: "text",
			Src:  "The woods are lovely, dark and deep",
		},
	}
	var adminAccessPrivate = albumtypes.Access{
		Type:  "private",
		Share: []string{},
	}
	app.albumsService.Create(r.Context(), admin.ID, "Admin private album", "private", adminAtlas, adminAccessPrivate, time.Date(2021, 3, 1, 12, 30, 0, 0, time.UTC))
	var adminAccessPublic = albumtypes.Access{
		Type:  "public",
		Share: []string{},
	}
	app.albumsService.Create(r.Context(), admin.ID, "Admin public album", "public", adminAtlas, adminAccessPublic, time.Date(2022, 5, 13, 12, 30, 0, 0, time.UTC))
	var adminAccessShared = albumtypes.Access{
		Type:  "shared",
		Share: []string{"user@test.test"},
	}
	app.albumsService.Create(r.Context(), admin.ID, "Admin shared album", "shared", adminAtlas, adminAccessShared, time.Date(2023, 7, 22, 12, 30, 0, 0, time.UTC))

	app.response.SuccessDataOnly(w, r, map[string]string{
		"playground": "admin created",
	})
}

/* Creates user with 3 albums */
func (app *app) PlaygroundCreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := app.usersService.Create(r.Context(), "user@test.test", "Just User", "user", "user")
	if err != nil {
		// User may be already created. Ignore
		app.response.SuccessDataOnly(w, r, map[string]string{
			"playground": "already created",
		})
		return
	}

	// Create user albums
	var userAtlas = []albumtypes.AtlasItem{
		{
			Type: "title",
			Src:  "Album title",
		},
		{
			Type: "text",
			Src:  "The woods are lovely, dark and deep",
		},
	}
	var userAccessPrivate = albumtypes.Access{
		Type:  "private",
		Share: []string{},
	}
	app.albumsService.Create(r.Context(), user.ID, "User private album", "private", userAtlas, userAccessPrivate, time.Date(2021, 3, 1, 12, 30, 0, 0, time.UTC))
	var userAccessPublic = albumtypes.Access{
		Type:  "public",
		Share: []string{},
	}
	app.albumsService.Create(r.Context(), user.ID, "User public album", "public", userAtlas, userAccessPublic, time.Date(2022, 5, 13, 12, 30, 0, 0, time.UTC))
	var userAccessShared = albumtypes.Access{
		Type:  "shared",
		Share: []string{"admin@test.test"},
	}
	app.albumsService.Create(r.Context(), user.ID, "User shared album", "shared", userAtlas, userAccessShared, time.Date(2023, 7, 22, 12, 30, 0, 0, time.UTC))

	app.response.SuccessDataOnly(w, r, map[string]string{
		"playground": "user created",
	})
}

/* Set access and refresh cookie for user */
func (app *app) PlaygroundGetUserCookies(w http.ResponseWriter, r *http.Request) {
	u, err := app.usersService.GetAvailableBySlug(r.Context(), "user")
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"user_tokens": "error, user not found",
		})
		return
	}

	// Delete all existed user tokens
	app.tokens.DeleteAllRefreshForUser(r.Context(), u.ID)

	// Get metadata for refresh token
	meta := request.GetMetaFromRequest(r)

	// Create new refresh token
	refresh, err := app.tokens.GenerateRefreshString()
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"user_tokens": "error, failed to generate refresh token",
		})
		return
	}
	hash := app.tokens.Hash(refresh)
	_, err = app.tokens.CreateRefresh(r.Context(), u.ID, hash, time.Now().Add(app.cfg.Paseto.RefreshTTL), meta)
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"user_tokens": "error, failed to create refresh token",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   app.cfg.Cookie.Secure,
		SameSite: app.cfg.Cookie.SameSite,
		MaxAge:   int(app.cfg.Paseto.RefreshTTL.Seconds()),
	})

	// Create access token
	access := app.tokens.CreateAccess(u.ID, u.Role, u.Email, app.cfg.Paseto.AccessTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.cfg.Cookie.Secure,
		SameSite: app.cfg.Cookie.SameSite,
		MaxAge:   int(app.cfg.Paseto.AccessTTL.Seconds()),
	})

	app.response.SuccessDataOnly(w, r, map[string]string{
		"user_tokens": "set",
	})
}

/* Set access and refresh cookie for admin */
func (app *app) PlaygroundGetAdminCookies(w http.ResponseWriter, r *http.Request) {
	u, err := app.usersService.GetAvailableBySlug(r.Context(), "admin")
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"admin_tokens": "error, admin not found",
		})
		return
	}

	// Delete all existed admin tokens
	app.tokens.DeleteAllRefreshForUser(r.Context(), u.ID)

	// Get metadata for refresh token
	meta := request.GetMetaFromRequest(r)

	// Create new refresh token
	refresh, err := app.tokens.GenerateRefreshString()
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"admin_tokens": "error, failed to generate refresh token",
		})
		return
	}
	hash := app.tokens.Hash(refresh)
	_, err = app.tokens.CreateRefresh(r.Context(), u.ID, hash, time.Now().Add(app.cfg.Paseto.RefreshTTL), meta)
	if err != nil {
		app.response.SuccessDataOnly(w, r, map[string]string{
			"admin_tokens": "error, failed to create refresh token",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   app.cfg.Cookie.Secure,
		SameSite: app.cfg.Cookie.SameSite,
		MaxAge:   int(app.cfg.Paseto.RefreshTTL.Seconds()),
	})

	// Create access token
	access := app.tokens.CreateAccess(u.ID, u.Role, u.Email, app.cfg.Paseto.AccessTTL)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.cfg.Cookie.Secure,
		SameSite: app.cfg.Cookie.SameSite,
		MaxAge:   int(app.cfg.Paseto.AccessTTL.Seconds()),
	})

	app.response.SuccessDataOnly(w, r, map[string]string{
		"admin_tokens": "set",
	})
}

/* Clears current access and refresh token cookies */
func (app *app) PlaygroundClearCookies(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
	})

	app.response.SuccessDataOnly(w, r, map[string]string{
		"cookies": "clear",
	})
}

/* Clears redis cache */
func (app *app) PlaygroundClearCache(w http.ResponseWriter, r *http.Request) {
	app.cache.ClearFullCache(r.Context())

	app.response.SuccessDataOnly(w, r, map[string]string{
		"cache": "clear",
	})
}
