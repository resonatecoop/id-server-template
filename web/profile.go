package web

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/gorilla/csrf"
)

func (s *Service) profileForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, _, responseType, _, err := s.authorizeCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	errMsg, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('{"applicationName":"%s"}')</script>`,
		client.ApplicationName.String,
	)

	renderTemplate(w, "authorize.html", map[string]interface{}{
		"error":           errMsg,
		"clientID":        client.Key,
		"applicationName": client.ApplicationName.String,
		"queryString":     getQueryString(query),
		"token":           responseType == "token",
		"initialState":    template.HTML(fragment),
		csrf.TemplateTag:  csrf.TemplateField(r),
	})
}

func (s *Service) profile(w http.ResponseWriter, r *http.Request) {
	// _, client, user, responseType, redirectURI, err := s.authorizeCommon(r)
	// if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	// Get the state parameter
	//state := r.Form.Get("state")
}

func (s *Service) profileCommon(r *http.Request) (session.ServiceInterface, *models.OauthClient, *models.OauthUser, string, *url.URL, error) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Fetch the user
	user, _, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	// Check the response_type is either "code" or "token"
	responseType := r.Form.Get("response_type")
	if responseType != "code" && responseType != "token" {
		return nil, nil, nil, "", nil, ErrIncorrectResponseType
	}

	// Fallback to the client redirect URI if not in query string
	redirectURI := r.Form.Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = client.RedirectURI.String
	}

	// // Parse the redirect URL
	parsedRedirectURI, err := url.ParseRequestURI(redirectURI)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	return sessionService, client, user, responseType, parsedRedirectURI, nil
}
