package response

import (
	"api/internal/platform/i18n"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Response struct {
	logger *slog.Logger
	i18n   *i18n.Translator
}

func New(logger *slog.Logger, i18n *i18n.Translator) *Response {
	return &Response{
		logger: logger,
		i18n:   i18n,
	}
}

/* Success response with AppSuccess message */
func (resp *Response) Success(w http.ResponseWriter, r *http.Request, message *AppSuccess) {
	lang := i18n.GetLanguageFromContext(r.Context())

	resp.logger.Info("Success", "method", r.Method, "path", r.URL.Path, "message", message.Slug, "lang", lang)

	resp.generalResponse(w, message.Code, map[string]any{
		"slug":    message.Slug,
		"message": resp.i18n.T(lang, message.Slug, ""),
	})
}

/* Success response with data */
func (resp *Response) SuccessWithData(w http.ResponseWriter, r *http.Request, message *AppSuccess, data any) {
	lang := i18n.GetLanguageFromContext(r.Context())

	resp.generalResponse(w, message.Code, map[string]any{
		"slug":    message.Slug,
		"message": resp.i18n.T(lang, message.Slug, ""),
		"data":    data,
	})
}

/* Success response with data only */
func (resp *Response) SuccessDataOnly(w http.ResponseWriter, r *http.Request, data any) {
	resp.generalResponse(w, http.StatusOK, data)
}

/* Error response */
func (resp *Response) Error(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError

	if !errors.As(err, &appErr) {
		// Not custom error, wrap to custom unknown error
		err = ErrUnknown.Wrap(err)
	}

	if errors.As(err, &appErr) {
		// Unwrap original error
		errOriginal := appErr.Unwrap()
		errOriginalText := "none"
		if errOriginal != nil {
			errOriginalText = errOriginal.Error()
		}

		// Fine tune errors
		if errOriginalText == "http: request body too large" {
			appErr = ErrRequestTooLarge
		}

		// User language
		lang := i18n.GetLanguageFromContext(r.Context())

		errLogText := resp.i18n.T("en", "err_"+appErr.Slug, "")
		errResponseText := resp.i18n.T(lang, "err_"+appErr.Slug, "")

		// Write logs
		resp.logger.Error("Error", "method", r.Method, "path", r.URL.Path, "error", errLogText, "original", errOriginalText)

		// Return error response
		resp.generalError(w, appErr.Code, map[string]any{
			"error": errResponseText,
			"slug":  appErr.Slug,
		})
		return
	}
}

/* General error helper function */
func (resp *Response) generalError(w http.ResponseWriter, code int, data any) {
	if data == nil {
		resp.jsonWrite(w, code, nil)
		return
	}

	switch v := data.(type) {
	case string:
		resp.jsonWrite(w, code, map[string]string{"error": v})
	default:
		resp.jsonWrite(w, code, v)
	}
}

/* General response helper function */
func (resp *Response) generalResponse(w http.ResponseWriter, code int, data any) {
	if data == nil {
		resp.jsonWrite(w, code, nil)
		return
	}

	switch v := data.(type) {
	case string:
		resp.jsonWrite(w, code, map[string]string{"message": v})
	default:
		resp.jsonWrite(w, code, v)
	}
}

/* Write JSON to ResponseWriter */
func (resp *Response) jsonWrite(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			resp.logger.Error("JSON encode error", "err", err)
		}
	}
}
