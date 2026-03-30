package request

import (
	"api/internal/platform/response"
	"encoding/json"
	"io"
	"net/http"
)

/* Decodes request JSON */
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, data any) error {
	// Create decoder
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// Main object decoding
	if err := decoder.Decode(data); err != nil {
		return response.ErrBadJSON
	}

	// Test on JSON Smuggling
	err := decoder.Decode(&struct{}{})
	if err != io.EOF {
		return response.ErrBadJSON
	}
	return nil
}
