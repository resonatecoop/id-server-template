package services

import (
	"reflect"

	"github.com/gorilla/sessions"
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/health"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/web"
	"github.com/uptrace/bun"
)

func init() {

}

var (
	// HealthService ...
	HealthService health.ServiceInterface

	// OauthService ...
	OauthService oauth.ServiceInterface

	// WebService ...
	WebService web.ServiceInterface

	// SessionService ...
	SessionService session.ServiceInterface
)

// UseHealthService sets the health service
func UseHealthService(h health.ServiceInterface) {
	HealthService = h
}

// UseOauthService sets the oAuth service
func UseOauthService(o oauth.ServiceInterface) {
	OauthService = o
}

// UseWebService sets the web service
func UseWebService(w web.ServiceInterface) {
	WebService = w
}

// UseSessionService sets the session service
func UseSessionService(s session.ServiceInterface) {
	SessionService = s
}

// Init starts up all services
func Init(cnf *config.Config, db *bun.DB) error {
	if nil == reflect.TypeOf(HealthService) {
		HealthService = health.NewService(db)
	}

	if nil == reflect.TypeOf(OauthService) {
		OauthService = oauth.NewService(cnf, db)
	}

	if nil == reflect.TypeOf(SessionService) {
		// note: default session store is CookieStore
		store := sessions.NewCookieStore([]byte(cnf.Session.Secret))

		store.Options = &sessions.Options{
			Path:     cnf.Session.Path,
			MaxAge:   cnf.Session.MaxAge,
			Secure:   cnf.Session.Secure,
			HttpOnly: cnf.Session.HTTPOnly,
		}

		SessionService = session.NewService(cnf, store)
	}

	if nil == reflect.TypeOf(WebService) {
		WebService = web.NewService(cnf, OauthService, SessionService)
	}

	return nil
}

// Close closes any open services
func Close() {
	HealthService.Close()
	OauthService.Close()
	WebService.Close()
	SessionService.Close()
}
