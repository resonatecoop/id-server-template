package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/session"

	"github.com/stripe/stripe-go/v72"
	stripeCheckoutSession "github.com/stripe/stripe-go/v72/checkout/session"
	cust "github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/product"
)

// Product is a subset of stripe.Product
type Product struct {
	Name        string   `json:"name"`
	Images      []string `json:"images"`
	Description string   `json:"description"`
	Quantity    int64    `json:"quantity"`
}

// New product creates new Product based on stripe.Product
func (*Service) NewProduct(p *stripe.Product, quantity int64) Product {
	return Product{
		Name:        p.Name,
		Description: p.Description,
		Images:      p.Images,
		Quantity:    quantity,
	}
}

func (s *Service) checkoutForm(w http.ResponseWriter, r *http.Request) {
	sessionService, client, user, isUserAccountComplete, credits, userSession, err := s.profileCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	csrfToken := csrf.Token(r)

	w.Header().Set("X-CSRF-Token", csrfToken)

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

	products := []Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		if err != nil {
			break
		}

		product := s.NewProduct(p, item.Quantity)

		products = append(products, product)
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
		credits,
		usergroups.Usergroup,
		nil,
		nil,
		products,
		csrfToken,
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
		"applicationName":       client.ApplicationName.String,
		"clientID":              client.Key,
		"flash":                 flash,
		"initialState":          template.HTML(fragment),
		"isUserAccountComplete": isUserAccountComplete,
		"products":              products,
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

// checkoutSuccess
func (s *Service) checkoutSuccess(w http.ResponseWriter, r *http.Request) {
	sessionService, _, _, _, _, _, err := s.profileCommon(r)

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

	if checkoutSession.ID == "" {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session has not started yet",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/account", query, w, r)
		return
	}

	cs, err := stripeCheckoutSession.Get(checkoutSession.ID, nil)

	if err != nil {
		// checkout session not started/empty
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

	if cs.Status != "complete" {
		// checkout session still in progress
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: fmt.Sprintf("Checkout session status is: %s", cs.Status),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		query := r.URL.Query()
		redirectWithQueryString("/web/checkout", query, w, r)
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

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "Checkout completed. You should receive an email shortly.",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query = r.URL.Query()
	redirectWithQueryString("/web/account", query, w, r)
	return
}

// checkoutCancel
func (s *Service) checkoutCancel(w http.ResponseWriter, r *http.Request) {
	sessionService, _, _, _, _, _, err := s.profileCommon(r)

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

	if checkoutSession.ID == "" {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: "Checkout session has not started yet",
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

	// expire checkout session
	_, err = stripeCheckoutSession.Expire(
		checkoutSession.ID,
		nil,
	)

	if err != nil {
		// checkout session not started/empty
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

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: "Checkout was canceled",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query = r.URL.Query()
	redirectWithQueryString("/web/account", query, w, r)
	return
}

func (s *Service) checkout(w http.ResponseWriter, r *http.Request) {
	sessionService, _, user, _, _, _, err := s.profileCommon(r)
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

	// set stripe checkout session params
	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("https://" + domain + "/checkout/success"),
		CancelURL:  stripe.String("https://" + domain + "/checkout/cancel"),
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	products := []*stripe.Product{}

	for _, item := range checkoutSession.Products {
		p, err := product.Get(item.ID, nil)

		log.INFO.Printf("Product id: %s", p.ID)

		if err != nil {
			break
		}

		quantity := int64(1) // 1 single share
		credits := int64(0)  // 0 credits amount
		products = append(products, p)

		if item.ID == s.cnf.Stripe.ListenerSubscription.ID ||
			item.ID == s.cnf.Stripe.ArtistMembership.ID ||
			item.ID == s.cnf.Stripe.LabelMembership.ID {
			params.Mode = stripe.String(string(stripe.CheckoutSessionModeSubscription))
			params.AddMetadata("product_id", item.ID)
		}

		if item.ID == s.cnf.Stripe.SupporterShares.ID {
			s := strconv.FormatInt(item.Quantity, 10)
			log.INFO.Printf("Number of shares: %s", s)
			params.AddMetadata("shares", s)
			quantity = item.Quantity
		}

		switch item.ID {
		case s.cnf.Stripe.StreamCredit5.ID:
			credits = 5000
		case s.cnf.Stripe.StreamCredit10.ID:
			credits = 10000
		case s.cnf.Stripe.StreamCredit20.ID:
			credits = 20000
		case s.cnf.Stripe.StreamCredit50.ID:
			credits = 50000
		}

		if credits > 0 {
			s := strconv.FormatInt(credits, 10)
			params.AddMetadata("credits", s)
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

	http.Redirect(w, r, cs.URL, http.StatusFound)
	return
}
