package web

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/session"
	pass "github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

var (
	ErrPasswordMismatch = errors.New("Password confirmation mismatch")
	ErrInvalidPassword  = errors.New("Invalid password")
)

func (s *Service) passwordUpdate(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, err := s.passwordCommon(r)

	if err != nil {
		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// verify current password
	if pass.VerifyPassword(user.Password.String, r.Form.Get("password")) != nil {
		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, ErrInvalidPassword.Error(), http.StatusBadRequest)
			return
		}
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Invalid password",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account-settings", r.URL.Query(), w, r)
		return
	}

	// compare new password and password confirmation
	if r.Form.Get("password_new") != r.Form.Get("password_confirm") {
		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, ErrPasswordMismatch.Error(), http.StatusBadRequest)
			return
		}
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: ErrPasswordMismatch.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account-settings", r.URL.Query(), w, r)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// set new password
	if s.oauthService.SetPassword(user, r.Form.Get("password_new")); err != nil {
		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account-settings", r.URL.Query(), w, r)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": "Your password has been successfully changed",
			"status":  http.StatusOK,
		}, http.StatusOK)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "Your password has been successfully changed",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Inform user by email password was changed
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email := model.NewOauthEmail(
		user.Username,
		"Password changed",
		"password-changed",
	)
	subject := email.Subject
	recipient := email.Recipient
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(email.Template) // set mailgun template
	err = message.AddTemplateVariable("email", recipient)

	if err != nil {
		log.ERROR.Print(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err = mg.Send(ctx, message)

	if err != nil {
		log.ERROR.Print(err)
	}

	redirectWithQueryString("/web/account-settings", r.URL.Query(), w, r)
}

func (s *Service) passwordCommon(r *http.Request) (
	session.ServiceInterface,
	*model.Client,
	*model.User,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return sessionService, client, user, nil
}
