package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/util"
	"github.com/form3tech-oss/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/mailgun/mailgun-go/v4"
)

var (
	ErrEmailTokenNotFound    = errors.New("This token was not found")
	ErrEmailTokenInvalid     = errors.New("This token is invalid or has expired")
	ErrInvalidEmailTokenLink = errors.New("Email token link is invalid")
)

// GetValidEmailToken ...
func (s *Service) GetValidEmailToken(token string) (*models.EmailTokenModel, string, error) {
	claims := &models.EmailTokenClaimsModel{}

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

	emailToken := new(models.EmailTokenModel)
	notFound := s.db.Where("reference = ?", claims.Reference).
		First(emailToken).RecordNotFound()

	if notFound {
		return nil, "", ErrEmailTokenNotFound
	}

	return emailToken, claims.Username, nil
}

// SendEmailToken ...
func (s *Service) SendEmailToken(
	email *models.MailgunEmailModel,
	emailTokenLink string,
) (*models.EmailTokenModel, error) {
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
	tx *gorm.DB,
	email *models.MailgunEmailModel,
	emailTokenLink string,
) (*models.EmailTokenModel, error) {
	return s.sendEmailTokenCommon(tx, email, emailTokenLink)
}

// CreateEmailToken ...
func (s *Service) CreateEmailToken(email string) (*models.EmailTokenModel, error) {
	expiresIn := 10 * time.Minute // 10 minutes

	emailToken := models.NewOauthEmailToken(&expiresIn)

	if err := s.db.Create(emailToken).Error; err != nil {
		return nil, err
	}

	return emailToken, nil
}

// createJwtTokenWithEmailTokenClaims ...
func (s *Service) createJwtTokenWithEmailTokenClaims(
	claims *models.EmailTokenClaimsModel,
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
	db *gorm.DB,
	email *models.MailgunEmailModel,
	link string,
) (
	*models.EmailTokenModel,
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
	claims := models.NewOauthEmailTokenClaims(email.Recipient, emailToken)

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
		models.EmailTokenModel{
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
	).Delete(&models.EmailTokenModel{}).Error
}

// DeleteEmailToken ...
func (s *Service) DeleteEmailToken(emailToken *models.EmailTokenModel, soft bool) error {
	if soft == true {
		return s.db.Delete(emailToken).Error
	}

	return s.db.Unscoped().Delete(emailToken).Error
}
