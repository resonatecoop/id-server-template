package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/stripe/stripe-go/v72"
	cust "github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/sub"
)

// Share
type Share struct {
	Amount        int64     `json:"amount"`
	DatePurchased time.Time `json:"datePurchased"`
}

// Membership
type Membership struct {
	SubscriptionID string    `json:"subscriptionID"`
	Name           string    `json:"name"` // ex: Listener membership
	DateFrom       time.Time `json:"dateFrom"`
	DateTo         time.Time `json:"dateTo"`
	Active         bool      `json:"active"`       // ex: active
	Contribution   string    `json:"contribution"` // ex: €5
}

// NewShare
func (s *Service) NewShare(invoice *stripe.Invoice, invoiceLine *stripe.InvoiceLine) Share {
	return Share{
		Amount:        invoiceLine.Amount / 100,
		DatePurchased: time.Unix(invoice.Created, 0).UTC(),
	}
}

// NewMembership
func (s *Service) NewMembership(subscription *stripe.Subscription) Membership {
	var (
		name         string = "Listener"
		sign         string = "€"
		contribution string = ""
	)

	item := subscription.Items.Data[0]

	if item.Price.Product.ID == s.cnf.Stripe.ArtistMembership.ID {
		name = "Artist"
	}

	if item.Price.Product.ID == s.cnf.Stripe.LabelMembership.ID {
		name = "Label"
	}

	if item.Price.Currency == "usd" {
		sign = "$"
	}

	amount := strconv.FormatInt(item.Price.UnitAmount/100, 10)

	if subscription.Status == "active" {
		contribution = "Paid (" + sign + amount + ")"
	}

	return Membership{
		SubscriptionID: subscription.ID,
		Name:           name,
		DateFrom:       time.Unix(subscription.CurrentPeriodStart, 0).UTC(),
		DateTo:         time.Unix(subscription.CurrentPeriodEnd, 0).UTC(),
		Active:         subscription.Status == "active",
		Contribution:   contribution,
	}
}

// membershipForm
func (s *Service) membershipForm(w http.ResponseWriter, r *http.Request) {
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

	ci := cust.List(customerListParams)
	err = ci.Err()

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Err",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account", r.URL.Query(), w, r)
		return
	}

	customer := &stripe.Customer{}

	for ci.Next() {
		customer = ci.Customer()
	}

	subcriptionListParams := &stripe.SubscriptionListParams{}
	subcriptionListParams.Filters.AddFilter("limit", "", "3")

	if query.Get("status") != "" {
		status := query.Get("status")
		subcriptionListParams.Filters.AddFilter("status", "", status)
	}

	subcriptionListParams.Filters.AddFilter("customer", "", customer.ID)

	si := sub.List(subcriptionListParams)
	err = si.Err()

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account", r.URL.Query(), w, r)
		return
	}

	// list subscriptions as memberships
	memberships := []Membership{}

	for si.Next() {
		subscription := si.Subscription()

		subscription, _ = sub.Get(subscription.ID, nil)
		// TODO handle err

		membership := s.NewMembership(subscription)

		memberships = append(memberships, membership)
	}

	// list invoices amounts as shares
	shares := []Share{}

	invoiceListParams := &stripe.InvoiceListParams{}
	invoiceListParams.Filters.AddFilter("limit", "", "50")
	invoiceListParams.Filters.AddFilter("customer", "", customer.ID)

	inv := invoice.List(invoiceListParams)
	err = inv.Err()

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account", r.URL.Query(), w, r)
		return
	}

	for inv.Next() {
		in := inv.Invoice()

		for _, inl := range in.Lines.Data {
			if inl.Price.Product.ID == s.cnf.Stripe.SupporterShares.ID {
				share := s.NewShare(in, inl)
				shares = append(shares, share)
			}
		}
	}

	usergroups, err := s.getUserGroupList(user, userSession.AccessToken)

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/account", r.URL.Query(), w, r)
		return
	}

	initialState, err := json.Marshal(NewInitialState(
		s.cnf,
		client,
		user,
		userSession,
		isUserAccountComplete,
		credits,
		usergroups.Usergroup,
		memberships,
		shares,
		nil,
		csrf.Token(r),
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

	err = renderTemplate(w, "membership.html", map[string]interface{}{
		"appURL":                s.cnf.AppURL,
		"applicationName":       client.ApplicationName.String,
		"clientID":              client.Key,
		"flash":                 flash,
		"initialState":          template.HTML(fragment),
		"isUserAccountComplete": isUserAccountComplete,
		"memberships":           memberships,
		"profile":               profile,
		"queryString":           getQueryString(query),
		"shares":                shares,
		"staticURL":             s.cnf.StaticURL,
		csrf.TemplateTag:        csrf.TemplateField(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// membership
func (s *Service) membership(w http.ResponseWriter, r *http.Request) {
	sessionService, _, _, _, _, _, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = sub.Cancel(r.Form.Get("id"), nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "Membership was cancelled.",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	redirectWithQueryString("/web/membership", query, w, r)
	return
}
