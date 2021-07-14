package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"

	"github.com/resonatecoop/user-api-client/client/users"
	"github.com/resonatecoop/user-api-client/models"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/resonatecoop/user-api-client/client"
)

func (s *Service) joinForm(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	initialState, _ := json.Marshal(map[string]interface{}{
		"clients": s.cnf.Clients,
	})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	q := gountries.New()
	countries := q.FindAllCountries()

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	err = renderTemplate(w, "join.html", map[string]interface{}{
		"flash":          flash,
		"countries":      countries,
		"initialState":   template.HTML(fragment),
		"queryString":    getQueryString(r.URL.Query()),
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) join(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpClient, _ := httptransport.TLSClient(httptransport.TLSClientOptions{
		InsecureSkipVerify: true,
	})

	transport := httptransport.NewWithClient("api.resonate.localhost", "", nil, httpClient)

	// create the API client, with the transport
	client := apiclient.New(transport, strfmt.Default)
	bearer := httptransport.BearerToken("test_token_superadmin")

	// Create a user
	params := users.NewResonateUserAddUserParamsWithTimeout(10000000)

	params.Body = &models.UserUserAddRequest{
		Username: r.Form.Get("email"),
		Country:  r.Form.Get("country"),
	}

	_, err = client.Users.ResonateUserAddUser(params, bearer)

	if err != nil {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, err.Error(), http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}

	// message := fmt.Sprintf(
	//	"A confirmation email will be sent to %s", user.Username,
	//)

	/*
		if r.Header.Get("Accept") == "application/json" {
			obj := map[string]interface{}{
				"message": message,
				"status":  http.StatusCreated,
			}
			response.WriteJSON(w, obj, http.StatusCreated)
		} else {
			query := r.URL.Query()
			query.Set("login_redirect_uri", "/web/welcome")
			redirectWithQueryString("/web/login", query, w, r)
		}
	*/

	_, err = s.oauthService.SendEmailToken(
		model.NewOauthEmail(
			r.Form.Get("email"), // Recipient
			"Member details",    // Subject
			"signup",            // Template (mailgun)
		),
		fmt.Sprintf(
			"https://%s/email-confirmation",
			s.cnf.Hostname,
		),
	)

	if err != nil {
		log.ERROR.Print(err)
	}
}
