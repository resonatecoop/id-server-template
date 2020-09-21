package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

func (s *Service) passwordResetForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	initialState, _ := json.Marshal(map[string]interface{}{
		"clients": s.cnf.Clients,
	})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	renderTemplate(w, "password_reset.html", map[string]interface{}{
		"clients":        s.cnf.Clients,
		"initialState":   template.HTML(fragment),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func (s *Service) passwordReset(w http.ResponseWriter, r *http.Request) {

}
