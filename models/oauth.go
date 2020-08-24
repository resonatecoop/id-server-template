package models

import (
	"database/sql"
	"time"

	"github.com/RichardKnop/go-oauth2-server/util"
	"github.com/RichardKnop/uuid"
	"github.com/jinzhu/gorm"
)

// OauthClient ...
type OauthClient struct {
	MyGormModel
	Key                 string         `sql:"type:varchar(254);unique;not null"`
	Secret              string         `sql:"type:varchar(60);not null"`
	RedirectURI         sql.NullString `sql:"type:varchar(200)"`
	ApplicationName     sql.NullString `sql:"type:varchar(200)"`
	ApplicationHostname sql.NullString `sql:"type:varchar(200)"`
}

// TableName specifies table name
func (c *OauthClient) TableName() string {
	return "oauth_clients"
}

// OauthScope ...
type OauthScope struct {
	MyGormModel
	Scope       string `sql:"type:varchar(200);unique;not null"`
	Description sql.NullString
	IsDefault   bool `sql:"default:false"`
}

// TableName specifies table name
func (s *OauthScope) TableName() string {
	return "oauth_scopes"
}

// OauthRole is a one of roles user can have (currently superuser or user)
type OauthRole struct {
	TimestampModel
	ID   string `gorm:"primary_key" sql:"type:varchar(20)"`
	Name string `sql:"type:varchar(50);unique;not null"`
}

// TableName specifies table name
func (r *OauthRole) TableName() string {
	return "oauth_roles"
}

// OauthUser ...
type OauthUser struct {
	MyGormModel
	RoleID   sql.NullString `sql:"type:varchar(20);index;not null"`
	Role     *OauthRole
	Username string         `sql:"type:varchar(254);unique;not null"`
	Password sql.NullString `sql:"type:varchar(60)"`
}

// WpUser ...
type WpUser struct {
	ID          uint64         `gorm:"column:ID;primary_key;auto_increment;not_null"`
	Email       string         `gorm:"column:user_email" sql:"type:varchar(100);unique;not null"`
	Login       string         `gorm:"column:user_login" sql:"type:varchar(60)"`
	Registered  time.Time      `gorm:"column:user_registered" sql:"type:datetime"`
	Nicename    string         `gorm:"column:user_nicename" sql:"type:varchar(50)"`
	DisplayName string         `gorm:"column:display_name" sql:"type:varchar(250)"`
	Password    sql.NullString `gorm:"column:user_pass" sql:"type:varchar(255)"`
}

// UserMeta ...
type WpUserMeta struct {
	ID        uint64 `gorm:"column:umeta_id;primary_key;auto_increment;not_null"`
	UserId    uint64 `gorm:"column:user_id" sql:"not null"`
	MetaKey   string `gorm:"column:meta_key" sql:"type:varchar(255)"`
	MetaValue string `gorm:"column:meta_value" sql:"type:longtext"`
}

// TableName specifies table name
func (u *OauthUser) TableName() string {
	return "oauth_users"
}

// OauthRefreshToken ...
type OauthRefreshToken struct {
	MyGormModel
	ClientID  sql.NullString `sql:"index;not null"`
	UserID    sql.NullString `sql:"index"`
	Client    *OauthClient
	User      *OauthUser
	Token     string    `sql:"type:varchar(40);unique;not null"`
	ExpiresAt time.Time `sql:"not null"`
	Scope     string    `sql:"type:varchar(200);not null"`
}

// TableName specifies table name
func (rt *OauthRefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}

// OauthAccessToken ...
type OauthAccessToken struct {
	MyGormModel
	ClientID  sql.NullString `sql:"index;not null"`
	UserID    sql.NullString `sql:"index"`
	Client    *OauthClient
	User      *OauthUser
	Token     string    `sql:"type:varchar(40);unique;not null"`
	ExpiresAt time.Time `sql:"not null"`
	Scope     string    `sql:"type:varchar(200);not null"`
}

// TableName specifies table name
func (at *OauthAccessToken) TableName() string {
	return "oauth_access_tokens"
}

// OauthAuthorizationCode ...
type OauthAuthorizationCode struct {
	MyGormModel
	ClientID    sql.NullString `sql:"index;not null"`
	UserID      sql.NullString `sql:"index;not null"`
	Client      *OauthClient
	User        *OauthUser
	Code        string         `sql:"type:varchar(40);unique;not null"`
	RedirectURI sql.NullString `sql:"type:varchar(200)"`
	ExpiresAt   time.Time      `sql:"not null"`
	Scope       string         `sql:"type:varchar(200);not null"`
}

// TableName specifies table name
func (ac *OauthAuthorizationCode) TableName() string {
	return "oauth_authorization_codes"
}

// NewOauthRefreshToken creates new OauthRefreshToken instance
func NewOauthRefreshToken(client *OauthClient, user *OauthUser, expiresIn int, scope string) *OauthRefreshToken {
	refreshToken := &OauthRefreshToken{
		MyGormModel: MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		ClientID:  util.StringOrNull(string(client.ID)),
		Token:     uuid.New(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		Scope:     scope,
	}
	if user != nil {
		refreshToken.UserID = util.StringOrNull(string(user.ID))
	}
	return refreshToken
}

// NewOauthAccessToken creates new OauthAccessToken instance
func NewOauthAccessToken(client *OauthClient, user *OauthUser, expiresIn int, scope string) *OauthAccessToken {
	accessToken := &OauthAccessToken{
		MyGormModel: MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		ClientID:  util.StringOrNull(string(client.ID)),
		Token:     uuid.New(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		Scope:     scope,
	}
	if user != nil {
		accessToken.UserID = util.StringOrNull(string(user.ID))
	}
	return accessToken
}

// NewOauthAuthorizationCode creates new OauthAuthorizationCode instance
func NewOauthAuthorizationCode(client *OauthClient, user *OauthUser, expiresIn int, redirectURI, scope string) *OauthAuthorizationCode {
	return &OauthAuthorizationCode{
		MyGormModel: MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		ClientID:    util.StringOrNull(string(client.ID)),
		UserID:      util.StringOrNull(string(user.ID)),
		Code:        uuid.New(),
		ExpiresAt:   time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		RedirectURI: util.StringOrNull(redirectURI),
		Scope:       scope,
	}
}

// OauthAuthorizationCodePreload sets up Gorm preloads for an auth code object
func OauthAuthorizationCodePreload(db *gorm.DB) *gorm.DB {
	return OauthAuthorizationCodePreloadWithPrefix(db, "")
}

// OauthAuthorizationCodePreloadWithPrefix sets up Gorm preloads for an auth code object,
// and prefixes with prefix for nested objects
func OauthAuthorizationCodePreloadWithPrefix(db *gorm.DB, prefix string) *gorm.DB {
	return db.
		Preload(prefix + "Client").Preload(prefix + "User")
}

// OauthAccessTokenPreload sets up Gorm preloads for an access token object
func OauthAccessTokenPreload(db *gorm.DB) *gorm.DB {
	return OauthAccessTokenPreloadWithPrefix(db, "")
}

// OauthAccessTokenPreloadWithPrefix sets up Gorm preloads for an access token object,
// and prefixes with prefix for nested objects
func OauthAccessTokenPreloadWithPrefix(db *gorm.DB, prefix string) *gorm.DB {
	return db.
		Preload(prefix + "Client").Preload(prefix + "User")
}

// OauthRefreshTokenPreload sets up Gorm preloads for a refresh token object
func OauthRefreshTokenPreload(db *gorm.DB) *gorm.DB {
	return OauthRefreshTokenPreloadWithPrefix(db, "")
}

// OauthRefreshTokenPreloadWithPrefix sets up Gorm preloads for a refresh token object,
// and prefixes with prefix for nested objects
func OauthRefreshTokenPreloadWithPrefix(db *gorm.DB, prefix string) *gorm.DB {
	return db.
		Preload(prefix + "Client").Preload(prefix + "User")
}
