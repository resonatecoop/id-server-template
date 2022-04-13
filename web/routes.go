package web

import (
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth_negroni"
	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/util/routes"
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
			Method:      "POST",
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
			Name:        "password_update",
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
			Name:        "account_form",
			Method:      "GET",
			Pattern:     "/account",
			HandlerFunc: s.accountForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "account",
			Method:      "POST",
			Pattern:     "/account",
			HandlerFunc: s.account,
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
			Name:        "account_settings_form",
			Method:      "GET",
			Pattern:     "/account-settings",
			HandlerFunc: s.accountSettingsForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "account_settings",
			Method:      "POST",
			Pattern:     "/account-settings",
			HandlerFunc: s.accountSettings,
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
			Name:        "account_settings_update",
			Method:      "PUT",
			Pattern:     "/account-settings",
			HandlerFunc: s.accountSettings,
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
			Name:        "account_settings_delete",
			Method:      "DELETE",
			Pattern:     "/account-settings",
			HandlerFunc: s.accountSettings,
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
			Name:        "membership_form",
			Method:      "GET",
			Pattern:     "/membership",
			HandlerFunc: s.membershipForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "cancel_subscription",
			Method:      "POST",
			Pattern:     "/membership",
			HandlerFunc: s.membership,
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
			Name:        "account_update",
			Method:      "PUT",
			Pattern:     "/account",
			HandlerFunc: s.account,
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
			Name:        "account_delete",
			Method:      "DELETE",
			Pattern:     "/account",
			HandlerFunc: s.account,
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
			Name:        "checkout_form",
			Method:      "GET",
			Pattern:     "/checkout",
			HandlerFunc: s.checkoutForm,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "checkout_success",
			Method:      "GET",
			Pattern:     "/checkout/success",
			HandlerFunc: s.checkoutSuccess,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "checkout_cancel_form",
			Method:      "GET",
			Pattern:     "/checkout/cancel",
			HandlerFunc: s.checkoutCancel,
			Middlewares: []negroni.Handler{
				new(parseFormMiddleware),
				newLoggedInMiddleware(s),
				newClientMiddleware(s),
			},
		},
		{
			Name:        "checkout",
			Method:      "POST",
			Pattern:     "/checkout",
			HandlerFunc: s.checkout,
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
				newClientMiddleware(s),
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
