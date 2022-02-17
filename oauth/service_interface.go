package oauth

import (
	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/routes"
	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

// ServiceInterface defines exported methods
type ServiceInterface interface {
	// Exported methods
	GetConfig() *config.Config
	RestrictToRoles(allowedRoles ...int32)
	IsRoleAllowed(role int32) bool
	FindRoleByID(id int32) (*model.AccessRole, error)
	GetRoutes() []routes.Route
	RegisterRoutes(router *mux.Router, prefix string)
	ClientExists(clientID string) bool
	FindClientByClientID(clientID string) (*model.Client, error)
	FindClientByApplicationURL(applicationURL string) (*model.Client, error)
	CreateClient(clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*model.Client, error)
	CreateClientTx(tx *bun.DB, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*model.Client, error)
	AuthClient(clientID, secret string) (*model.Client, error)
	GetValidEmailToken(token string) (*model.EmailToken, *model.User, error)
	ClearExpiredEmailTokens() error
	DeleteEmailToken(*model.EmailToken, bool) error
	SendEmailToken(email *model.Email, emailTokenLink string) (*model.EmailToken, error)
	SendEmailTokenTx(db *bun.DB, email *model.Email, emailTokenLink string) (*model.EmailToken, error)
	UserExists(username string) bool
	FindUserByUsername(username string) (*model.User, error)
	FindUserByEmail(email string) (*model.User, error)
	DeleteUser(user *model.User, password string) error
	DeleteUserTx(tx *bun.DB, user *model.User, password string) error
	ConfirmUserEmail(email string) error
	SetPassword(user *model.User, password string) error
	SetPasswordTx(tx *bun.DB, user *model.User, password string) error
	UpdateUsername(user *model.User, username, password string) error
	UpdateUsernameTx(db *bun.DB, user *model.User, username, password string) error
	UpdateUser(user *model.User, fullName, firstName, lastName, country string, newsletter bool) error
	SetUserCountry(user *model.User, country string) error
	SetUserCountryTx(db *bun.DB, user *model.User, country string) error
	AuthUser(username, thePassword string) (*model.User, error)
	GetScope(requestedScope string) (string, error)
	GetDefaultScope() string
	ScopeExists(requestedScope string) bool
	Login(client *model.Client, user *model.User, scope string) (*model.AccessToken, *model.RefreshToken, error)
	GrantAuthorizationCode(client *model.Client, user *model.User, expiresIn int, redirectURI, scope string) (*model.AuthorizationCode, error)
	GrantAccessToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.AccessToken, error)
	GetOrCreateRefreshToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.RefreshToken, error)
	GetValidRefreshToken(token string, client *model.Client) (*model.RefreshToken, error)
	Authenticate(token string) (*model.AccessToken, error)
	NewIntrospectResponseFromAccessToken(accessToken *model.AccessToken) (*IntrospectResponse, error)
	NewIntrospectResponseFromRefreshToken(refreshToken *model.RefreshToken) (*IntrospectResponse, error)
	ClearUserTokens(userSession *session.UserSession)
	Close()
}
