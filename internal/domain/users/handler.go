package users

import (
	"api/internal/domain/albums"
	"api/internal/domain/tokens"
	"api/internal/platform/cookies"
	"api/internal/platform/request"
	"api/internal/platform/response"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	users     *Service
	response  *response.Response
	validator *validator.Validate
	cookies   *cookies.Manager
}

func NewHandler(service *Service, response *response.Response, val *validator.Validate, cookies *cookies.Manager) *Handler {
	return &Handler{
		users:     service,
		response:  response,
		validator: val,
		cookies:   cookies,
	}
}

/* Get album list by user slug */
func (h *Handler) AlbumList(w http.ResponseWriter, r *http.Request) {
	userSlug := chi.URLParam(r, "slug")
	// Get claims from context
	claims, _ := tokens.GetClaimsFromContext(r.Context())
	// Parse pagination parameters
	query := r.URL.Query()
	cursor := query.Get("cursor")
	limit := 50
	if limitParam := query.Get("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			limit = parsed
		}
	}
	// Get album list
	a, nextCursor, err := h.users.AlbumList(r.Context(), userSlug, claims.UserID, claims.Email, cursor, limit)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	// Map to public
	albums := albums.ToPublicAlbumList(a)

	// Return albums
	h.response.Paginated(w, r, albums, nextCursor)
}

/* Get authenticated user info */
func (h *Handler) CurrentInfo(w http.ResponseWriter, r *http.Request) {
	// Get claims from context
	claims, ok := tokens.GetClaimsFromContext(r.Context())
	if !ok {
		h.response.Error(w, r, response.ErrNoClaims)
		return
	}
	// Get user by ID
	u, err := h.users.GetAvailable(r.Context(), claims.UserID)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	h.response.SuccessDataOnly(w, r, u)
}

/* Get user info by user slug */
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	userSlug := chi.URLParam(r, "slug")
	// Get user
	u, err := h.users.GetAvailableBySlug(r.Context(), userSlug)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	// Map to public user
	user := ToPublic(u)

	// Return user info
	h.response.SuccessDataOnly(w, r, user)
}

/* Update user info */
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// Parse UUID
	idStr := chi.URLParam(r, "uuid")
	updatingUserID, err := uuid.Parse(idStr)
	if err != nil {
		h.response.Error(w, r, response.ErrBadUUID.Wrap(err))
		return
	}

	// JSON decode
	input := struct {
		Username string `json:"username" validate:"required,min=2,max=255"`
		Slug     string `json:"slug" validate:"required,min=2,max=255,slug,notreserved"`
	}{}
	if err := request.DecodeJSONBody(w, r, &input); err != nil {
		h.response.Error(w, r, response.ErrBadJSON.Wrap(err))
		return
	}

	// Validate input
	if err := h.validator.Struct(&input); err != nil {
		h.response.ValidationError(w, r, err)
		return
	}

	// Get user claims
	claims, ok := tokens.GetClaimsFromContext(r.Context())
	if !ok {
		h.response.Error(w, r, response.ErrNoClaims)
		return
	}

	// Update user
	u, err := h.users.Update(r.Context(), claims.UserID, updatingUserID, input.Slug, input.Username)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	h.response.SuccessWithData(w, r, response.SuccessUserUpdated, u)
}

/* Delete user */
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {

	// User id from url
	idStr := chi.URLParam(r, "uuid")
	deletingUserID, err := uuid.Parse(idStr)
	if err != nil {
		h.response.Error(w, r, response.ErrBadUUID.Wrap(err))
		return
	}

	// Get user claims
	claims, ok := tokens.GetClaimsFromContext(r.Context())
	if !ok {
		h.response.Error(w, r, response.ErrNoClaims)
		return
	}

	// Delete user
	_, err = h.users.Delete(r.Context(), claims.UserID, deletingUserID)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	// Delete cookies
	h.cookies.UnsetAccessCookie(w)
	h.cookies.UnsetRefreshCookie(w)

	h.response.Success(w, r, response.SuccessUserDeleted)
}
