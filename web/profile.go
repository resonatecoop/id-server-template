package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/RichardKnop/go-oauth2-server/util/response"
	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"
)

func (s *Service) profileForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, wpuser, nickname, country, role, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	q := gountries.New()
	countries := q.FindAllCountries()

	gountry, _ := q.FindCountryByName(strings.ToLower(country))

	profile := &Profile{
		ID:             wpuser.ID,
		Email:          wpuser.Email,
		DisplayName:    nickname,
		Country:        gountry.Codes.Alpha2,
		Role:           role,
		EmailConfirmed: user.EmailConfirmed,
	}

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		profile,
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

	err = renderTemplate(w, "profile.html", map[string]interface{}{
		"flash":           flash,
		"clientID":        client.Key,
		"countries":       countries,
		"applicationName": client.ApplicationName.String,
		"profile":         profile,
		"queryString":     getQueryString(query),
		"initialState":    template.HTML(fragment),
		csrf.TemplateTag:  csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) profile(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, wpuser, _, _, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	method := strings.ToLower(r.Form.Get("_method"))

	message := "Profile not updated"

	if method == "delete" || r.Method == http.MethodDelete {
		if s.oauthService.DeleteUser(
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

		message = "Account is now scheduled for deletion"
	}

	if method == "put" || r.Method == http.MethodPut {
		if r.Form.Get("role") != "" {
			// update wpuser role
			if err = s.oauthService.UpdateWpUserRole(
				wpuser,
				r.Form.Get("role"),
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
		}

		// username is always email
		if r.Form.Get("email") != "" {
			if s.oauthService.UpdateUsername(
				user,
				r.Form.Get("email"),
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
		}

		if r.Form.Get("nickname") != "" {
			// update wpuser nickname
			if s.oauthService.UpdateWpUserMetaValue(
				wpuser.ID,
				"nickname",
				r.Form.Get("nickname"),
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
		}

		if r.Form.Get("country") != "" {
			// update wpuser country
			if s.oauthService.UpdateWpUserCountry(
				wpuser,
				r.Form.Get("country"),
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
		}

		message = "Profile updated"
	}

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"status":  http.StatusOK,
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
	http.Redirect(w, r, r.RequestURI, http.StatusFound)
}

func (s *Service) profileCommon(r *http.Request) (
	session.ServiceInterface,
	*models.OauthClient,
	*models.OauthUser,
	*models.WpUser,
	string,
	string,
	string,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	// Fetch the wp user
	wpuser, err := s.oauthService.FindWpUserByEmail(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	nickname, err := s.oauthService.FindWpUserMetaValue(wpuser.ID, "nickname")
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	country, err := s.oauthService.FindWpUserMetaValue(wpuser.ID, "country")
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	role, err := s.oauthService.FindWpUserMetaValue(wpuser.ID, "role")
	if err != nil {
		return nil, nil, nil, nil, "", "", "", err
	}

	return sessionService, client, user, wpuser, nickname, country, role, nil
}
