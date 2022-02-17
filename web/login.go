package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) loginForm(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
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

	flash, _ := sessionService.GetFlashMessage()

	err = renderTemplate(w, "login.html", map[string]interface{}{
		"appURL":         s.cnf.AppURL,
		"flash":          flash,
		"initialState":   template.HTML(fragment),
		"queryString":    getQueryString(r.URL.Query()),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate the user
	user, err := s.oauthService.AuthUser(
		r.Form.Get("email"),    // email/username
		r.Form.Get("password"), // password
	)

	if err != nil {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, err.Error(), http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}

	// Email should be confirmed (click autologin link in email)
	if !user.EmailConfirmed {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, "Please confirm your email", http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: "Please confirm your email",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}

		// resend email
		// TODO resend only if a last token has expired
		email := model.NewOauthEmail(
			user.Username,
			"Confirm your email",
			"email-confirmation",
		)
		_, _ = s.oauthService.SendEmailToken(
			email,
			fmt.Sprintf(
				"https://%s/email-confirmation",
				s.cnf.Hostname,
			),
		)

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

	scopes := strings.Split(accessToken.Scope, " ")

	// Log in the user and store the user session in a cookie
	userSession := &session.UserSession{
		ClientID:     client.Key,
		Username:     user.Username,
		Role:         scopes[1],
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

	// Redirect to the authorize page by default but allow redirection to other
	// pages by specifying a path with login_redirect_uri query string param
	loginRedirectURI := r.URL.Query().Get("login_redirect_uri")
	if loginRedirectURI == "" {
		loginRedirectURI = "/web/authorize"
	}
	redirectWithQueryString(loginRedirectURI, r.URL.Query(), w, r)
}
