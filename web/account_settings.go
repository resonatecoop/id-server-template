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
)

func (s *Service) accountSettingsForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, credits, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	usergroups, _ := s.getUserGroupList(user, userSession.AccessToken)

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		isUserAccountComplete,
		credits,
		usergroups.Usergroup,
		nil,
		nil,
		nil,
		"",
		nil,
	))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	profile := NewProfile(user, usergroups.Usergroup, isUserAccountComplete, credits, userSession.Role)

	err = renderTemplate(w, "account_settings.html", map[string]interface{}{
		"appURL":                s.cnf.AppURL,
		"applicationName":       client.ApplicationName.String,
		"clientID":              client.Key,
		"flash":                 flash,
		"initialState":          template.HTML(fragment),
		"isUserAccountComplete": isUserAccountComplete,
		"profile":               profile,
		"queryString":           getQueryString(query),
		"staticURL":             s.cnf.StaticURL,
		csrf.TemplateTag:        csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) accountSettings(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, isUserAccountComplete, _, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isUserAccountComplete {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: "Account not complete",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	method := strings.ToLower(r.Form.Get("_method"))

	message := "Account not updated"

	if method == "delete" || r.Method == http.MethodDelete {
		if err = s.oauthService.DeleteUser(
			user,
			r.Form.Get("password"),
		); err != nil {
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

		// Delete the access and refresh tokens
		s.oauthService.ClearUserTokens(userSession)

		// Delete the user session
		err = sessionService.ClearUserSession()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		message = "Your account is now scheduled for deletion"
	}

	if method == "put" || r.Method == http.MethodPut {
		// update email, requires password, sends notification
		if r.Form.Get("email") != "" && r.Form.Get("email") != user.Username {
			if err = s.oauthService.UpdateUsername(
				user,
				r.Form.Get("email"),
				r.Form.Get("password"),
			); err != nil {
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

			// Delete the access and refresh tokens
			s.oauthService.ClearUserTokens(userSession)

			// Delete the user session
			err = sessionService.ClearUserSession()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			message = "Email updated"
		}
	}

	redirectURI := "/web/account-settings"

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"data": map[string]interface{}{
				"success_redirect_url": redirectURI,
				"profile_redirection":  redirectURI == "/web/profile",
				"account_complete":     isUserAccountComplete,
			},
			"status": http.StatusOK,
		}, http.StatusOK)
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
	query := r.URL.Query()

	redirectWithQueryString(redirectURI, query, w, r)
	return
}
