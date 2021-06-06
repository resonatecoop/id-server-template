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
	"github.com/resonatecoop/id/util"
	pass "github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
	uuid "github.com/satori/go.uuid"
	"github.com/trustelem/zxcvbn"
	"github.com/uptrace/bun"
)

var (
	// MinPasswordLength defines minimum password length
	MinPasswordLength = 9
	MaxPasswordLength = 72
	MaxLoginLength    = 50
	MinLoginLength    = 3

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

	// ErrPasswordTooShort ...
	ErrPasswordTooShort = fmt.Errorf(
		"Password must be at least %d characters long",
		MinPasswordLength,
	)

	// ErrPasswordTooLong ...
	ErrPasswordTooLong = fmt.Errorf(
		"Password must be at maximum %d characters long",
		MaxPasswordLength,
	)

	// ErrLoginRequired ...
	ErrLoginRequired = errors.New("Login is required")
	// ErrDisplayNameRequired ...
	ErrDisplayNameRequired = errors.New("Display Name is required")
	// ErrPasswordRequired ...
	ErrPasswordRequired = errors.New("Password is required")
	// ErrUsernameRequired ...
	ErrUsernameRequired = errors.New("Email is required")
	// ErrLoginTaken ...
	ErrLoginTaken = errors.New("Login taken")
	// ErrUserNotFound ...
	ErrUserNotFound = errors.New("User not found")
	// ErrInvalidUserPassword ...
	ErrInvalidUserPassword = errors.New("Invalid user password")
	// ErrPasswordTooWeak ...
	ErrPasswordTooWeak = errors.New("Password is too weak")
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
)

// UserExists returns true if user exists
func (s *Service) UserExists(username string) bool {
	_, err := s.FindUserByUsername(username)
	return err == nil
}

func (s *Service) LoginTaken(login string) bool {
	_, err := s.FindWpUserByLogin(login)
	return err == nil
}

// FindUserByUsername looks up a user by username (email)
func (s *Service) FindUserByUsername(username string) (*model.User, error) {
	// Usernames are case insensitive
	user := new(model.User)
	notFound := s.db.Where("username = LOWER(?)", username).
		First(user).RecordNotFound()

	if notFound {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CreateUser saves a new user to database
func (s *Service) CreateUser(roleID, username, password string) (*model.User, error) {
	return s.createUserCommon(s.db, roleID, username, password)
}

// CreateUserTx saves a new user to database using injected db object
func (s *Service) CreateUserTx(tx *bun.DB, roleID, username, password string) (*model.User, error) {
	return s.createUserCommon(tx, roleID, username, password)
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
func (s *Service) UpdateUsername(user *model.User, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	if user.Username == username {
		return ErrUsernameTaken
	}
	// Check the email/username is available
	if s.UserExists(username) {
		return ErrUsernameTaken
	}

	return s.updateUsernameCommon(s.db, user, username)
}

// UpdateUsernameTx ...
func (s *Service) UpdateUsernameTx(tx *bun.DB, user *model.User, username string) error {
	return s.updateUsernameCommon(tx, user, username)
}

func (s *Service) ConfirmUserEmail(email string) error {
	user, err := s.FindUserByUsername(email)

	if err != nil {
		return err
	}

	return s.db.Model(user).UpdateColumn("email_confirmed", true).Error
}

func (s *Service) createUserCommon(db *bun.DB, roleID, username, password string) (*model.User, error) {
	if password == "" {
		return nil, ErrPasswordRequired
	}

	if username == "" {
		return nil, ErrUsernameRequired
	}

	user := &model.User{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		RoleID:   util.StringOrNull(roleID),
		Username: strings.ToLower(username),
		Password: util.StringOrNull(""),
	}

	if len(password) < MinPasswordLength {
		return nil, ErrPasswordTooShort
	}

	if len(password) > MaxPasswordLength {
		return nil, ErrPasswordTooLong
	}

	// enforce strong enough passwords
	passwordStrength := zxcvbn.PasswordStrength(password, nil)

	if passwordStrength.Score < 3 {
		return nil, ErrPasswordTooWeak
	}

	// hash bcrypt password
	passwordHash, err := pass.HashPassword(password)

	if err != nil {
		return nil, err
	}

	user.Password = util.StringOrNull(string(passwordHash))

	// Check if email address is valid
	if !util.ValidateEmail(user.Username) {
		return nil, ErrEmailInvalid
	}

	// Check the email/username is available
	if s.UserExists(user.Username) {
		return nil, ErrUsernameTaken
	}

	// Create the user
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
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
	// Check that the password is set
	/*
		if !user.Password.Valid {
			return ErrUserPasswordNotSet
		}

		// Verify the password
		if pass.VerifyPassword(user.Password.String, password) != nil {
			return ErrInvalidUserPassword
		}
	*/

	// will set deleted_at to current time
	if db.Delete(&user).Error != nil {
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
	err := message.AddTemplateVariable("email", recipient)

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
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	// enforce strong enough passwords
	passwordStrength := zxcvbn.PasswordStrength(password, nil)

	if passwordStrength.Score < 3 {
		return ErrPasswordTooWeak
	}

	// Create a bcrypt hash
	passwordHash, err := pass.HashPassword(password)
	if err != nil {
		return err
	}

	// Set the password on the user object
	err = db.Model(user).UpdateColumns(model.User{
		Password:    util.StringOrNull(string(passwordHash)),
		MyGormModel: model.MyGormModel{UpdatedAt: time.Now().UTC()},
	}).Error

	if err != nil {
		return err
	}

	// Inform user by email password was changed
	mg := mailgun.NewMailgun(s.cnf.Mailgun.Domain, s.cnf.Mailgun.Key)
	sender := s.cnf.Mailgun.Sender
	body := ""
	email := model.NewOauthEmail(
		user.Username,
		"Password changed",
		"password-changed",
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

// Update wp user country (resolve from common name and official name, fallback to alpha code otherwise)
func (s *Service) UpdateUserCountry(user *model.User, country string) error {
	// validate country name
	query := gountries.New()
	_, err := query.FindCountryByName(strings.ToLower(country))

	if err != nil {
		// fallback to code
		result, err := query.FindCountryByAlpha(strings.ToLower(country))
		if err != nil {
			return ErrCountryNotFound
		}
		country = result.Name.Common
	}

	return s.UpdateUserMetaValue(user.ID, "country", country)
}

func (s *Service) updateUsernameCommon(db *bun.DB, user *model.User, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	// Check the email/username is available
	if s.UserExists(username) {
		return ErrUsernameTaken
	}
	return db.Model(user).UpdateColumn("username", strings.ToLower(username)).Error
}
