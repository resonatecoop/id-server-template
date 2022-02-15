package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api-client/models"
	"github.com/resonatecoop/user-api/model"
	"github.com/rs/xid"
	"github.com/thanhpk/randstr"
)

func (s *Service) clientForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, err := s.clientCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		nil,
		false,
		"",
		[]*models.UserUserGroupPrivateResponse{},
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

	profile := &Profile{
		EmailConfirmed: user.EmailConfirmed,
	}

	err = renderTemplate(w, "client.html", map[string]interface{}{
		"applicationName": client.ApplicationName.String,
		"clientID":        client.Key,
		"flash":           flash,
		"initialState":    template.HTML(fragment),
		"profile":         profile,
		"queryString":     getQueryString(query),
		"staticURL":       s.cnf.StaticURL,
		csrf.TemplateTag:  csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) client(w http.ResponseWriter, r *http.Request) {
	sessionService, _, _, err := s.clientCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	guid := xid.New()

	secret := randstr.Hex(16)

	// Create a new client
	client, err := s.oauthService.CreateClient(
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
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: "New client created",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/apps", r.URL.Query(), w, r)
	}
}

func (s *Service) clientDelete(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := s.clientCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// TODO
}

func (s *Service) clientCommon(r *http.Request) (
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

	// Fetch the wpuser
	// wpuser, err := s.oauthService.FindUserByEmail(
	// 	userSession.Username,
	// )
	// if err != nil {
	// 	return nil, nil, nil, err
	// }

	// nickname, err := s.oauthService.FindNicknameByWpUserID(wpuser.ID)
	// if err != nil {
	// 	return nil, nil, nil, err
	// }

	return sessionService, client, user, nil
}
