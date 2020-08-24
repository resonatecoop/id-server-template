package web

import (
	"fmt"
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/util/mailer"
	"github.com/RichardKnop/go-oauth2-server/util/response"
	"github.com/gorilla/csrf"

	"github.com/RichardKnop/go-oauth2-server/oauth/roles"
	"github.com/RichardKnop/go-oauth2-server/util"
)

func (s *Service) registerForm(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	errMsg, _ := sessionService.GetFlashMessage()
	renderTemplate(w, "register.html", map[string]interface{}{
		"error":          errMsg,
		"queryString":    getQueryString(r.URL.Query()),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func (s *Service) register(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !util.ValidateEmail(r.Form.Get("email")) {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, "Email is not valid", http.StatusBadRequest)
		default:
			sessionService.SetFlashMessage("Email is not valid")
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}

	// Create a user
	_, _, err = s.oauthService.CreateUser(
		roles.User,                 // role ID
		r.Form.Get("email"),        // username
		r.Form.Get("password"),     // password
		r.Form.Get("login"),        // wp login
		r.Form.Get("display_name"), // wp display name
	)
	if err != nil {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, err.Error(), http.StatusBadRequest)
		default:
			sessionService.SetFlashMessage(err.Error())
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}

	recipient := r.Form.Get("email")

	_, _, err = mailer.Send(s.cnf, recipient, "signup")

	if err != nil {
		sessionService.SetFlashMessage(err.Error())
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		message := fmt.Sprintf(
			"A confirmation email has been sent to %s", recipient,
		)
		obj := map[string]interface{}{
			"message": message,
			"status":  http.StatusCreated,
		}
		response.WriteJSON(w, obj, http.StatusCreated)
	} else {
		// Redirect to the login page
		redirectWithQueryString("/web/login", r.URL.Query(), w, r)
	}
}
