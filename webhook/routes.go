package webhook

import (
	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/util/routes"
)

// RegisterRoutes registers route handlers for the health service
func (s *Service) RegisterRoutes(router *mux.Router, prefix string) {
	subRouter := router.PathPrefix(prefix).Subrouter()
	routes.AddRoutes(s.GetRoutes(), subRouter)
}

// GetRoutes returns []routes.Route slice for the health service
func (s *Service) GetRoutes() []routes.Route {
	return []routes.Route{
		{
			Name:        "stripe_payment",
			Method:      "POST",
			Pattern:     "/payment",
			HandlerFunc: s.stripePayment,
		},
	}
}
