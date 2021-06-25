package web

import (
	"net/http"
)

func (s *Service) welcomeForm(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, _, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flash, _ := sessionService.GetFlashMessage()

	profile := &Profile{
		EmailConfirmed: user.EmailConfirmed,
	}

	err = renderTemplate(w, "welcome.html", map[string]interface{}{
		"profile":     profile,
		"flash":       flash,
		"queryString": getQueryString(r.URL.Query()),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
