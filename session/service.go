package session

import (
	"encoding/gob"
	"errors"
	"net/http"

	//"github.com/resonatecoop/id/config"
	"github.com/gorilla/sessions"
	"github.com/resonatecoop/id/config"
)

type Level string

const (
	LogLevelUnspecified Level = "Unspecified"
	LogLevelTrace             = "Trace"
	LogLevelInfo              = "Info"
	LogLevelWarning           = "Warning"
	LogLevelError             = "Error"
)

type Flash struct {
	Type    Level  `json:"type"`
	Message string `json:"message"`
}

// Service wraps session functionality
type Service struct {
	sessionStore   sessions.Store
	sessionOptions *sessions.Options
	session        *sessions.Session
	r              *http.Request
	w              http.ResponseWriter
}

// CheckoutSession has stripe checkout data
type CheckoutSession struct {
	ID       string           // stripe checkout session id
	Products []config.Product // stripe products
}

// UserSession has user data stored in a session after logging in
type UserSession struct {
	ClientID               string
	Username               string
	Role                   string // user, artist, label, admin, tenantadmin, ...
	AccessToken            string
	RefreshToken           string
	CheckoutSessionID      string
	CheckoutSessionPriceID string
}

var (
	// StorageSessionName ...
	StorageSessionName = "go_oauth2_server_session"
	// UserSessionKey ...
	UserSessionKey = "go_oauth2_server_user"
	// CheckoutSessionKey ...
	CheckoutSessionKey = "go_oauth2_server_checkout"
	// ErrSessonNotStarted ...
	ErrSessonNotStarted = errors.New("Session not started")
)

func init() {
	gob.Register(new(Flash))
	// Register a new datatype for storage in sessions
	gob.Register(new(UserSession))
	gob.Register(new(CheckoutSession))
}

// NewService returns a new Service instance
func NewService(cnf *config.Config, sessionStore sessions.Store) *Service {
	return &Service{
		// Session cookie storage
		sessionStore: sessionStore,
		// Session options
		sessionOptions: &sessions.Options{
			Path:     cnf.Session.Path,
			MaxAge:   cnf.Session.MaxAge,
			Secure:   cnf.Session.Secure,
			HttpOnly: cnf.Session.HTTPOnly,
		},
	}
}

// SetSessionService sets the request and responseWriter on the session service
func (s *Service) SetSessionService(r *http.Request, w http.ResponseWriter) {
	s.r = r
	s.w = w
}

// StartSession starts a new session. This method must be called before other
// public methods of this struct as it sets the internal session object
func (s *Service) StartSession() error {
	session, err := s.sessionStore.Get(s.r, StorageSessionName)
	if err != nil {
		return err
	}
	s.session = session
	return nil
}

// GetUserSession returns the user session
func (s *Service) GetUserSession() (*UserSession, error) {
	// Make sure StartSession has been called
	if s.session == nil {
		return nil, ErrSessonNotStarted
	}

	// Retrieve our user session struct and type-assert it
	userSession, ok := s.session.Values[UserSessionKey].(*UserSession)
	if !ok {
		return nil, errors.New("User session type assertion error")
	}

	return userSession, nil
}

// SetUserSession saves the user session
func (s *Service) SetUserSession(userSession *UserSession) error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Set a new user session
	s.session.Values[UserSessionKey] = userSession
	return s.session.Save(s.r, s.w)
}

// ClearUserSession deletes the user session
func (s *Service) ClearUserSession() error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Delete the user session
	delete(s.session.Values, UserSessionKey)
	return s.session.Save(s.r, s.w)
}

// GetCheckoutSession returns the checkout session
func (s *Service) GetCheckoutSession() (*CheckoutSession, error) {
	// Make sure StartSession has been called
	if s.session == nil {
		return nil, ErrSessonNotStarted
	}

	// Retrieve our checkout session struct and type-assert it
	checkoutSession, ok := s.session.Values[CheckoutSessionKey].(*CheckoutSession)
	if !ok {
		return nil, errors.New("Checkout session type assertion error")
	}

	return checkoutSession, nil
}

// SetCheckoutSession saves the checkout session
func (s *Service) SetCheckoutSession(checkoutSession *CheckoutSession) error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Set a new checkout session
	s.session.Values[CheckoutSessionKey] = checkoutSession
	return s.session.Save(s.r, s.w)
}

// ClearCheckoutSession deletes the checkout session
func (s *Service) ClearCheckoutSession() error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Delete the checkout session
	delete(s.session.Values, CheckoutSessionKey)
	return s.session.Save(s.r, s.w)
}

// SetFlashMessage sets a flash message,
// useful for displaying an error after 302 redirection
func (s *Service) SetFlashMessage(flash *Flash) error {
	// Make sure StartSession has been called
	if s.session == nil {
		return ErrSessonNotStarted
	}

	// Add the flash message
	s.session.AddFlash(flash)
	return s.session.Save(s.r, s.w)
}

// GetFlashMessage returns the first flash message
func (s *Service) GetFlashMessage() (interface{}, error) {
	// Make sure StartSession has been called
	if s.session == nil {
		return nil, ErrSessonNotStarted
	}

	// Get the last flash message from the stack
	if flashes := s.session.Flashes(); len(flashes) > 0 {
		// We need to save the session, otherwise the flash message won't be removed
		err := s.session.Save(s.r, s.w)
		if err != nil {
			return nil, err
		}
		return flashes[0].(*Flash), nil
	}

	// No flash messages in the stack
	return nil, nil
}

// Close stops any running services
func (s *Service) Close() {}
