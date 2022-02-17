package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"
	"github.com/shopspring/decimal"
)

func (s *Service) profileForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, credits, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	if !isUserAccountComplete {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: "Account not complete",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	q := gountries.New()
	countries := q.FindAllCountries()

	usergroups, _ := s.getUserGroupList(user, userSession.AccessToken)

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		isUserAccountComplete,
		credits,
		usergroups.Usergroup,
		nil,
		nil,
		nil,
		"",
		nil,
	))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Inject initial state into choo app
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		string(initialState),
	)

	profile := NewProfile(user, usergroups.Usergroup, isUserAccountComplete, credits, userSession.Role)

	err = renderTemplate(w, "profile.html", map[string]interface{}{
		"appURL":                s.cnf.AppURL,
		"applicationName":       client.ApplicationName.String,
		"clientID":              client.Key,
		"countries":             countries,
		"flash":                 flash,
		"initialState":          template.HTML(fragment),
		"isUserAccountComplete": isUserAccountComplete,
		"profile":               profile,
		"queryString":           getQueryString(query),
		"staticURL":             s.cnf.StaticURL,
		csrf.TemplateTag:        csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) profileCommon(r *http.Request) (
	session.ServiceInterface,
	*model.Client,
	*model.User,
	bool,
	string,
	*session.UserSession,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, nil, nil, false, "", nil, err
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		return nil, nil, nil, false, "", nil, err
	}

	// Get the user session
	userSession, err := sessionService.GetUserSession()
	if err != nil {
		return nil, nil, nil, false, "", nil, err
	}

	// Fetch the user
	user, err := s.oauthService.FindUserByUsername(
		userSession.Username,
	)
	if err != nil {
		return nil, nil, nil, false, "", nil, err
	}

	result, err := s.getUserCredits(user, userSession.AccessToken)

	// Check if user account is complete
	isUserAccountComplete := s.isUserAccountComplete(userSession)

	return sessionService, client, user, isUserAccountComplete, formatCredit(result.Total), userSession, nil
}

// formatCredit
func formatCredit(credits string) string {
	val, _ := decimal.NewFromString(credits)
	result := val.Div(decimal.NewFromInt(1000))
	return result.StringFixed(4)
}

// isUserAccountComplete checks if user account completeness (email confirmation, ...)
func (s *Service) isUserAccountComplete(userSession *session.UserSession) bool {
	user, err := s.oauthService.FindUserByUsername(userSession.Username)

	if err != nil {
		return false
	}

	// is email address confirmed
	if !user.EmailConfirmed {
		return false
	}

	result, err := s.getUserGroupList(user, userSession.AccessToken)

	if err != nil {
		return false
	}

	if len(result.Usergroup) == 0 {
		return false
	}

	// listeners only need to confirm their email address
	if user.RoleID == int32(model.UserRole) {
		return true
	}

	// if user.FirstName == "" || user.LastName == "" || user.FullName == "" {
	// 	return false
	// }

	if user.Country == "" {
		return false
	}

	return true
}
