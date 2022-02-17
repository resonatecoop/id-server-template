package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"

	"github.com/resonatecoop/user-api-client/client/users"
	"github.com/resonatecoop/user-api-client/models"
)

var (
	// ErrEmailInvalid
	ErrEmailInvalid = errors.New("Not a valid email")
)

type Country struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (s *Service) joinForm(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	q := gountries.New()
	countries := q.FindAllCountries()

	var countryList []Country

	for i := range countries {
		countryList = append(countryList, Country{
			Name: countries[i].Name.Common,
			Code: countries[i].Codes.Alpha2,
		})
	}

	initialState, _ := json.Marshal(map[string]interface{}{
		"clients":   s.cnf.Clients,
		"countries": countryList,
	})

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	err = renderTemplate(w, "join.html", map[string]interface{}{
		"appURL":         s.cnf.AppURL,
		"countries":      countries,
		"flash":          flash,
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

	user, err := s.createUser(r)

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

	message := fmt.Sprintf(
		"A confirmation email will be sent to %s", user.Username,
	)

	if r.Header.Get("Accept") == "application/json" {
		obj := map[string]interface{}{
			"message": message,
			"status":  http.StatusCreated,
		}
		response.WriteJSON(w, obj, http.StatusCreated)
	} else {
		query := r.URL.Query()
		query.Set("login_redirect_uri", "/web/profile")
		redirectWithQueryString("/web/login", query, w, r)
	}

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

func (s *Service) createUser(r *http.Request) (
	*model.User,
	error,
) {
	// first validate password before calling user-api
	if err := password.ValidatePassword(r.Form.Get("password")); err != nil {
		return nil, err
	}

	// Check if email address is valid
	if !util.ValidateEmail(r.Form.Get("email")) {
		return nil, ErrEmailInvalid
	}

	client := config.NewAPIClient(s.cnf.UserAPIHostname, s.cnf.UserAPIPort)

	params := users.NewResonateUserAddUserParams()

	params.Body = &models.UserUserAddRequest{
		Username: r.Form.Get("email"),
		Country:  r.Form.Get("country"),
	}

	switch r.Form.Get("role") {
	case "artist":
		params.Body.RoleID = int32(model.ArtistRole)
	}
	// case "label":
	//	params.Body.RoleID = int32(model.LabelRole)
	//}

	// Create a user
	_, err := client.Users.ResonateUserAddUser(params, nil)

	if err != nil {
		if casted, ok := err.(*users.ResonateUserAddUserDefault); ok {
			err = errors.New(casted.Payload.Message)
			return nil, err
		}
	}

	user, err := s.oauthService.FindUserByUsername(r.Form.Get("email"))

	if err != nil {
		return nil, err
	}

	if err = s.oauthService.SetPassword(user, r.Form.Get("password")); err != nil {
		return nil, err
	}

	return user, nil
}
