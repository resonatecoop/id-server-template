package models

import (
	"database/sql"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"

	"github.com/RichardKnop/go-oauth2-server/util"
	"github.com/RichardKnop/uuid"
	"github.com/form3tech-oss/jwt-go"
	"github.com/jinzhu/gorm"
)

// OauthClient ...
type OauthClient struct {
	MyGormModel
	Key                 string      `json:"key,omitempty" sql:"type:varchar(254);unique;not null"`
	Secret              string      `json:"secret,omitempty" sql:"type:varchar(60);not null"`
	UserID              null.String `json:"userId,omitempty,string" sql:"index"`
	RedirectURI         null.String `json:"redirectUri,omitempty,string" sql:"type:varchar(200)"`
	ApplicationName     null.String `json:"applicationName,omitempty,string" sql:"type:varchar(200)"`
	ApplicationHostname null.String `json:"applicationHostname,omitempty,string" sql:"type:varchar(200)"`
	ApplicationURL      null.String `json:"applicationUrl,omitempty,string" sql:"type:varchar(200)"`
	Active              bool        `json:"active" sql:"default:false;not null"`
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
	RoleID         sql.NullString `sql:"type:varchar(20);index;not null"`
	Role           *OauthRole
	Username       string         `sql:"type:varchar(254);unique;not null"`
	Password       sql.NullString `sql:"type:varchar(60)"`
	EmailConfirmed bool           `sql:"default:false;not null"`
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

// TableName specifies table name
func (ac *EmailTokenModel) TableName() string {
	return "oauth_email_tokens"
}

// NewOauthClient creates new OauthClient instance
func NewOauthClient(
	user *OauthUser,
	key string,
	secret string,
	redirectURI string,
	applicationName string,
	applicationHostname string,
	applicationURL string,
	active bool,
) *OauthClient {
	oauthClient := &OauthClient{
		MyGormModel: MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Key:                 key,
		Active:              active,
		Secret:              secret,
		RedirectURI:         null.StringFrom(redirectURI),
		ApplicationName:     null.StringFrom(applicationName),
		ApplicationHostname: null.StringFrom(strings.ToLower(applicationHostname)),
		ApplicationURL:      null.StringFrom(strings.ToLower(applicationURL)),
	}
	if user != nil {
		oauthClient.UserID = null.StringFrom(string(user.ID))
	}
	return oauthClient
}

// NewEmailToken creates new OauthEmailToken instance
func NewOauthEmailToken(expiresIn *time.Duration) *EmailTokenModel {
	return &EmailTokenModel{
		MyGormModel: MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Reference: uuid.New(),
		EmailSent: false,
		ExpiresAt: time.Now().UTC().Add(*expiresIn),
	}
}

// NewOauthEmailTokenClaims creates new NewOauthEmailTokenClaims instance
func NewOauthEmailTokenClaims(email string, emailToken *EmailTokenModel) *EmailTokenClaimsModel {
	return &EmailTokenClaimsModel{
		Username:  email,
		Reference: emailToken.Reference,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: emailToken.ExpiresAt.Unix(),
		},
	}
}

// NewOauthEmail
func NewOauthEmail(email, subject, template string) *MailgunEmailModel {
	return &MailgunEmailModel{
		Recipient: email,
		Subject:   subject,
		Template:  template,
	}
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
