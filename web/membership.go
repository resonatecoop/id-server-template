package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/pariz/gountries"
	"github.com/stripe/stripe-go/v72"
	cust "github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/sub"
)

func (s *Service) membershipForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	stripe.Key = s.cnf.Stripe.Secret

	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "resonatecoop/id",
		Version: "0.0.1",
		URL:     "https://github.com/resonatecoop/id",
	})

	// retrieve stripe customer by email
	customerListParams := &stripe.CustomerListParams{}
	customerListParams.Filters.AddFilter("limit", "", "1")
	customerListParams.Filters.AddFilter("email", "", user.Username)

	i := cust.List(customerListParams)

	customer := &stripe.Customer{}

	for i.Next() {
		customer = i.Customer()
	}

	subcriptionListParams := &stripe.SubscriptionListParams{}
	subcriptionListParams.Filters.AddFilter("limit", "", "1")
	subcriptionListParams.Filters.AddFilter("customer", "", customer.ID)

	si := sub.List(subcriptionListParams)

	subscription := &stripe.Subscription{}

	for si.Next() {
		subscription = si.Subscription()
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

	profile := &Profile{
		Email:          user.Username,
		LegacyID:       user.LegacyID,
		FullName:       user.FullName,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
		Complete:       isUserAccountComplete,
		Usergroups:     usergroups.Usergroup,
	}

	if len(usergroups.Usergroup) > 0 {
		profile.DisplayName = usergroups.Usergroup[0].DisplayName
	}

	err = renderTemplate(w, "membership.html", map[string]interface{}{
		"customer":              customer,
		"subscription":          subscription,
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

func (s *Service) membership(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) cancelSubscription(w http.ResponseWriter, r *http.Request) {
}
