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
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"github.com/thanhpk/randstr"
)

func (s *Service) clientForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, wpuser, nickname, err := s.clientCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	apps, err := s.oauthService.FindClientsByUserId(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	profile := &Profile{
		ID:             wpuser.ID,
		Email:          wpuser.Email,
		DisplayName:    nickname,
		EmailConfirmed: user.EmailConfirmed,
	}

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		apps,
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

	renderTemplate(w, "client.html", map[string]interface{}{
		"flash":           flash,
		"clientID":        client.Key,
		"apps":            apps,
		"applicationName": client.ApplicationName.String,
		"profile":         profile,
		"queryString":     getQueryString(query),
		"initialState":    template.HTML(fragment),
		csrf.TemplateTag:  csrf.TemplateField(r),
	})
}

func (s *Service) clientDeleteForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, wpuser, nickname, err := s.clientCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	params := mux.Vars(r)

	key := params["id"]

	app, err := s.oauthService.FindClientByClientID(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	profile := &Profile{
		ID:             wpuser.ID,
		Email:          wpuser.Email,
		DisplayName:    nickname,
		EmailConfirmed: user.EmailConfirmed,
	}

	renderTemplate(w, "client_delete.html", map[string]interface{}{
		"flash":          flash,
		"clientID":       client.Key,
		"app":            app,
		"profile":        profile,
		"queryString":    getQueryString(query),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func (s *Service) client(w http.ResponseWriter, r *http.Request) {
	sessionService, _, oauthUser, _, _, err := s.clientCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	guid := xid.New()

	secret := randstr.Hex(16)

	// Create a new client
	client, err := s.oauthService.CreateClient(
		oauthUser,
		guid.String(), // client id
		secret,        // client secret
		r.Form.Get("redirect_uri"),
		r.Form.Get("application_name"), // name or short description
		r.Form.Get("application_hostname"),
		r.Form.Get("application_url"),
	)

	if err != nil {
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

	switch r.Header.Get("Accept") {
	case "application/json":
		data := map[string]interface{}{
			"clientId":            client.Key,
			"secret":              secret,
			"applicationName":     client.ApplicationName,
			"applicationHostname": client.ApplicationHostname,
			"applicationURL":      client.ApplicationURL,
		}

		response.WriteJSON(w, map[string]interface{}{
			"data":   data,
			"status": http.StatusCreated,
		}, http.StatusCreated)
	default:
		sessionService.SetFlashMessage(&session.Flash{
			Type: "Info",
			Message: fmt.Sprintf(
				`New client created. Your secret is "%s". Make sure to store it in a safe place.`,
				secret,
			),
		})
		redirectWithQueryString("/web/apps", r.URL.Query(), w, r)
	}
}

func (s *Service) clientDelete(w http.ResponseWriter, r *http.Request) {
	sessionService, _, oauthUser, _, _, err := s.clientCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	if strings.ToLower(r.Form.Get("_method")) == "delete" || r.Method == http.MethodDelete {
		clientId := r.Form.Get("client_id")
		applicationName := r.Form.Get("application_name")

		client, err := s.oauthService.FindClientByClientID(clientId)

		if client.ApplicationName.String != applicationName {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.oauthService.DeleteClient(client.Key, oauthUser)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Header.Get("Accept") {
		case "application/json":
			response.WriteJSON(w, map[string]interface{}{
				"message": "Client deleted",
				"status":  http.StatusFound,
			}, http.StatusCreated)
		default:
			sessionService.SetFlashMessage(&session.Flash{
				Type:    "Info",
				Message: "Client deleted",
			})
			redirectWithQueryString("/web/apps", r.URL.Query(), w, r)
		}

		return
	}

	http.Error(w, err.Error(), http.StatusBadRequest)
	return
}

func (s *Service) clientCommon(r *http.Request) (
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

	// Fetch the wpuser
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
