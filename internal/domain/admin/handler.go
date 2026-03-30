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

	// Delete user
	_, err = h.admin.PurgeUser(r.Context(), claims.UserID, claims.Role, deletingUserID)
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

	// Restoring user
	_, _, err = h.admin.RestoreUser(r.Context(), claims.UserID, claims.Role, restoringUserID)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}

	h.response.Success(w, r, response.SuccessUserRestored)
}
