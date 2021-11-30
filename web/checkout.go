package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"

	"github.com/stripe/stripe-go/v72"
	stripeCheckoutSession "github.com/stripe/stripe-go/v72/checkout/session"
	cust "github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/product"
)

func (s *Service) checkoutForm(w http.ResponseWriter, r *http.Request) {
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

	checkoutSession, err := sessionService.GetCheckoutSession()
	if err != nil {
		// checkout session not started/empty
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session is empty",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// li := &stripe.LineItem{}
	// p := &stripe.Product{}

	if len(checkoutSession.Products) == 0 {
		// checkout session not started/empty
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "No product set",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	products := []stripe.Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		if err != nil {
			break
		}

		products = append(products, *p)
	}

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// cs, err := stripeCheckoutSession.Get(
	// 	checkoutSession.ID,
	// 	nil,
	// )
	//
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	//
	// params := &stripe.CheckoutSessionListLineItemsParams{}
	//
	// params.Filters.AddFilter("limit", "", "5")
	// i := stripeCheckoutSession.ListLineItems(cs.ID, params)
	//
	// for i.Next() {
	// 	li = i.LineItem()
	// }

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
		"products":              products,
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
	sessionService, client, user, isUserAccountComplete, userSession, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	checkoutSession, err := sessionService.GetCheckoutSession()
	if err != nil {
		// checkout session not started/empty
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session is empty",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	products := []*stripe.Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		if err != nil {
			break
		}

		products = append(products, p)
	}

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// Delete the checkout session
	err = sessionService.ClearCheckoutSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	err = renderTemplate(w, "checkout_success.html", map[string]interface{}{
		"products":              products,
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

func (s *Service) checkoutCancelForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, userSession, err := s.profileCommon(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	flash, _ := sessionService.GetFlashMessage()
	query := r.URL.Query()
	query.Set("login_redirect_uri", r.URL.Path)

	checkoutSession, err := sessionService.GetCheckoutSession()
	if err != nil {
		// checkout session not started/empty
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session is empty",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	products := []*stripe.Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		if err != nil {
			break
		}

		products = append(products, p)
	}

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	// Delete the checkout session
	err = sessionService.ClearCheckoutSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	err = renderTemplate(w, "checkout_cancel.html", map[string]interface{}{
		"products":              products,
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

func (s *Service) checkout(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, _, _, err := s.profileCommon(r)
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

	checkoutSession, err := sessionService.GetCheckoutSession()
	if err != nil {
		// checkout session not started/empty
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session is empty",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	domain := s.cnf.Stripe.Domain

	// retrieve stripe customer by email
	customerListParams := &stripe.CustomerListParams{}
	customerListParams.Filters.AddFilter("limit", "", "1")
	customerListParams.Filters.AddFilter("email", "", user.Username)

	i := cust.List(customerListParams)
	err = i.Err()

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	customer := &stripe.Customer{}

	for i.Next() {
		customer = i.Customer()
	}

	// set checkout session params
	params := &stripe.CheckoutSessionParams{
		// SubmitType: stripe.String("donate"),
		// BillingAddressCollection: stripe.String("auto"),
		// LineItems: []*stripe.CheckoutSessionLineItemParams{
		// 	&stripe.CheckoutSessionLineItemParams{
		// 		Price:    stripe.String(s.cnf.Stripe.ListenerSubscription.PriceID),
		// 		Quantity: stripe.Int64(1),
		// 	},
		// },
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String("https://" + domain + "/checkout/success"),
		CancelURL:  stripe.String("https://" + domain + "/checkout/cancel"),
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	products := []*stripe.Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		if err != nil {
			break
		}

		products = append(products, p)

		if item.ID == s.cnf.Stripe.ListenerSubscription.ID {
			params.AddMetadata("product_id", item.ID)
		}

		quantity := int64(1)

		if item.ID == s.cnf.Stripe.SupporterShares.ID {
			s := strconv.FormatInt(item.Quantity, 10)
			params.AddMetadata("shares", s)
			quantity = item.Quantity
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(quantity),
		})
	}

	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	params.LineItems = lineItems

	if customer.ID != "" {
		// use existing customer
		params.Customer = stripe.String(customer.ID)
	} else {
		// should create new customer
		params.CustomerEmail = stripe.String(user.Username)
	}

	cs, err := stripeCheckoutSession.New(params)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	checkoutSession.ID = cs.ID

	http.Redirect(w, r, cs.URL, http.StatusFound)
	return
}
