package admin

import (
	"api/internal/domain/tokens"
	"api/internal/platform/response"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	admin    *Service
	response *response.Response
}

func NewHandler(service *Service, resp *response.Response) *Handler {
	return &Handler{
		admin:    service,
		response: resp,
	}
}

/* Hard delete user (with all albums via db onDelete) */
func (h *Handler) PurgeUser(w http.ResponseWriter, r *http.Request) {
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

	// Check if admin is not deleting himself
	if claims.UserID == deletingUserID {
		h.response.Error(w, r, response.ErrNoPermission)
		return
	}

	// Delete user
	_, err = h.admin.PurgeUser(r.Context(), deletingUserID)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	h.response.Success(w, r, response.SuccessUserDeleted)
}

/* Restore deleted user */
func (h *Handler) RestoreUser(w http.ResponseWriter, r *http.Request) {

	// User id from url
	idStr := chi.URLParam(r, "uuid")
	restoringUserID, err := uuid.Parse(idStr)
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

	// Check if admin is not restoring himself
	if claims.UserID == restoringUserID {
		h.response.Error(w, r, response.ErrNoPermission)
		return
	}

	// Restoring user
	_, _, err = h.admin.RestoreUser(r.Context(), restoringUserID)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	h.response.Success(w, r, response.SuccessUserRestored)
}
