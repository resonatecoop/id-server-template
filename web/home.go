package web

import (
	"net/http"

	"github.com/gorilla/csrf"
)

func (s *Service) homeForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	renderTemplate(w, "home.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}
