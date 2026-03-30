package main

import (
	"net/http"
)

/* Simple health check handler */
func (app *app) healthHandler(w http.ResponseWriter, r *http.Request) {
	app.response.SuccessDataOnly(w, r, "healthy")
}
