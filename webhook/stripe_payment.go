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
	"github.com/uptrace/bun"

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
	case "customer.created":
		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("Customer was created!")
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
			err := s.GrantMemberStatus(customer.Email, false)
			if err != nil {
				log.ERROR.Print(err)
			}
		}

		if err = s.sendEmail(customer.Email, "Sorry you are leaving!", "cancel-subscription"); err != nil {
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

		if productID != "" {
			fmt.Printf("Product id: %s", productID)
		}

		customer, err := cus.Get(session.Customer.ID, nil)

		if err != nil {
			log.ERROR.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if session.Subscription != nil {
			if err = s.processMembership(customer.Email, session.Subscription.ID, productID); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if session.Metadata["shares"] != "" {
			log.INFO.Printf("Number of shares: %s", session.Metadata["shares"])

			shares, err := strconv.ParseInt(session.Metadata["shares"], 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			user, err := s.oauthService.FindUserByUsername(customer.Email)

			if err != nil {
				log.ERROR.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			invoiceID := ""

			if session.Subscription != nil {
				if session.Subscription.LatestInvoice != nil {
					invoiceID = session.Subscription.LatestInvoice.ID
				}
			}

			if err = s.AddShares(user, invoiceID, shares); err != nil {
				log.ERROR.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if session.Metadata["credits"] != "" {
			log.INFO.Printf("Number of credits: %s", session.Metadata["credits"])

			credits, err := strconv.ParseInt(session.Metadata["credits"], 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			user, err := s.oauthService.FindUserByUsername(customer.Email)

			if err != nil {
				log.ERROR.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err = s.AddCredits(user, credits); err != nil {
				log.ERROR.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
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
func (s *Service) processMembership(customerEmail, subscriptionID, productID string) error {
	user, err := s.oauthService.FindUserByUsername(customerEmail)

	if err != nil {
		log.ERROR.Print(err)
		return err
	}

	templateName := ""

	switch productID {
	case s.cnf.Stripe.ListenerSubscription.ID:
		templateName = "listener-subscription"
	case s.cnf.Stripe.ArtistMembership.ID:
		templateName = "artist-subscription"
	case s.cnf.Stripe.LabelMembership.ID:
		templateName = "label-subscription"
	}

	if templateName != "" {
		if err = s.AddMembership(user, productID, subscriptionID); err != nil {
			log.ERROR.Print(err)
			return err
		}

		if err = s.GrantMemberStatus(customerEmail, true); err != nil {
			log.ERROR.Print(err)
			return err
		}

		if err = s.sendEmail(customerEmail, "Welcome to Resonate!", templateName); err != nil {
			log.ERROR.Print(err)
		}
	}

	return nil
}

// sendEmail
func (s *Service) sendEmail(to, subject, templateName string) error {
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email := model.NewOauthEmail(
		to,
		subject,
		templateName,
	)

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

// GrantMemberStatus
func (s *Service) GrantMemberStatus(email string, status bool) error {
	ctx := context.Background()
	user, err := s.oauthService.FindUserByUsername(email)

	if err != nil {
		return err
	}

	_, err = s.db.NewUpdate().
		Model(user).
		Set("member = ?", status).
		WherePK().
		Exec(ctx)

	return err
}

// AddShares ...
func (s *Service) AddShares(user *model.User, invoiceID string, quantity int64) error {
	return s.addSharesCommon(s.db, user, invoiceID, quantity)
}

// addSharesCommon
func (s *Service) addSharesCommon(db *bun.DB, user *model.User, invoiceID string, quantity int64) error {
	ctx := context.Background()

	shareTransaction := &model.ShareTransaction{
		UserID:    user.ID,
		InvoiceID: invoiceID,
		Quantity:  quantity,
	}

	_, err := db.NewInsert().
		Column(
			"id",
			"user_id",
			"invoice_id", // optional
			"quantity",
		).
		Model(shareTransaction).
		Exec(ctx)

	return err
}

// AddMembership
func (s *Service) AddMembership(user *model.User, productID, subscriptionID string) error {
	return s.addMembershipCommon(s.db, user, productID, subscriptionID)
}

// addMembershipCommon
func (s *Service) addMembershipCommon(db *bun.DB, user *model.User, productID, subscriptionID string) error {
	ctx := context.Background()

	membershipClass := new(model.MembershipClass)

	err := s.db.NewSelect().
		Model(membershipClass).
		Where("product_id = ?", productID).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return err
	}

	membership := &model.UserMembership{
		UserID:            user.ID,
		SubscriptionID:    subscriptionID,
		MembershipClassID: membershipClass.ID,
		MembershipClass:   membershipClass,
		Start:             time.Now().UTC(),
		End:               time.Now().UTC(),
	}

	_, err = db.NewInsert().
		Column(
			"id",
			"user_id",
			"subscription_id",
			"membership_class_id",
			"membership_class",
			"start",
			"end",
		).
		Model(membership).
		Exec(ctx)

	return err
}

// AddCredits ...
func (s *Service) AddCredits(user *model.User, tokens int64) error {
	return s.addCreditsCommon(s.db, user, tokens)
}

// addCreditsCommon ...
func (s *Service) addCreditsCommon(db *bun.DB, user *model.User, amount int64) error {
	ctx := context.Background()

	credit := new(model.Credit)

	err := s.db.NewSelect().
		Model(credit).
		Where("user_id = ?", user.IDRecord.ID).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil
	}

	total := credit.Total + amount

	_, err = db.NewUpdate().
		Model(credit).
		Set("total = ?", total).
		Where("user_id = ?", user.IDRecord.ID).
		Exec(ctx)

	return err
}
