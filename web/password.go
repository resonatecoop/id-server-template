package web

import (
	"errors"
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/oauth"
	"github.com/RichardKnop/go-oauth2-server/session"
	pass "github.com/RichardKnop/go-oauth2-server/util/password"
	"github.com/RichardKnop/go-oauth2-server/util/response"
	"github.com/gorilla/csrf"
)

var (
	ErrPasswordMismatch = errors.New("Password confirmation mismatch")
)

func (s *Service) passwordUpdate(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, _, err := s.passwordCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	err = s.oauthService.SetPassword(user, r.Form.Get("password_new"))

	if err != nil {
		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		http.Redirect(w, r, r.RequestURI, http.StatusBadRequest)
		return
	}

	// Check that the password is set
	if !user.Password.Valid {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message := "We have sent you a password reset link to your e-mail. Please check your inbox"

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"status":  http.StatusAccepted,
		}, http.StatusAccepted)
		return
	}

	sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: message,
	})
	http.Redirect(w, r, r.RequestURI, http.StatusAccepted)
}

func (s *Service) passwordCommon(r *http.Request) (
	session.ServiceInterface,
	*models.OauthClient,
	*models.OauthUser,
	*models.WpUser,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	wpuser, err := s.oauthService.FindWpUserByEmail(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if pass.VerifyPassword(user.Password.String, r.Form.Get("password")) != nil {
		return nil, nil, nil, nil, oauth.ErrInvalidUserPassword
	}

	if r.Form.Get("password_new") != r.Form.Get("password_confirm") {
		return nil, nil, nil, nil, ErrPasswordMismatch
	}

	return sessionService, client, user, wpuser, nil
}
