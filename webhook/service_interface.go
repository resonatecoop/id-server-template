package webhook

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/util/routes"
)

// ServiceInterface defines exported methods
type ServiceInterface interface {
	GetConfig() *config.Config
	GetOauthService() oauth.ServiceInterface
	GetRoutes() []routes.Route
	RegisterRoutes(router *mux.Router, prefix string)
	Close()

	stripePayment(w http.ResponseWriter, r *http.Request)
}
