package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/pariz/gountries"
	"github.com/resonatecoop/id/log"
	pass "github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

var (
	// MinPasswordLength defines minimum password length
	MaxLoginLength = 50
	MinLoginLength = 3

	// ErrLoginTooShort ...
	ErrLoginTooShort = fmt.Errorf(
		"Login must be at least %d characters long",
		MinLoginLength,
	)

	// ErrLoginTooShort ...
	ErrLoginTooLong = fmt.Errorf(
		"Login must be at maximum %d characters long",
		MaxLoginLength,
	)

	// ErrLoginRequired ...
	ErrLoginRequired = errors.New("Login is required")
	// ErrDisplayNameRequired ...
	ErrDisplayNameRequired = errors.New("Display Name is required")
	// ErrUsernameRequired ...
	ErrUsernameRequired = errors.New("Email is required")
	// ErrUserNotFound ...
	ErrUserNotFound = errors.New("User not found")
	// ErrInvalidUserPassword ...
	ErrInvalidUserPassword = errors.New("Invalid user password")
	// ErrCannotSetEmptyUsername ...
	ErrCannotSetEmptyUsername = errors.New("Cannot set empty username")
	// ErrUserPasswordNotSet ...
	ErrUserPasswordNotSet = errors.New("User password not set")
	// ErrUsernameTaken ...
	ErrUsernameTaken = errors.New("Email is not available")
	// ErrEmailInvalid
	ErrEmailInvalid = errors.New("Not a valid email")
	// ErrEmailNotFound
	ErrEmailNotFound = errors.New("We can't find an account registered with that address or username")
	// ErrAccountDeletionFailed
	ErrAccountDeletionFailed = errors.New("Account could not be deleted. Please reach to us now")
	// ErrEmailAsLogin
	ErrEmailAsLogin = errors.New("Username cannot be an email address")
	// ErrCountryNotFound
	ErrCountryNotFound = errors.New("Country cannot be found")
	// ErrEmailNotConfirmed
	ErrEmailNotConfirmed = errors.New("Please confirm your email address")
)

// UserExists returns true if user exists
func (s *Service) UserExists(username string) bool {
	_, err := s.FindUserByUsername(username)
	return err == nil
}

// FindUserByUsername looks up a user by username (email)
func (s *Service) FindUserByUsername(username string) (*model.User, error) {
	ctx := context.Background()
	// Usernames are case insensitive
	user := new(model.User)
	err := s.db.NewSelect().
		Model(user).
		Where("username = LOWER(?)", username).
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *Service) FindUserByEmail(email string) (*model.User, error) {
	ctx := context.Background()
	user := new(model.User)
	err := s.db.NewSelect().
		Model(user).
		Where("user_email = ?", email).
		Limit(1).
		Scan(ctx)

	// Not found
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// SetPassword sets a user password
func (s *Service) SetPassword(user *model.User, password string) error {
	return s.setPasswordCommon(s.db, user, password)
}

// SetPasswordTx sets a user password in a transaction
func (s *Service) SetPasswordTx(tx *bun.DB, user *model.User, password string) error {
	return s.setPasswordCommon(tx, user, password)
}

// AuthUser authenticates user
func (s *Service) AuthUser(username, password string) (*model.User, error) {
	// Fetch the user
	user, err := s.FindUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// Check that the password is set
	if !user.Password.Valid {
		return nil, ErrUserPasswordNotSet
	}

	// Verify the password
	if pass.VerifyPassword(user.Password.String, password) != nil {
		return nil, ErrInvalidUserPassword
	}

	return user, nil
}

// UpdateUsername ...
func (s *Service) UpdateUsername(user *model.User, username, password string) error {
	return s.updateUsernameCommon(s.db, user, username, password)
}

// UpdateUsernameTx ...
func (s *Service) UpdateUsernameTx(tx *bun.DB, user *model.User, username, password string) error {
	return s.updateUsernameCommon(tx, user, username, password)
}

func (s *Service) ConfirmUserEmail(email string) error {
	ctx := context.Background()
	user, err := s.FindUserByUsername(email)

	if err != nil {
		return err
	}

	_, err = s.db.NewUpdate().
		Model(user).
		Set("email_confirmed = ?", true).
		WherePK().
		Exec(ctx)

	return err
}

// UpdateUser ...
func (s *Service) UpdateUser(user *model.User, fullName, firstName, lastName, country string, newsletter bool) error {
	return s.updateUserCommon(s.db, user, fullName, firstName, lastName, country, newsletter)
}

// SetUserCountry ...
func (s *Service) SetUserCountry(user *model.User, country string) error {
	return s.setUserCountryCommon(s.db, user, country)
}

// SetUserCountryTx
func (s *Service) SetUserCountryTx(tx *bun.DB, user *model.User, country string) error {
	return s.setUserCountryCommon(tx, user, country)
}

// Delete user will soft delete  user
func (s *Service) DeleteUser(user *model.User, password string) error {
	return s.deleteUserCommon(s.db, user, password)
}

// DeleteUserTx deletes a user in a transaction
func (s *Service) DeleteUserTx(tx *bun.DB, user *model.User, password string) error {
	return s.deleteUserCommon(tx, user, password)
}

func (s *Service) deleteUserCommon(db *bun.DB, user *model.User, password string) error {
	ctx := context.Background()

	// Check that the password is set
	if !user.Password.Valid {
		return ErrUserPasswordNotSet
	}

	// Verify the password
	if pass.VerifyPassword(user.Password.String, password) != nil {
		return ErrInvalidUserPassword
	}

	// will set deleted_at to current time using soft delete
	_, err := db.NewDelete().
		Model(user).
		WherePK().
		Exec(ctx)

	if err != nil {
		return ErrAccountDeletionFailed
	}

	// Inform user account is scheduled for deletion
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email := model.NewOauthEmail(
		user.Username,
		"Account deleted",
		"account-deleted",
	)
	subject := email.Subject
	recipient := email.Recipient
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(email.Template) // set mailgun template
	err = message.AddTemplateVariable("email", recipient)

	if err != nil {
		log.ERROR.Print(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err = mg.Send(ctx, message)

	if err != nil {
		log.ERROR.Print(err)
	}

	return nil
}

func (s *Service) setPasswordCommon(db *bun.DB, user *model.User, password string) error {
	ctx := context.Background()

	// Create a bcrypt hash
	passwordHash, err := pass.HashPassword(password)
	if err != nil {
		return err
	}

	// Set the password on the user object
	_, err = db.NewUpdate().
		Model(user).
		Set("updated_at = ?", time.Now().UTC()).
		Set("last_password_change = ?", time.Now().UTC()).
		Set("password = ?", string(passwordHash)).
		Where("id = ?", user.IDRecord.ID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

// updateUserCommon ...
func (s *Service) updateUserCommon(db *bun.DB, user *model.User, fullName, firstName, lastName, country string, newsletter bool) error {
	ctx := context.Background()

	update := db.NewUpdate().Model(user)

	if country != "" {
		// validate country code
		query := gountries.New()
		_, err := query.FindCountryByAlpha(strings.ToLower(country))

		if err != nil {
			// fallback to name
			result, err := query.FindCountryByName(strings.ToLower(country))
			if err != nil {
				return ErrCountryNotFound
			}
			country = result.Codes.Alpha2
		}

		if country != user.Country {
			update.Set("country = ?", country)
		}
	}

	if newsletter != user.NewsletterNotification {
		update.Set("newsletter_notification = ?", newsletter)
	}

	if fullName != user.FullName && fullName != "" {
		update.Set("full_name = ?", fullName)
	}

	if firstName != user.FirstName && firstName != "" {
		update.Set("first_name = ?", firstName)
	}

	if lastName != user.LastName && lastName != "" {
		update.Set("last_name = ?", lastName)
	}

	_, err := update.Where("id = ?", user.ID).
		Exec(ctx)

	return err
}

// updateUserCountryCommon Update wp user country (resolve from alpha2 or alpha3 code, fallback to common name otherwise)
func (s *Service) setUserCountryCommon(db *bun.DB, user *model.User, country string) error {
	ctx := context.Background()

	// validate country code
	query := gountries.New()
	gountry, err := query.FindCountryByAlpha(strings.ToLower(country))

	if err != nil {
		// fallback to name
		gountry, err = query.FindCountryByName(strings.ToLower(country))
		if err != nil {
			return ErrCountryNotFound
		}
	}

	countryCode := gountry.Codes.Alpha2

	_, err = db.NewUpdate().
		Model(user).
		Set("country = ?", countryCode).
		Where("id = ?", user.ID).
		Exec(ctx)

	return err
}

// updateUsernameCommon ...
func (s *Service) updateUsernameCommon(db *bun.DB, user *model.User, username, password string) error {
	ctx := context.Background()

	if username == "" {
		return ErrCannotSetEmptyUsername
	}

	// Check the email/username is available
	if s.UserExists(username) {
		return ErrUsernameTaken
	}

	// Check that the password is set
	if !user.Password.Valid {
		return ErrUserPasswordNotSet
	}

	// Verify the password
	if pass.VerifyPassword(user.Password.String, password) != nil {
		return ErrInvalidUserPassword
	}

	_, err := db.NewUpdate().
		Model(user).
		Set("username = ?", strings.ToLower(username)).
		Set("email_confirmed = ?", false).
		WherePK().
		Exec(ctx)

	// sends email with token for verification
	email := model.NewOauthEmail(
		username,
		"Confirm email change",
		"email-change-confirmation",
	)

	_, err = s.SendEmailToken(
		email,
		fmt.Sprintf(
			"https://%s/email-confirmation",
			s.cnf.Hostname,
		),
	)

	if err != nil {
		log.ERROR.Print(err)
	}

	// notify current email address
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email = model.NewOauthEmail(
		user.Username,
		"Email change notification",
		"email-change-notification",
	)
	subject := email.Subject
	recipient := email.Recipient
	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(email.Template) // set mailgun template
	err = message.AddTemplateVariable("email", recipient)

	if err != nil {
		log.ERROR.Print(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	_, _, err = mg.Send(ctx, message)

	if err != nil {
		log.ERROR.Print(err)
	}

	return nil
}
