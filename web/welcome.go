package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

func (s *Service) welcomeForm(w http.ResponseWriter, r *http.Request) {
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	initialState, _ := json.Marshal(map[string]interface{}{})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	flash, _ := sessionService.GetFlashMessage()
	err = renderTemplate(w, "welcome.html", map[string]interface{}{
		"flash":        flash,
		"initialState": template.HTML(fragment),
		"queryString":  getQueryString(r.URL.Query()),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
