package endpoints

import "net/http"

type HealthResponse struct{}

func (e *endpoints) GetHealth(w http.ResponseWriter, r *http.Request)
