package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	jwt "github.com/form3tech-oss/jwt-go"
	uuid "github.com/google/uuid"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

var (
	ErrEmailTokenNotFound    = errors.New("this token was not found")
	ErrEmailTokenInvalid     = errors.New("this token is invalid or has expired")
	ErrInvalidEmailTokenLink = errors.New("email token link is invalid")
)

// GetValidEmailToken ...
func (s *Service) GetValidEmailToken(token string) (*model.EmailToken, *model.User, error) {
	ctx := context.Background()
	claims := &model.EmailTokenClaims{}

	jwtKey := []byte(s.cnf.EmailTokenSecretKey)

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, nil, err
	}

	if !tkn.Valid {
		return nil, nil, ErrEmailTokenInvalid
	}

	emailToken := new(model.EmailToken)

	err = s.db.NewSelect().
		Model(emailToken).
		Where("reference = ?", claims.Reference).
		Limit(1).
		Scan(ctx)

	// Not Found!
	if err != nil {
		return nil, nil, ErrEmailTokenNotFound
	}

	user, err := s.FindUserByUsername(claims.Username)

	if err != nil {
		return nil, nil, ErrEmailTokenNotFound
	}

	return emailToken, user, nil
}

// SendEmailToken ...
func (s *Service) SendEmailToken(
	email *model.Email,
	emailTokenLink string,
) (*model.EmailToken, error) {
	if !util.ValidateEmail(email.Recipient) {
		return nil, ErrEmailInvalid
	}

	// Check if user is registered
	_, err := s.FindUserByUsername(email.Recipient)

	if err != nil {
		return nil, err
	}

	return s.sendEmailTokenCommon(s.db, email, emailTokenLink)
}

// SendEmailTokenTx ...
func (s *Service) SendEmailTokenTx(
	tx *bun.DB,
	email *model.Email,
	emailTokenLink string,
) (*model.EmailToken, error) {
	return s.sendEmailTokenCommon(tx, email, emailTokenLink)
}

// CreateEmailToken ...
func (s *Service) CreateEmailToken(email string) (*model.EmailToken, error) {
	expiresIn := 10 * time.Minute // 10 minutes

	emailToken := model.NewOauthEmailToken(&expiresIn)

	emailToken.EmailSentAt = &time.Time{}
	emailToken.Reference = uuid.New()

	ctx := context.Background()

	_, err := s.db.NewInsert().Column(
		"id",
		"reference",
		"email_sent_at",
		"email_sent",
		"expires_at",
	).
		Model(emailToken).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return emailToken, nil
}

// createJwtTokenWithEmailTokenClaims ...
func (s *Service) createJwtTokenWithEmailTokenClaims(
	claims *model.EmailTokenClaims,
) (string, error) {
	jwtKey := []byte(s.cnf.EmailTokenSecretKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// sendEmailTokenCommon ...
func (s *Service) sendEmailTokenCommon(
	db *bun.DB,
	email *model.Email,
	link string,
) (
	*model.EmailToken,
	error,
) {
	// Check if email token link is valid
	_, err := url.ParseRequestURI(link)

	if err != nil {
		return nil, ErrInvalidEmailTokenLink
	}

	recipient := email.Recipient

	emailToken, err := s.CreateEmailToken(recipient)

	if err != nil {
		return nil, err
	}

	// Create the JWT claims, which includes the username, expiry time and uuid reference
	claims := model.NewOauthEmailTokenClaims(email.Recipient, emailToken)

	token, err := s.createJwtTokenWithEmailTokenClaims(claims)

	if err != nil {
		return nil, err
	}

	emailTokenLink := fmt.Sprintf(
		"%s?token=%s",
		link, // base url for email token link
		token,
	)

	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	subject := email.Subject
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(email.Template) // set mailgun template
	err = message.AddTemplateVariable("email", email.Recipient)
	if err != nil {
		return nil, err
	}
	err = message.AddTemplateVariable("emailTokenLink", emailTokenLink)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err = mg.Send(ctx, message)

	if err != nil {
		return nil, err
	}

	_, err = s.db.NewUpdate().
		Model(emailToken).
		Set("email_sent = ?", true).
		Set("email_sent_at = ?", time.Now().UTC()).
		Where("reference = ?", emailToken.Reference).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return emailToken, nil
}

// ClearExpiredEmailTokens ...
func (s *Service) ClearExpiredEmailTokens() error {
	ctx := context.Background()

	now := time.Now().UTC()

	emailToken := new(model.EmailToken)

	_, err := s.db.NewDelete().
		Model(emailToken).
		Where(
			"expires_at < ?",
			now.AddDate(0, -30, 0), // 30 days ago
		).
		ForceDelete().
		Exec(ctx)

	return err
}

// DeleteEmailToken ...
func (s *Service) DeleteEmailToken(emailToken *model.EmailToken, soft bool) error {
	ctx := context.Background()

	if soft {

		_, err := s.db.NewDelete().
			Model(emailToken).
			WherePK().
			Exec(ctx)

		return err
	}

	_, err := s.db.NewDelete().
		Model(emailToken).
		WherePK().
		ForceDelete().
		Exec(ctx)

	return err
}
