package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

var (
	ErrEmailTokenNotFound    = errors.New("This token was not found")
	ErrEmailTokenInvalid     = errors.New("This token is invalid or has expired")
	ErrInvalidEmailTokenLink = errors.New("Email token link is invalid")
)

// GetValidEmailToken ...
func (s *Service) GetValidEmailToken(token string) (*model.EmailToken, string, error) {
	claims := &model.EmailToken{}

	jwtKey := []byte(s.cnf.EmailTokenSecretKey)

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, "", err
	}

	if !tkn.Valid {
		return nil, "", ErrEmailTokenInvalid
	}

	emailToken := new(model.EmailTokenModel)
	notFound := s.db.Where("reference = ?", claims.Reference).
		First(emailToken).RecordNotFound()

	if notFound {
		return nil, "", ErrEmailTokenNotFound
	}

	return emailToken, claims.Username, nil
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

	// Check if wp user is registered
	_, err = s.FindWpUserByEmail(email.Recipient)

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

	if err := s.db.Create(emailToken).Error; err != nil {
		return nil, err
	}

	return emailToken, nil
}

// createJwtTokenWithEmailTokenClaims ...
func (s *Service) createJwtTokenWithEmailTokenClaims(
	claims *model.EmailToken,
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

	now := time.Now().UTC()

	if err := db.Model(emailToken).Select("email_sent", "email_sent_at").UpdateColumns(
		model.EmailTokenModel{
			EmailSent:   true,
			EmailSentAt: &now,
		},
	).Error; err != nil {
		return nil, err
	}

	return emailToken, nil
}

// ClearExpiredEmailTokens ...
func (s *Service) ClearExpiredEmailTokens() error {
	now := time.Now().UTC()

	return s.db.Unscoped().Where(
		"expires_at < ?",
		now.AddDate(0, -30, 0), // 30 days ago
	).Delete(&model.EmailTokenModel{}).Error
}

// DeleteEmailToken ...
func (s *Service) DeleteEmailToken(emailToken *model.EmailToken, soft bool) error {
	if soft == true {
		return s.db.Delete(emailToken).Error
	}

	return s.db.Unscoped().Delete(emailToken).Error
}
