package health

import (
	"net/http"

	"github.com/resonatecoop/id/util/response"
)

// Handles health check requests (GET /v1/health)
func (s *Service) healthcheck(w http.ResponseWriter, r *http.Request) {
	_, err := s.db.Exec("SELECT 1=1")
	//	defer rows.Close()

	var healthy bool
	if err == nil {
		healthy = true
	}

	response.WriteJSON(w, map[string]interface{}{
		"healthy": healthy,
	}, 200)
}
