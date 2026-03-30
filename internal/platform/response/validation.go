package response

import (
	"api/internal/platform/i18n"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ValidationResponse struct {
	Error   string            `json:"error"`
	Slug    string            `json:"slug"`
	Details map[string]string `json:"details,omitempty"`
}

func (resp *Response) ValidationError(w http.ResponseWriter, r *http.Request, err error) {
	lang := i18n.GetLanguageFromContext(r.Context())
	errorsMap := make(map[string]string)

	if vErrors, ok := err.(validator.ValidationErrors); ok {
		for _, f := range vErrors {
			vError := resp.i18n.T(lang, "validation_"+f.Tag(), f.Param())
			if vError == "" {
				vError = resp.i18n.T(lang, "validation_invalid", "")
			}
			errorsMap[f.Field()] = vError
		}
	}

	resp.generalError(w, http.StatusUnprocessableEntity, ValidationResponse{
		Error:   resp.i18n.T(lang, "validation_failed", ""),
		Slug:    "validation_failed",
		Details: errorsMap,
	})
}
