package web

import (
	"net/http"

	"github.com/gorilla/csrf"
)

func (s *Service) passwordUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
}
