package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
)

func (s *Service) checkoutForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

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
		Country:        user.Country,
		EmailConfirmed: user.EmailConfirmed,
		Complete:       isUserAccountComplete,
		Usergroups:     usergroups.Usergroup,
	}

	if len(usergroups.Usergroup) > 0 {
		profile.DisplayName = usergroups.Usergroup[0].DisplayName
	}

	err = renderTemplate(w, "checkout.html", map[string]interface{}{
		"appURL":                s.cnf.AppURL,
		"staticURL":             s.cnf.StaticURL,
		"isUserAccountComplete": isUserAccountComplete,
		"flash":                 flash,
		"clientID":              client.Key,
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

func (s *Service) checkoutSuccessForm(w http.ResponseWriter, r *http.Request) {
	_, _, _, _, _, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	err = renderTemplate(w, "checkout_success.html", map[string]interface{}{
		"appURL":         s.cnf.AppURL,
		"staticURL":      s.cnf.StaticURL,
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) checkoutCancelForm(w http.ResponseWriter, r *http.Request) {
	_, _, _, _, _, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	err = renderTemplate(w, "checkout_cancel.html", map[string]interface{}{
		"appURL":         s.cnf.AppURL,
		"staticURL":      s.cnf.StaticURL,
		csrf.TemplateTag: csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) checkout(w http.ResponseWriter, r *http.Request) {
	_, _, user, _, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	stripe.Key = s.cnf.Stripe.Secret

	domain := s.cnf.Stripe.Domain

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(s.cnf.Stripe.ListenerSubscriptionPriceID),
				Quantity: stripe.Int64(1),
			},
		},
		CustomerEmail: stripe.String(user.Username),
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:    stripe.String("https://" + domain + "/checkout/success"),
		CancelURL:     stripe.String("https://" + domain + "/checkout/cancel"),
	}

	checkoutSession, err := session.New(params)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userSession.CheckoutSessionID = checkoutSession.ID

	query := r.URL.Query()

	redirectWithQueryString(checkoutSession.URL, query, w, r)
	return
}
