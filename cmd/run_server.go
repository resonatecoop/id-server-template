package cmd

import (
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/services"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"
)

// RunServer runs the app
func RunServer(configBackend string) error {
	cnf, db, err := initConfigDB(true, true, configBackend)
	if err != nil {
		return err
	}
	defer db.Close()

	// start the services
	if err := services.Init(cnf, db); err != nil {
		return err
	}
	defer services.Close()

	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          false, // already set in web/render.go
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		IsDevelopment:      cnf.IsDevelopment,
	})
	// Start a classic negroni app
	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(gzip.Gzip(gzip.DefaultCompression))
	app.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	app.Use(negroni.NewStatic(http.Dir("public")))

	// Create a router instance
	router := mux.NewRouter()

	// Add routes
	services.HealthService.RegisterRoutes(router, "/v1")
	services.OauthService.RegisterRoutes(router, "/v1/oauth")
	services.WebHookService.RegisterRoutes(router, "/webhook")

	webRoutes := mux.NewRouter()
	services.WebService.RegisterRoutes(webRoutes, "/web")

	CSRF := csrf.Protect(
		[]byte(cnf.CSRF.Key),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.TrustedOrigins([]string{cnf.CSRF.Origins}),
	)

	router.PathPrefix("").Handler(negroni.New(
		negroni.Wrap(CSRF(webRoutes)),
	))

	// Set the router
	app.UseHandler(router)

	log.INFO.Printf("Starting server on localhost%v", cnf.Port)
	// Run the server on port 8080 by default, gracefully stop on SIGTERM signal
	graceful.Run(cnf.Port, 5*time.Second, app)

	return nil
}
