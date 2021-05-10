package oauth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/util"
	pass "github.com/RichardKnop/go-oauth2-server/util/password"
	"github.com/RichardKnop/uuid"
	"github.com/jinzhu/gorm"
	"github.com/trustelem/zxcvbn"
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
	ErrUsernameRequired = errors.New("Username/Email is required")
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
	ErrUsernameTaken = errors.New("Username/Email taken")
	// ErrEmailInvalid
	ErrEmailInvalid = errors.New("Not a valid email")
	// ErrEmailNotFound
	ErrEmailNotFound = errors.New("We can't find an account registered with that address or username")
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
func (s *Service) FindUserByUsername(username string) (*models.OauthUser, error) {
	// Usernames are case insensitive
	user := new(models.OauthUser)
	notFound := s.db.Where("username = LOWER(?)", username).
		First(user).RecordNotFound()

	if notFound {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CreateUser saves a new user to database
func (s *Service) CreateUser(roleID, username, password string) (*models.OauthUser, error) {
	return s.createUserCommon(s.db, roleID, username, password)
}

// CreateUserTx saves a new user to database using injected db object
func (s *Service) CreateUserTx(tx *gorm.DB, roleID, username, password string) (*models.OauthUser, error) {
	return s.createUserCommon(tx, roleID, username, password)
}

// SetPassword sets a user password
func (s *Service) SetPassword(user *models.OauthUser, password string) error {
	return s.setPasswordCommon(s.db, user, password)
}

// SetPasswordTx sets a user password in a transaction
func (s *Service) SetPasswordTx(tx *gorm.DB, user *models.OauthUser, password string) error {
	return s.setPasswordCommon(tx, user, password)
}

// AuthUser authenticates user
func (s *Service) AuthUser(username, password string) (*models.OauthUser, error) {
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
func (s *Service) UpdateUsername(user *models.OauthUser, username string) error {
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
func (s *Service) UpdateUsernameTx(tx *gorm.DB, user *models.OauthUser, username string) error {
	return s.updateUsernameCommon(tx, user, username)
}

func (s *Service) ConfirmUserEmail(email string) error {
	user, err := s.FindUserByUsername(email)

	if err != nil {
		return err
	}

	return s.db.Model(user).UpdateColumn("email_confirmed", true).Error
}

func (s *Service) createUserCommon(db *gorm.DB, roleID, username, password string) (*models.OauthUser, error) {
	if password == "" {
		return nil, ErrPasswordRequired
	}

	if username == "" {
		return nil, ErrUsernameRequired
	}

	user := &models.OauthUser{
		MyGormModel: models.MyGormModel{
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

func (s *Service) setPasswordCommon(db *gorm.DB, user *models.OauthUser, password string) error {
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
	err = db.Model(user).UpdateColumns(models.OauthUser{
		Password:    util.StringOrNull(string(passwordHash)),
		MyGormModel: models.MyGormModel{UpdatedAt: time.Now().UTC()},
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) updateUsernameCommon(db *gorm.DB, user *models.OauthUser, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	// Check the email/username is available
	if s.UserExists(username) {
		return ErrUsernameTaken
	}
	return db.Model(user).UpdateColumn("username", strings.ToLower(username)).Error
}
