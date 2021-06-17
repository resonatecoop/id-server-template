package oauth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/resonatecoop/id/log"
	"github.com/resonatecoop/id/util"
	pass "github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
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
	// ErrEmailAsLogin
	ErrEmailAsLogin = errors.New("Username cannot be an email address")
	// ErrCountryNotFound
	ErrCountryNotFound = errors.New("Country cannot be found")
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

// CreateUser saves a new user to database
func (s *Service) CreateUser(roleID int32, username, password string) (*model.User, error) {
	return s.createUserCommon(s.db, roleID, username, password)
}

// CreateUserTx saves a new user to database using injected db object
func (s *Service) CreateUserTx(tx *bun.DB, roleID int32, username, password string) (*model.User, error) {
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

func (s *Service) createUserCommon(db *bun.DB, roleID int32, username, password string) (*model.User, error) {
	ctx := context.Background()

	if password == "" {
		return nil, ErrPasswordRequired
	}

	if username == "" {
		return nil, ErrUsernameRequired
	}

	user := &model.User{
		RoleID:   roleID,
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
	_, err = db.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
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
	ctx := context.Background()
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
	_, err := db.NewUpdate().
		Model(&user).
		Set("DeletedAt", time.Now().UTC()).
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

	// userUpdates := &model.User{
	// 	IDRecord: model.IDRecord{
	// 		UpdatedAt: time.Now().UTC(),
	// 	},
	// 	Password: ,
	// }

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

// // Update wp user country (resolve from common name and official name, fallback to alpha code otherwise)
// func (s *Service) UpdateUserCountry(user *model.User, country string) error {
// 	// validate country name
// 	query := gountries.New()
// 	_, err := query.FindCountryByName(strings.ToLower(country))

// 	if err != nil {
// 		// fallback to code
// 		result, err := query.FindCountryByAlpha(strings.ToLower(country))
// 		if err != nil {
// 			return ErrCountryNotFound
// 		}
// 		country = result.Name.Common
// 	}

// 	return s.UpdateUserMetaValue(user.ID, "country", country)
// }

func (s *Service) updateUsernameCommon(db *bun.DB, user *model.User, username string) error {
	ctx := context.Background()
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	// Check the email/username is available
	if s.UserExists(username) {
		return ErrUsernameTaken
	}

	err := db.NewSelect().
		Model(user).
		Where("username = ?", user.Username).
		Scan(ctx)

	if err != nil {
		return err
	}

	_, err = db.NewUpdate().
		Model(user).
		Set("username = ?", strings.ToLower(username)).
		Where("id = ?", user.ID).
		Exec(ctx)

	return err
}
