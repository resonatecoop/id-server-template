package web

import (
	"net/http"
)

func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the access and refresh tokens
	s.oauthService.ClearUserTokens(userSession)

	// Delete the user session
	err = sessionService.ClearUserSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the checkout session
	err = sessionService.ClearCheckoutSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the login page
	redirectWithQueryString("/web/login", r.URL.Query(), w, r)
}
