package cmd

import (
	"net/http"
	"time"

	"github.com/RichardKnop/go-oauth2-server/services"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"
	"github.com/RichardKnop/go-oauth2-server/log"
)

// RunServer runs the app
func RunServer(configBackend string) error {
	cnf, db, db2, err := initConfigDB(true, true, configBackend)
	log.INFO.Printf("initConfigDB: %v %v %v %v", cnf, db, db2, err)
	if err != nil {
		return err
	}
	defer db.Close()
	defer db2.Close()

	// start the services
	if err := services.Init(cnf, db, db2); err != nil {
		log.INFO.Printf("Start services %v", err)
		return err
	}
	defer services.Close()

	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          false, // already set in web/render.go
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		IsDevelopment:      cnf.IsDevelopment,
	})
	log.INFO.Print("Starting app")
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
	log.INFO.Printf("Starting server %v", cnf.Port)
	// Run the server on port 8080 by default, gracefully stop on SIGTERM signal
	graceful.Run(cnf.Port, 5*time.Second, app)

	return nil
}
