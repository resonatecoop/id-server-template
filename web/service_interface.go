package web

import (
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/config"
	"github.com/RichardKnop/go-oauth2-server/oauth"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/RichardKnop/go-oauth2-server/util/routes"
	"github.com/gorilla/mux"
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
	profile(w http.ResponseWriter, r *http.Request)
	clientForm(w http.ResponseWriter, r *http.Request)
	client(w http.ResponseWriter, r *http.Request)
	clientDelete(w http.ResponseWriter, r *http.Request)
	loginForm(w http.ResponseWriter, r *http.Request)
	login(w http.ResponseWriter, r *http.Request)
	logout(w http.ResponseWriter, r *http.Request)
	joinForm(w http.ResponseWriter, r *http.Request)
	join(w http.ResponseWriter, r *http.Request)
}
