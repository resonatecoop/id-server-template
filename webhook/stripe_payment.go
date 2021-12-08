package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/user-api/model"

	"github.com/stripe/stripe-go/v72"
	cus "github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
)

// stripePayment is the webhook entry point for stripe payments events
func (s *Service) stripePayment(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	endpointSecret := s.cnf.Stripe.WebHookSecret
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), endpointSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	stripe.Key = s.cnf.Stripe.Secret

	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "resonatecoop/id",
		Version: "0.0.1",
		URL:     "https://github.com/resonatecoop/id",
	})

	switch event.Type {
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("Subscription was created!")
	case "customer.subscription.updated":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("Subscription was updated!")
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO may add some product metadata in email
		// p := &stripe.Product{}
		//
		// for _, item := range subscription.Items.Data {
		// 	if item.Price.Product.ID != s.cnf.Stripe.SupporterShares.ID {
		// 		p, _ = product.Get(item.Price.Product.ID, nil)
		// 		// do something
		// 		break
		// 	}
		// }

		fmt.Println("Subscription was deleted!")

		customer, err := cus.Get(subscription.Customer.ID, nil)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting customer data: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subcriptionListParams := &stripe.SubscriptionListParams{}
		subcriptionListParams.Filters.AddFilter("customer", "", customer.ID)

		subscriptionList := sub.List(subcriptionListParams)
		err = subscriptionList.Err()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting subscription data: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !subscriptionList.Next() {
			err := s.oauthService.GrantMemberStatus(customer.Email, false)
			if err != nil {
				log.ERROR.Print(err)
			}
		}

		mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
		sender := s.cnf.Mailgun.Sender
		body := ""
		email := model.NewOauthEmail(
			customer.Email,
			"Sorry you are leaving!",
			"cancel-subscription",
		)
		subject := email.Subject
		recipient := email.Recipient
		message := mg.NewMessage(sender, subject, body, recipient)
		message.SetTemplate(email.Template) // set mailgun template

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		// Send the message with a 10 second timeout
		_, _, err = mg.Send(ctx, message)

		if err != nil {
			log.ERROR.Print(err)
		}
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.ERROR.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		productID := session.Metadata["product_id"]

		customer, err := cus.Get(session.Customer.ID, nil)

		if err != nil {
			log.ERROR.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = s.processMembership(customer, productID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shares, err := strconv.ParseInt(session.Metadata["shares"], 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if shares > 0 {
			// handle shares
		}
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			log.ERROR.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("PaymentIntent was successful!")
	case "payment_method.attached":
		var paymentMethod stripe.PaymentMethod
		err := json.Unmarshal(event.Data.Raw, &paymentMethod)
		if err != nil {
			log.ERROR.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("PaymentMethod was attached to a Customer!")
		// ... handle other event types
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
	return
}

// processMembership
func (s *Service) processMembership(customer *stripe.Customer, productID string) error {
	member := false

	switch productID {
	case s.cnf.Stripe.ListenerSubscription.ID:
		_ = s.sendWelcomeEmail(customer, "listener-subscription")

		member = true
	case s.cnf.Stripe.ArtistMembership.ID:
		_ = s.sendWelcomeEmail(customer, "artist-subscription")

		// TODO
	case s.cnf.Stripe.LabelMembership.ID:
		_ = s.sendWelcomeEmail(customer, "label-subscription")

		// TODO
	}

	if member == true {
		err := s.oauthService.GrantMemberStatus(customer.Email, true)

		if err != nil {
			log.ERROR.Print(err)
			return err
		}
	}

	return nil
}

// sendWelcomeEmail
func (s *Service) sendWelcomeEmail(customer *stripe.Customer, templateName string) error {
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email := model.NewOauthEmail(
		customer.Email,
		"Welcome to Resonate!",
		templateName,
	)

	subject := email.Subject
	recipient := email.Recipient
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(email.Template) // set mailgun template

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err := mg.Send(ctx, message)

	if err != nil {
		log.ERROR.Print(err)
	}

	return nil
}
