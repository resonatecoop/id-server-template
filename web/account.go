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
	sessionService, client, user, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isUserAccountComplete := s.isUserAccountComplete(user, userSession.AccessToken)

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

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
		"",
		isUserAccountComplete,
		usergroups.Usergroup,
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

	displayName := ""

	if len(usergroups.Usergroup) > 0 {
		displayName = usergroups.Usergroup[0].DisplayName
	}

	profile := &Profile{
		Email:          user.Username,
		DisplayName:    displayName,
		FullName:       user.FullName,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
		Complete:       isUserAccountComplete,
		Usergroups:     usergroups.Usergroup,
	}

	err = renderTemplate(w, "account.html", map[string]interface{}{
		"isUserAccountComplete": isUserAccountComplete,
		"flash":                 flash,
		"clientID":              client.Key,
		"countries":             countries,
		"applicationName":       client.ApplicationName.String,
		"profile":               profile,
		"queryString":           getQueryString(query),
		"initialState":          template.HTML(fragment),
		csrf.TemplateTag:        csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) account(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	method := strings.ToLower(r.Form.Get("_method"))

	message := "Profile not updated"

	if method == "delete" || r.Method == http.MethodDelete {
		if s.oauthService.DeleteUser(
			user,
			r.Form.Get("password"),
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

		// Delete the access and refresh tokens
		s.oauthService.ClearUserTokens(userSession)

		// Delete the user session
		err = sessionService.ClearUserSession()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		message = "Account is now scheduled for deletion"
	}

	if method == "put" || r.Method == http.MethodPut {
		// username is always email
		if r.Form.Get("email") != "" && r.Form.Get("email") != user.Username {
			if err = s.oauthService.UpdateUsername(
				user,
				r.Form.Get("email"),
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
		}

		// update user (all optional)
		if err = s.oauthService.UpdateUser(
			user,
			r.Form.Get("fullName"),
			r.Form.Get("firstName"),
			r.Form.Get("lastName"),
			r.Form.Get("country"),
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
			} else {
				if len(result.Usergroup) == 0 {
					err = s.createUserGroup(user, r.Form.Get("displayName"), userSession.AccessToken)

					if err != nil {
						log.ERROR.Print(err)
					}
				}
			}
		}

		message = "Account updated"
	}

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"status":  http.StatusOK,
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
	redirectWithQueryString("/web/profile", query, w, r)
}

func (s *Service) isUserAccountComplete(user *model.User, accessToken string) bool {
	// is email address confirmed
	if !user.EmailConfirmed {
		return false
	}

	result, err := s.getUserGroupList(user, accessToken)

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

	if user.FirstName == "" || user.LastName == "" || user.FullName == "" {
		return false
	}

	if user.Country == "" {
		return false
	}

	return true
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

	if err != nil {
		if casted, ok := err.(*usergroups.ResonateUserListUsersUserGroupsDefault); ok {
			return nil, casted
		}
	}

	return result.Payload, err
}

func (s *Service) createUserGroup(user *model.User, displayName, accessToken string) error {
	client := config.NewAPIClient(s.cnf.UserAPIHostname, s.cnf.UserAPIPort)

	bearer := httptransport.BearerToken(accessToken)

	params := usergroups.NewResonateUserAddUserGroupParams()

	params.WithID(user.ID.String())

	params.Body = &models.UserUserGroupCreateRequest{
		DisplayName: displayName,
		GroupType:   "persona",
	}

	_, err := client.Usergroups.ResonateUserAddUserGroup(params, bearer)

	if err != nil {
		// silent
		log.ERROR.Print(err)
	}

	return nil
}
