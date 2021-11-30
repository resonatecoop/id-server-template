package session

import "net/http"

// ServiceInterface defines exported methods
type ServiceInterface interface {
	SetSessionService(r *http.Request, w http.ResponseWriter)
	StartSession() error
	GetUserSession() (*UserSession, error)
	SetUserSession(userSession *UserSession) error
	GetCheckoutSession() (*CheckoutSession, error)
	SetCheckoutSession(userSession *CheckoutSession) error
	ClearCheckoutSession() error
	ClearUserSession() error
	SetFlashMessage(flash *Flash) error
	GetFlashMessage() (interface{}, error)
	Close()
}
