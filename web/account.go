package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"

	"github.com/resonatecoop/user-api-client/client/usergroups"
	"github.com/resonatecoop/user-api-client/models"

	httptransport "github.com/go-openapi/runtime/client"
)

func (s *Service) accountForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	// Render the template
	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	q := gountries.New()
	countries := q.FindAllCountries()

	var countryList []Country

	for i := range countries {
		countryList = append(countryList, Country{
			Name: countries[i].Name.Common,
			Code: countries[i].Codes.Alpha2,
		})
	}

	usergroups, _ := s.getUserGroupList(user, userSession.AccessToken)

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		isUserAccountComplete,
		usergroups.Usergroup,
		"",
		countryList,
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

	profile := NewProfile(user, usergroups.Usergroup, isUserAccountComplete, userSession.Role)

	err = renderTemplate(w, "account.html", map[string]interface{}{
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

func (s *Service) account(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, isUserAccountComplete, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	method := strings.ToLower(r.Form.Get("_method"))

	message := "Profile not updated"

	if method == "put" || r.Method == http.MethodPut {
		// update user (all optional)
		if err = s.oauthService.UpdateUser(
			user,
			r.Form.Get("fullName"),
			r.Form.Get("firstName"),
			r.Form.Get("lastName"),
			r.Form.Get("country"),
			r.Form.Get("newsletter") == "subscribe",
		); err != nil {
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

		// init usergroup
		if r.Form.Get("displayName") != "" {
			result, err := s.getUserGroupList(user, userSession.AccessToken)

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

			if len(result.Usergroup) == 0 {
				_, err = s.createUserGroup(user, r.Form.Get("displayName"), userSession.AccessToken)

				if err != nil {
					log.ERROR.Print(err)
				}
			}
		}

		message = "Account updated"
	}

	redirectURI := "/web/account"

	// if !isUserAccountComplete {
	// 	// if account was completed now, redirects to profile
	// 	isUserAccountComplete = s.isUserAccountComplete(userSession)
	//
	// 	if isUserAccountComplete {
	// 		redirectURI = "/web/profile"
	// 	}
	// }

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"data": map[string]interface{}{
				"success_redirect_url": redirectURI,
				"profile_redirection":  redirectURI == "/web/profile",
				"account_complete":     isUserAccountComplete,
			},
			"status": http.StatusOK,
		}, http.StatusOK)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: message,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	query := r.URL.Query()

	redirectWithQueryString(redirectURI, query, w, r)
	return
}

func (s *Service) getUserGroupList(user *model.User, accessToken string) (
	*models.UserUserGroupListResponse,
	error,
) {
	client := config.NewAPIClient(s.cnf.UserAPIHostname, s.cnf.UserAPIPort)

	bearer := httptransport.BearerToken(accessToken)

	params := usergroups.NewResonateUserListUsersUserGroupsParams()

	params.WithID(user.ID.String())

	result, err := client.Usergroups.ResonateUserListUsersUserGroups(params, bearer)

	if result == nil {
		panic("User API not started")
	}

	if err != nil {
		if casted, ok := err.(*usergroups.ResonateUserListUsersUserGroupsDefault); ok {
			return nil, casted
		}
	}

	return result.Payload, err
}

func (s *Service) createUserGroup(user *model.User, displayName, accessToken string) (*models.UserUserRequest, error) {
	client := config.NewAPIClient(s.cnf.UserAPIHostname, s.cnf.UserAPIPort)

	bearer := httptransport.BearerToken(accessToken)

	params := usergroups.NewResonateUserAddUserGroupParams()

	params.WithID(user.ID.String())

	params.Body = &models.UserUserGroupCreateRequest{
		DisplayName: displayName,
		GroupType:   "persona",
	}

	result, err := client.Usergroups.ResonateUserAddUserGroup(params, bearer)

	if result == nil {
		panic("User API not started")
	}

	if err != nil {
		return nil, err
	}

	return result.Payload, nil
}
