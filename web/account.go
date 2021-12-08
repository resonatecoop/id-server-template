package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
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

	usergroups, _ := s.getUserGroupList(user, userSession.AccessToken)

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		isUserAccountComplete,
		usergroups.Usergroup,
		nil,
		nil,
		nil,
		"",
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

	profile := &Profile{
		Email:                  user.Username,
		LegacyID:               user.LegacyID,
		FullName:               user.FullName,
		FirstName:              user.FirstName,
		LastName:               user.LastName,
		Member:                 user.Member,
		Country:                user.Country,
		NewsletterNotification: user.NewsletterNotification,
		EmailConfirmed:         user.EmailConfirmed,
		Complete:               isUserAccountComplete,
		Usergroups:             usergroups.Usergroup,
	}

	if len(usergroups.Usergroup) > 0 {
		profile.DisplayName = usergroups.Usergroup[0].DisplayName
	}

	err = renderTemplate(w, "account.html", map[string]interface{}{
		"appURL":                s.cnf.AppURL,
		"staticURL":             s.cnf.StaticURL,
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
	sessionService, _, user, isUserAccountComplete, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	method := strings.ToLower(r.Form.Get("_method"))

	message := "Profile not updated"
	membership := false
	shares := int64(0)
	credits := int64(0)

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
		if r.Form.Get("membership") != "" && user.Member == false {
			membership = true // get membership
		}

		if r.Form.Get("credits") != "" {
			casted, err := strconv.ParseInt(r.Form.Get("credits"), 10, 64)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			credits = casted
		}

		if r.Form.Get("shares") != "" {
			// process supporter shares
			casted, err := strconv.ParseInt(r.Form.Get("shares"), 10, 64)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			shares = casted
		}

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
			r.Form.Get("newsletter") == "subscribe",
			// credits ?
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
					_, err = s.createUserGroup(user, r.Form.Get("displayName"), userSession.AccessToken)

					if err != nil {
						log.ERROR.Print(err)
					}
				}
			}
		}

		message = "Account updated"
	}

	redirectURI := "/web/account"

	if !isUserAccountComplete {
		// if account was completed now, redirects to profile
		isUserAccountComplete = s.isUserAccountComplete(userSession)

		if isUserAccountComplete {
			redirectURI = "/web/profile"
		}
	}

	products := []config.Product{}

	if credits > 0 {
		product := config.Product{}

		switch credits {
		case 5:
			product = s.cnf.Stripe.StreamCredit5
		case 10:
			product = s.cnf.Stripe.StreamCredit10
		case 20:
			product = s.cnf.Stripe.StreamCredit20
		case 50:
			product = s.cnf.Stripe.StreamCredit50
		}

		if product.ID != "" {
			products = append(products, product)
		}
	}

	// should get membership
	if membership == true {
		product := config.Product{}

		switch user.RoleID {
		case int32(model.ArtistRole):
			product = s.cnf.Stripe.ArtistMembership
		case int32(model.LabelRole):
			product = s.cnf.Stripe.LabelMembership
		default:
			product = s.cnf.Stripe.ListenerSubscription
		}

		products = append(products, product)
	}

	if shares > 0 && shares%5 == 0 {
		supporterShares := s.cnf.Stripe.SupporterShares
		supporterShares.Quantity = shares
		products = append(products, supporterShares)
	}

	if len(products) > 0 {
		checkoutSession := &session.CheckoutSession{
			Products: products,
		}
		if err := sessionService.SetCheckoutSession(checkoutSession); err != nil {
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
			return
		}
		redirectURI = "/web/checkout"
	}

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

	if err != nil {
		return nil, err
	}

	return result.Payload, nil
}
