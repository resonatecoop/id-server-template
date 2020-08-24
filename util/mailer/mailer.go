package mailer

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/RichardKnop/go-oauth2-server/config"
)

func Send(cnf *config.Config, email, template string) (string, string, error) {
	mg := mailgun.NewMailgun(cnf.Mailgun.Domain, cnf.Mailgun.Key)

	sender := cnf.Mailgun.Sender
	subject := "Resonate member details"
	body := ""
	recipient := email

	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(template) // set mailgun template
	message.AddTemplateVariable("email", recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		return "", "", err
	}

	return resp, id, nil
}
