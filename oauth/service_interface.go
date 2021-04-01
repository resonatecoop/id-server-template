package oauth

import (
	"github.com/RichardKnop/go-oauth2-server/config"
	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/session"
	"github.com/RichardKnop/go-oauth2-server/util/routes"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// ServiceInterface defines exported methods
type ServiceInterface interface {
	// Exported methods
	GetConfig() *config.Config
	RestrictToRoles(allowedRoles ...string)
	IsRoleAllowed(role string) bool
	FindRoleByID(id string) (*models.OauthRole, error)
	GetRoutes() []routes.Route
	RegisterRoutes(router *mux.Router, prefix string)
	ClientExists(clientID string) bool
	DeleteClient(clientID string, oauthUser *models.OauthUser) error
	FindClientByClientID(clientID string) (*models.OauthClient, error)
	FindClientsByUserId(oauthUser *models.OauthUser) ([]models.OauthClient, error)
	FindClientByApplicationURL(applicationURL string) (*models.OauthClient, error)
	CreateClient(oauthUser *models.OauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*models.OauthClient, error)
	CreateClientTx(tx *gorm.DB, oauthUser *models.OauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*models.OauthClient, error)
	AuthClient(clientID, secret string) (*models.OauthClient, error)
	GetValidEmailToken(token string) (*models.EmailTokenModel, string, error)
	ClearExpiredEmailTokens() error
	DeleteEmailToken(*models.EmailTokenModel, bool) error
	SendEmailToken(email *models.MailgunEmailModel, emailTokenLink string) (*models.EmailTokenModel, error)
	SendEmailTokenTx(db *gorm.DB, email *models.MailgunEmailModel, emailTokenLink string) (*models.EmailTokenModel, error)
	UserExists(username string) bool
	LoginTaken(login string) bool
	FindUserByUsername(username string) (*models.OauthUser, error)
	CreateUser(roleID, username, password string) (*models.OauthUser, error)
	CreateUserTx(tx *gorm.DB, roleID, username, password string) (*models.OauthUser, error)
	CreateWpUser(email, password, login, displayName string) (*models.WpUser, error)
	CreateWpUserTx(tx *gorm.DB, username, password, login, displayName string) (*models.WpUser, error)
	FindWpUserByLogin(login string) (*models.WpUser, error)
	FindNicknameByWpUserID(id uint64) (string, error)
	FindWpUserByEmail(email string) (*models.WpUser, error)
	ConfirmUserEmail(email string) error
	SetPassword(user *models.OauthUser, password string) error
	SetPasswordTx(tx *gorm.DB, user *models.OauthUser, password string) error
	SetWpPassword(user *models.WpUser, password string) error
	SetWpPasswordTx(tx *gorm.DB, wpuser *models.WpUser, password string) error
	UpdateUsername(user *models.OauthUser, username string) error
	UpdateUsernameTx(db *gorm.DB, user *models.OauthUser, username string) error
	AuthUser(username, thePassword string) (*models.OauthUser, error)
	GetScope(requestedScope string) (string, error)
	GetDefaultScope() string
	ScopeExists(requestedScope string) bool
	Login(client *models.OauthClient, user *models.OauthUser, scope string) (*models.OauthAccessToken, *models.OauthRefreshToken, error)
	GrantAuthorizationCode(client *models.OauthClient, user *models.OauthUser, expiresIn int, redirectURI, scope string) (*models.OauthAuthorizationCode, error)
	GrantAccessToken(client *models.OauthClient, user *models.OauthUser, expiresIn int, scope string) (*models.OauthAccessToken, error)
	GetOrCreateRefreshToken(client *models.OauthClient, user *models.OauthUser, expiresIn int, scope string) (*models.OauthRefreshToken, error)
	GetValidRefreshToken(token string, client *models.OauthClient) (*models.OauthRefreshToken, error)
	Authenticate(token string) (*models.OauthAccessToken, error)
	NewIntrospectResponseFromAccessToken(accessToken *models.OauthAccessToken) (*IntrospectResponse, error)
	NewIntrospectResponseFromRefreshToken(refreshToken *models.OauthRefreshToken) (*IntrospectResponse, error)
	ClearUserTokens(userSession *session.UserSession)
	Close()
}
