package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) passwordResetForm(w http.ResponseWriter, r *http.Request) {
	sessionService, err := s.passwordResetCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	initialState, _ := json.Marshal(map[string]interface{}{
		"clients": s.cnf.Clients,
	})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	layoutTemplate := "password_reset.html"
	token := r.Form.Get("token")

	if token != "" {
		_, _, err = s.oauthService.GetValidEmailToken(
			token,
		)
		// TODO renew if close to expiration time ?
		if err != nil {
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			query := r.URL.Query()
			query.Del("token")
			redirectWithQueryString("/web/password-reset", query, w, r)
			return
		}
		layoutTemplate = "password_reset_update_password.html"
	}

	flash, _ := sessionService.GetFlashMessage()

	err = renderTemplate(w, layoutTemplate, map[string]interface{}{
		"token":          token,
		"flash":          flash,
		"clients":        s.cnf.Clients,
		"initialState":   template.HTML(fragment),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) passwordReset(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.ToLower(r.Form.Get("_method")) == "put" || r.Method == http.MethodPut {
		err = s.passwordResetUpdatePassword(r)

		if err != nil {
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
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
			return
		}

		message := "Your password was updated successfully. A confirmation email has been sent."

		if r.Header.Get("Accept") == "application/json" {
			response.WriteJSON(w, map[string]interface{}{
				"message": message,
			}, http.StatusAccepted)
			return
		}

		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: message,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/login", r.URL.Query(), w, r)
		return
	}

	// send password reset token
	_, err = s.oauthService.SendEmailToken(
		model.NewOauthEmail(
			r.Form.Get("email"),
			"Reset your password",
			"password-reset",
		),
		fmt.Sprintf(
			"https://%s/password-reset",
			s.cnf.Hostname,
		),
	)

	if err != nil {
		status := http.StatusBadRequest

		switch err {
		case oauth.ErrUsernameRequired:
		case oauth.ErrEmailNotFound:
			status = http.StatusNotFound
		default:
			status = http.StatusInternalServerError // assume email could not be sent
		}

		if r.Header.Get("Accept") == "application/json" {
			response.Error(w, err.Error(), status)
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
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
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

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: message,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, r.RequestURI, http.StatusFound)
	return
}

func (s *Service) passwordResetUpdatePassword(r *http.Request) error {
	emailToken, user, err := s.oauthService.GetValidEmailToken(r.Form.Get("token"))

	if err != nil {
		return err
	}

	if r.Form.Get("password_new") != r.Form.Get("password_confirm") {
		return ErrPasswordMismatch
	}

	err = s.oauthService.SetPassword(user, r.Form.Get("password_new"))

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

func (s *Service) passwordResetCommon(r *http.Request) (
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
