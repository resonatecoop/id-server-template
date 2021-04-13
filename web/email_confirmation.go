package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/gorilla/csrf"
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

	err = s.emailConfirm(r)

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
		redirectWithQueryString("/web/profile", query, w, r)
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

	redirectWithQueryString("/web/profile", query, w, r)
}

func (s *Service) emailConfirm(r *http.Request) error {
	if r.Form.Get("token") == "" {
		return ErrTokenMissing
	}

	token := r.Form.Get("token")

	emailToken, email, err := s.oauthService.GetValidEmailToken(token)

	if err != nil {
		return err
	}

	// set email_confirmed to true
	err = s.oauthService.ConfirmUserEmail(email)

	if err != nil {
		return err
	}

	softDelete := true
	err = s.oauthService.DeleteEmailToken(emailToken, softDelete)

	if err != nil {
		return err
	}

	return nil
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

	email := models.NewOauthEmail(
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

	redirectWithQueryString("/web/profile", r.URL.Query(), w, r)
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
