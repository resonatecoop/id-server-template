package web

import (
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
)

// parseFormMiddleware parses the form so r.Form becomes available
type parseFormMiddleware struct{}

// ServeHTTP as per the negroni.Handler interface
func (m *parseFormMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	next(w, r)
}

// skipCSRFMiddleware just initialises session
type skipCSRFMiddleware struct {
	service ServiceInterface
}

// newSkipCSRFMiddleware creates a new skipCSRFMiddleware instance
func newSkipCSRFMiddleware(service ServiceInterface) *skipCSRFMiddleware {
	return &skipCSRFMiddleware{service: service}
}

// ServeHTTP as per the negroni.Handler interface
func (m *skipCSRFMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	r = csrf.UnsafeSkipCheck(r)

	next(w, r)
}

// guestMiddleware just initialises session
type guestMiddleware struct {
	service ServiceInterface
}

// newGuestMiddleware creates a new guestMiddleware instance
func newGuestMiddleware(service ServiceInterface) *guestMiddleware {
	return &guestMiddleware{service: service}
}

// ServeHTTP as per the negroni.Handler interface
func (m *guestMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Initialise the session service
	m.service.setSessionService(r, w)
	sessionService := m.service.GetSessionService()

	// Attempt to start the session
	if err := sessionService.StartSession(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	context.Set(r, sessionServiceKey, sessionService)

	// Try to get a user session
	_, err := sessionService.GetUserSession()
	if err == nil {
		query := r.URL.Query()
		query.Set("login_redirect_uri", r.URL.Path)
		sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: "You are already logged in",
		})
		redirectWithQueryString("/web/profile", query, w, r)
		return
	}

	next(w, r)
}

// loggedInMiddleware initialises session and makes sure the user is logged in
type loggedInMiddleware struct {
	service ServiceInterface
}

// newLoggedInMiddleware creates a new loggedInMiddleware instance
func newLoggedInMiddleware(service ServiceInterface) *loggedInMiddleware {
	return &loggedInMiddleware{service: service}
}

// ServeHTTP as per the negroni.Handler interface
func (m *loggedInMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Initialise the session service
	m.service.setSessionService(r, w)
	sessionService := m.service.GetSessionService()

	// Attempt to start the session
	if err := sessionService.StartSession(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	context.Set(r, sessionServiceKey, sessionService)

	// Try to get a user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		query := r.URL.Query()
		query.Set("login_redirect_uri", r.URL.Path)
		redirectWithQueryString("/web/login", query, w, r)
		return
	}

	// Authenticate
	if err := m.authenticate(userSession); err != nil {
		// Delete the user session
		err = sessionService.ClearUserSession()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete the checkout session
		err = sessionService.ClearCheckoutSession()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query := r.URL.Query()
		query.Set("login_redirect_uri", r.URL.Path)
		redirectWithQueryString("/web/login", query, w, r)
		return
	}

	// Update the user session
	err = sessionService.SetUserSession(userSession)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	next(w, r)
}

func (m *loggedInMiddleware) authenticate(userSession *session.UserSession) error {
	// Try to authenticate with the stored access token
	_, err := m.service.GetOauthService().Authenticate(userSession.AccessToken)
	if err == nil {
		// Access token valid, return
		return nil
	}
	// Access token might be expired, let's try refreshing...

	// Fetch the client
	client, err := m.service.GetOauthService().FindClientByClientID(
		userSession.ClientID, // client ID
	)
	if err != nil {
		return err
	}

	// Validate the refresh token
	theRefreshToken, err := m.service.GetOauthService().GetValidRefreshToken(
		userSession.RefreshToken, // refresh token
		client,                   // client
	)
	if err != nil {
		return err
	}

	// Log in the user
	accessToken, refreshToken, err := m.service.GetOauthService().Login(
		theRefreshToken.Client,
		theRefreshToken.User,
		theRefreshToken.Scope,
	)
	if err != nil {
		return err
	}

	scopes := strings.Split(accessToken.Scope, " ")

	userSession.Role = scopes[1] // user, artist, label, admin, tenantadmin, ...
	userSession.AccessToken = accessToken.Token
	userSession.RefreshToken = refreshToken.Token

	return nil
}

// clientMiddleware takes client_id param from the query string and
// makes a database lookup for a client with the same client ID
type clientMiddleware struct {
	service ServiceInterface
}

// newClientMiddleware creates a new clientMiddleware instance
func newClientMiddleware(service ServiceInterface) *clientMiddleware {
	return &clientMiddleware{service: service}
}

// ServeHTTP as per the negroni.Handler interface
func (m *clientMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cnf := m.service.GetConfig()
	redirect := cnf.ApplicationURL // get default application URL

	if r.Form.Get("redirect") != "" {
		redirect = r.Form.Get("redirect")
	}

	if r.Form.Get("client_id") != "" {
		// Fetch the client
		client, err := m.service.GetOauthService().FindClientByClientID(
			r.Form.Get("client_id"), // client ID
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, clientKey, client)
	} else {
		// fallback to default application uri
		client, err := m.service.GetOauthService().FindClientByApplicationURL(
			redirect,
		)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, clientKey, client)
	}

	next(w, r)
}
