package oauth

import (
	"errors"
	"time"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/util"
	pass "github.com/RichardKnop/go-oauth2-server/util/password"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/trustelem/zxcvbn"
)

var (
	ErrEmailAsLogin = errors.New("Username cannot be an email address")
)

// FindUserByLogin looks up a user by login (wordpress)
func (s *Service) FindWpUserByLogin(login string) (*models.WpUser, error) {
	wpuser := new(models.WpUser)
	notFound := s.db2.Table("rsntr_users").Where("user_login = ?", login).
		First(wpuser).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrUserNotFound
	}

	return wpuser, nil
}

func (s *Service) FindNicknameByWpUserID(id uint64) (string, error) {
	wpusermeta := new(models.WpUserMeta)
	notFound := s.db2.Table("rsntr_usermeta").Select("meta_value").Where("user_id = ? AND meta_key = ?", id, "nickname").
		First(wpusermeta).RecordNotFound()

	// Not found
	if notFound {
		return "", errors.New("Data not found")
	}

	return wpusermeta.MetaValue, nil
}

// CreateUser saves a new user to database
func (s *Service) CreateWpUser(email, password, login, displayName string) (*models.WpUser, error) {
	return s.createWpUserCommon(s.db2, email, password, login, displayName)
}

// CreateUserTx saves a new user to database using injected db object
func (s *Service) CreateWpUserTx(tx *gorm.DB, email, password, login, displayName string) (*models.WpUser, error) {
	return s.createWpUserCommon(tx, email, password, login, displayName)
}

// SetWpPassword sets a wpuser password
func (s *Service) SetWpPassword(wpuser *models.WpUser, password string) error {
	return s.setWpPasswordCommon(s.db2, wpuser, password)
}

// SetWpPasswordTx sets a wpuser password in a transaction
func (s *Service) SetWpPasswordTx(tx *gorm.DB, wpuser *models.WpUser, password string) error {
	return s.setWpPasswordCommon(tx, wpuser, password)
}

func (s *Service) FindWpUserByEmail(email string) (*models.WpUser, error) {
	wpuser := new(models.WpUser)
	wpuserNotFound := s.db2.Table("rsntr_users").Where("user_email = ?", email).
		First(wpuser).RecordNotFound()

	// Not found
	if wpuserNotFound {
		return nil, ErrUserNotFound
	}

	return wpuser, nil
}

func (s *Service) createWpUserCommon(db *gorm.DB, email, password, login, displayName string) (*models.WpUser, error) {
	if displayName == "" {
		return nil, ErrDisplayNameRequired
	}

	if login == "" {
		return nil, ErrLoginRequired
	}

	if password == "" {
		return nil, ErrPasswordRequired
	}

	if email == "" {
		return nil, ErrUsernameRequired
	}

	wpuser := &models.WpUser{
		Email:       email,
		Registered:  time.Now(),
		DisplayName: displayName,
		Login:       login,
		Password:    util.StringOrNull(""),
	}

	// Check if email address is valid
	if !util.ValidateEmail(wpuser.Email) {
		return nil, ErrEmailInvalid
	}

	if len(wpuser.Login) < MinLoginLength {
		return nil, ErrLoginTooShort
	}

	if len(wpuser.Login) > MaxLoginLength {
		return nil, ErrLoginTooLong
	}

	if util.ValidateEmail(wpuser.Login) {
		return nil, ErrEmailAsLogin
	}

	// Check the login is available
	if s.LoginTaken(wpuser.Login) {
		return nil, ErrLoginTaken
	}

	wpuser.Nicename = slug.Make(wpuser.Login)

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

	passwordHashWp, err := pass.HashWpPassword(password)

	if err != nil {
		return nil, err
	}

	wpuser.Password = util.StringOrNull(string(passwordHashWp))

	// Create the wp user
	if err := db.Create(wpuser).Error; err != nil {
		return nil, err
	}

	role := &models.WpUserMeta{
		MetaKey:   "role",
		MetaValue: "fans",
		UserId:    wpuser.ID,
	}

	// Set user role
	if err := db.Create(role).Error; err != nil {
		return nil, err
	}

	nickname := &models.WpUserMeta{
		MetaKey:   "nickname",
		MetaValue: wpuser.DisplayName,
		UserId:    wpuser.ID,
	}

	// Set user nickname
	if err := db.Create(nickname).Error; err != nil {
		return nil, err
	}

	return wpuser, nil
}

func (s *Service) setWpPasswordCommon(db *gorm.DB, wpuser *models.WpUser, password string) error {
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

	passwordHashWp, err := pass.HashWpPassword(password)

	if err != nil {
		return err
	}

	// Set the password on the wpuser object
	err = db.Model(wpuser).UpdateColumn("user_pass", util.StringOrNull(string(passwordHashWp))).Error

	if err != nil {
		return err
	}

	return nil
}
