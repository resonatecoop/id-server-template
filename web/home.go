package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

func (s *Service) homeForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	initialState, _ := json.Marshal(map[string]interface{}{
		"clients": s.cnf.Clients,
	})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	err := renderTemplate(w, "home.html", map[string]interface{}{
		"appURL":         s.cnf.AppURL,
		"clients":        s.cnf.Clients,
		"initialState":   template.HTML(fragment),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
