package oauth

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/util"
	pass "github.com/RichardKnop/go-oauth2-server/util/password"
	"github.com/RichardKnop/uuid"
	"github.com/apokalyptik/phpass"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

var (
	phpassVar = phpass.New(phpass.NewConfig())

	phpassMutex = &sync.Mutex{}

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

	// ErrLoginTaken ...
	ErrLoginTaken = errors.New("Login taken")
	// ErrUserNotFound ...
	ErrUserNotFound = errors.New("User not found")
	// ErrInvalidUserPassword ...
	ErrInvalidUserPassword = errors.New("Invalid user password")
	// ErrCannotSetEmptyUsername ...
	ErrCannotSetEmptyUsername = errors.New("Cannot set empty username")
	// ErrUserPasswordNotSet ...
	ErrUserPasswordNotSet = errors.New("User password not set")
	// ErrEmailTaken ...
	ErrEmailTaken = errors.New("Email taken")
)

// UserExists returns true if user exists
func (s *Service) UserExists(username string) bool {
	_, _, err := s.FindUserByUsername(username)
	return err == nil
}

func (s *Service) LoginTaken(login string) bool {
	_, err := s.FindUserByLogin(login)
	return err == nil
}

// FindUserByUsername looks up a user by username (email)
func (s *Service) FindUserByUsername(username string) (*models.OauthUser, *models.WpUser, error) {
	// Usernames are case insensitive
	user := new(models.OauthUser)
	notFound := s.db.Where("username = LOWER(?)", username).
		First(user).RecordNotFound()

	if notFound {
		return nil, nil, ErrUserNotFound
	}

	wpuser := new(models.WpUser)
	wpuserNotFound := s.db2.Table("rsntr_users").Select("user_email").Where("user_email = ?", username).
		First(wpuser).RecordNotFound()

	// Not found
	if wpuserNotFound {
		return nil, nil, ErrUserNotFound
	}

	return user, wpuser, nil
}

// FindUserByLogin looks up a user by login (wordpress)
func (s *Service) FindUserByLogin(login string) (*models.WpUser, error) {
	wpuser := new(models.WpUser)
	notFound := s.db2.Table("rsntr_users").Select("user_login").Where("user_login = ?", login).
		First(wpuser).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrUserNotFound
	}

	return wpuser, nil
}

// CreateUser saves a new user to database
func (s *Service) CreateUser(roleID, username, password, login, displayName string) (*models.OauthUser, *models.WpUser, error) {
	return s.createUserCommon(s.db, s.db2, roleID, username, password, login, displayName)
}

// CreateUserTx saves a new user to database using injected db object
func (s *Service) CreateUserTx(tx *gorm.DB, tx2 *gorm.DB, roleID, username, password, login, displayName string) (*models.OauthUser, *models.WpUser, error) {
	return s.createUserCommon(tx, tx2, roleID, username, password, login, displayName)
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
	user, _, err := s.FindUserByUsername(username)
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

	return s.updateUsernameCommon(s.db, user, username)
}

// UpdateUsernameTx ...
func (s *Service) UpdateUsernameTx(tx *gorm.DB, user *models.OauthUser, username string) error {
	return s.updateUsernameCommon(tx, user, username)
}

func (s *Service) createUserCommon(db *gorm.DB, db2 *gorm.DB, roleID, username, password, login, displayName string) (*models.OauthUser, *models.WpUser, error) {
	// Start with a user without a password
	user := &models.OauthUser{
		MyGormModel: models.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		RoleID:   util.StringOrNull(roleID),
		Username: strings.ToLower(username),
		Password: util.StringOrNull(""),
	}

	wpuser := &models.WpUser{
		Email:      username,
		Registered: time.Now(),
		Password:   util.StringOrNull(""),
	}

	if displayName != "" {
		wpuser.DisplayName = displayName
	}

	if login != "" {
		if len(login) < MinLoginLength {
			return nil, nil, ErrLoginTooShort
		}

		if len(login) > MaxLoginLength {
			return nil, nil, ErrLoginTooLong
		}

		wpuser.Login = login
		wpuser.Nicename = slug.Make(login)
	}

	// If the password is being set already, create a bcrypt and phpass
	if password != "" {
		if len(password) < MinPasswordLength {
			return nil, nil, ErrPasswordTooShort
		}
		if len(password) > MaxPasswordLength {
			return nil, nil, ErrPasswordTooLong
		}
		passwordHash, err := pass.HashPassword(password)
		if err != nil {
			return nil, nil, err
		}
		user.Password = util.StringOrNull(string(passwordHash))

		phpassMutex.Lock()
		passwordHashWp, err := phpassVar.Hash([]byte(password))
		phpassMutex.Unlock()

		if err != nil {
			return nil, nil, err
		}

		wpuser.Password = util.StringOrNull(string(passwordHashWp))
	}

	// Check the email/username is available
	if s.UserExists(user.Username) {
		return nil, nil, ErrEmailTaken
	}

	// Check the login is available
	if s.LoginTaken(wpuser.Login) {
		return nil, nil, ErrLoginTaken
	}

	// Create the user
	if err := db.Create(user).Error; err != nil {
		return nil, nil, err
	}

	// Create the wp user
	if err := db2.Table("rsntr_users").Create(wpuser).Error; err != nil {
		return nil, nil, err
	}

	role := &models.WpUserMeta{
		MetaKey:   "role",
		MetaValue: roleID, // default to `user`
		UserId:    wpuser.ID,
	}

	// Set user role
	if err := db2.Table("rsntr_usermeta").Create(role).Error; err != nil {
		return nil, nil, err
	}

	nickname := &models.WpUserMeta{
		MetaKey:   "nickname",
		MetaValue: wpuser.DisplayName,
		UserId:    wpuser.ID,
	}

	// Set user nickname
	if err := db2.Table("rsntr_usermeta").Create(nickname).Error; err != nil {
		return nil, nil, err
	}

	return user, wpuser, nil
}

func (s *Service) setPasswordCommon(db *gorm.DB, user *models.OauthUser, password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	// Create a bcrypt hash
	passwordHash, err := pass.HashPassword(password)
	if err != nil {
		return err
	}

	// Set the password on the user object
	return db.Model(user).UpdateColumns(models.OauthUser{
		Password:    util.StringOrNull(string(passwordHash)),
		MyGormModel: models.MyGormModel{UpdatedAt: time.Now().UTC()},
	}).Error
}

func (s *Service) updateUsernameCommon(db *gorm.DB, user *models.OauthUser, username string) error {
	if username == "" {
		return ErrCannotSetEmptyUsername
	}
	return db.Model(user).UpdateColumn("username", strings.ToLower(username)).Error
}
