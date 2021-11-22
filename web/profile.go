package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) profileForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	isUserAccountComplete := s.oauthService.IsUserAccountComplete(user)

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

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	q := gountries.New()
	countries := q.FindAllCountries()

	usergroup := r.URL.Query().Get("usergroup")

	if usergroup == "" {
		usergroup = s.getDefaultUserGroupType(user) // artist, label or user
	}

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		usergroup,
		isUserAccountComplete,
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

	// default template
	templateName := "profile.html"

	switch usergroup {
	case "artist":
		templateName = "profile_artist.html"
	case "label":
		templateName = "profile_label.html"
	}

	profile := &Profile{
		Email:          user.Username,
		FullName:       user.FullName,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
		Complete:       isUserAccountComplete,
	}

	err = renderTemplate(w, templateName, map[string]interface{}{
		"isUserAccountComplete": isUserAccountComplete,
		"flash":                 flash,
		"clientID":              client.Key,
		"countries":             countries,
		"applicationName":       client.ApplicationName.String,
		"profile":               profile,
		"queryString":           getQueryString(query),
		"initialState":          template.HTML(fragment),
		csrf.TemplateTag:        csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) getDefaultUserGroupType(user *model.User) string {
	var usergroup = "user"

	switch (model.AccessRole)(user.RoleID) {
	case model.ArtistRole:
		usergroup = "artist"
	case model.LabelRole:
		usergroup = "label"
	}

	return usergroup
}

func (s *Service) profileCommon(r *http.Request) (
	session.ServiceInterface,
	*model.Client,
	*model.User,
	*session.UserSession,
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

	return sessionService, client, user, userSession, nil
}
