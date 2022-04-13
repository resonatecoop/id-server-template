package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/routes"
)

// ServiceInterface defines exported methods
type ServiceInterface interface {
	// Exported methods
	GetConfig() *config.Config
	GetOauthService() oauth.ServiceInterface
	GetSessionService() session.ServiceInterface
	GetRoutes() []routes.Route
	RegisterRoutes(router *mux.Router, prefix string)
	Close()

	// Needed for the newRoutes to be able to register handlers
	setSessionService(r *http.Request, w http.ResponseWriter)
	authorizeForm(w http.ResponseWriter, r *http.Request)
	authorize(w http.ResponseWriter, r *http.Request)
	homeForm(w http.ResponseWriter, r *http.Request)
	passwordResetForm(w http.ResponseWriter, r *http.Request)
	passwordReset(w http.ResponseWriter, r *http.Request)
	getEmailConfirmationToken(w http.ResponseWriter, r *http.Request)
	resendEmailConfirmationToken(w http.ResponseWriter, r *http.Request)
	profileForm(w http.ResponseWriter, r *http.Request)
	accountForm(w http.ResponseWriter, r *http.Request)
	account(w http.ResponseWriter, r *http.Request)
	accountSettingsForm(w http.ResponseWriter, r *http.Request)
	accountSettings(w http.ResponseWriter, r *http.Request)
	membershipForm(w http.ResponseWriter, r *http.Request)
	membership(w http.ResponseWriter, r *http.Request)
	checkoutForm(w http.ResponseWriter, r *http.Request)
	checkoutCancel(w http.ResponseWriter, r *http.Request)
	checkoutSuccess(w http.ResponseWriter, r *http.Request)
	checkout(w http.ResponseWriter, r *http.Request)
	clientForm(w http.ResponseWriter, r *http.Request)
	client(w http.ResponseWriter, r *http.Request)
	clientDelete(w http.ResponseWriter, r *http.Request)
	loginForm(w http.ResponseWriter, r *http.Request)
	login(w http.ResponseWriter, r *http.Request)
	logout(w http.ResponseWriter, r *http.Request)
	joinForm(w http.ResponseWriter, r *http.Request)
	join(w http.ResponseWriter, r *http.Request)
}
