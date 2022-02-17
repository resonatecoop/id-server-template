package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"
)

var (
	ErrTokenMissing = errors.New("Email confirmation token is missing")
)

func (s *Service) getEmailConfirmationToken(w http.ResponseWriter, r *http.Request) {
	sessionService, err := s.emailConfirmationCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	user, err := s.emailConfirm(r)

	query := r.URL.Query()
	query.Del("token")

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the scope string
	scope, err := s.oauthService.GetScope("read_write")
	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	// Log in the user
	accessToken, refreshToken, err := s.oauthService.Login(
		client,
		user,
		scope,
	)
	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	// Log in the user and store the user session in a cookie
	userSession := &session.UserSession{
		ClientID:     client.Key,
		Username:     user.Username,
		Role:         strings.Split(accessToken.Scope, " ")[1],
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	}
	if err := sessionService.SetUserSession(userSession); err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "Thank your for confirming your email",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectWithQueryString("/web/account", query, w, r)
}

func (s *Service) emailConfirm(r *http.Request) (*model.User, error) {
	token := r.URL.Query().Get("token")

	if token == "" {
		return nil, ErrTokenMissing
	}

	emailToken, user, err := s.oauthService.GetValidEmailToken(token)

	if err != nil {
		return nil, err
	}

	// set email_confirmed to true
	err = s.oauthService.ConfirmUserEmail(user.Username)

	if err != nil {
		return nil, err
	}

	softDelete := true
	err = s.oauthService.DeleteEmailToken(emailToken, softDelete)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) resendEmailConfirmationToken(w http.ResponseWriter, r *http.Request) {
	sessionService, err := s.emailConfirmationCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the client from the request context
	_, err = getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.EmailConfirmed {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: "Email is already confirmed",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	email := model.NewOauthEmail(
		user.Username,
		"Confirm your email",
		"email-confirmation",
	)
	_, err = s.oauthService.SendEmailToken(
		email,
		fmt.Sprintf(
			"https://%s/email-confirmation",
			s.cnf.Hostname,
		),
	)

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "An email is on its way",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectWithQueryString("/web/account-settings", r.URL.Query(), w, r)
}

func (s *Service) emailConfirmationCommon(r *http.Request) (
	session.ServiceInterface,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, err
	}

	return sessionService, nil
}
