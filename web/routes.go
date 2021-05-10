package web

import (
	"github.com/RichardKnop/go-oauth2-server/util/routes"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_negroni"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
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
			Name:        "home",
			Method:      "GET",
			Pattern:     "/",
			HandlerFunc: s.homeForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newGuestMiddleware(s),
			},
		},
		{
			Name:        "join_form",
			Method:      "GET",
			Pattern:     "/join",
			HandlerFunc: s.joinForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newGuestMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "join",
			Method:      "POST",
			Pattern:     "/join",
			HandlerFunc: s.join,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newGuestMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "login_form",
			Method:      "GET",
			Pattern:     "/login",
			HandlerFunc: s.loginForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newGuestMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "login",
			Method:      "POST",
			Pattern:     "/login",
			HandlerFunc: s.login,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newGuestMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "logout",
			Method:      "GET",
			Pattern:     "/logout",
			HandlerFunc: s.logout,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
			},
		},
		{
			Name:        "authorize_form",
			Method:      "GET",
			Pattern:     "/authorize",
			HandlerFunc: s.authorizeForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "authorize",
			Method:      "POST",
			Pattern:     "/authorize",
			HandlerFunc: s.authorize,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "password_reset_form",
			Method:      "GET",
			Pattern:     "/password-reset",
			HandlerFunc: s.passwordResetForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newGuestMiddleware(s),
			},
		},
		{
			Name:        "password_reset",
			Method:      "POST",
			Pattern:     "/password-reset",
			HandlerFunc: s.passwordReset,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newGuestMiddleware(s),
			},
		},
		{
			Name:        "password_reset_update_password",
			Method:      "PUT",
			Pattern:     "/password-reset",
			HandlerFunc: s.passwordReset,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newGuestMiddleware(s),
			},
		},
		{
			Name:        "password",
			Method:      "PUT",
			Pattern:     "/password",
			HandlerFunc: s.passwordUpdate,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "profile_form",
			Method:      "GET",
			Pattern:     "/profile",
			HandlerFunc: s.profileForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "profile",
			Method:      "POST",
			Pattern:     "/profile",
			HandlerFunc: s.profile,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "profile_update",
			Method:      "PUT",
			Pattern:     "/profile",
			HandlerFunc: s.profile,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "profile_delete",
			Method:      "POST",
			Pattern:     "/profile/delete",
			HandlerFunc: s.profile,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "client_form",
			Method:      "GET",
			Pattern:     "/apps",
			HandlerFunc: s.clientForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "client",
			Method:      "POST",
			Pattern:     "/apps",
			HandlerFunc: s.client,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "client_delete",
			Method:      "DELETE",
			Pattern:     "/apps",
			HandlerFunc: s.clientDelete,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "get_email_confirmation_token",
			Method:      "GET",
			Pattern:     "/email-confirmation",
			HandlerFunc: s.getEmailConfirmationToken,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newGuestMiddleware(s),
			},
		},
		{
			Name:        "resend_email_confirmation_token",
			Method:      "GET",
			Pattern:     "/resend-email-confirmation",
			HandlerFunc: s.resendEmailConfirmationToken,
			Middlewares: []negroni.Handler{
				tollbooth_negroni.LimitHandler(
					tollbooth.NewLimiter(1, nil),
				),
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
	}
}
