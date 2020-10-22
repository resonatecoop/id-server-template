package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/RichardKnop/go-oauth2-server/util/response"
	"github.com/gorilla/csrf"
)

func (s *Service) profileForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, wpuser, nickname, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	profile := &Profile{
		ID:             wpuser.ID,
		Email:          wpuser.Email,
		DisplayName:    nickname,
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

	switch r.Header.Get("Accept") {
	case "application/json":
		response.WriteJSON(w, profile, http.StatusCreated)
	default:
		renderTemplate(w, "profile.html", map[string]interface{}{
			"flash":           flash,
			"clientID":        client.Key,
			"applicationName": client.ApplicationName.String,
			"profile":         profile,
			"queryString":     getQueryString(query),
			"initialState":    template.HTML(fragment),
			csrf.TemplateTag:  csrf.TemplateField(r),
		})
	}
}

func (s *Service) profileUpdate(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, _, _, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	if s.oauthService.UpdateUsername(
		user,
		r.Form.Get("email"),
	); err != nil {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, err.Error(), http.StatusBadRequest)
		default:
			sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}
}

/**
 * Soft delete profiles by setting deleted_at to current date
 * Deletion will automatically happen after x days
 */

func (s *Service) profileDelete(w http.ResponseWriter, r *http.Request) {
	_, _, _, _, _, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// TODO
}

func (s *Service) profileCommon(r *http.Request) (
	session.ServiceInterface,
	*models.OauthClient,
	*models.OauthUser,
	*models.WpUser,
	string,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	// Fetch the wp user
	wpuser, err := s.oauthService.FindWpUserByEmail(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	nickname, err := s.oauthService.FindNicknameByWpUserID(wpuser.ID)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}

	return sessionService, client, user, wpuser, nickname, nil
}
