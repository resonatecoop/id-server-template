package mocks

import (
	"github.com/resonatecoop/id/config"
	//"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/user-api/model"

	//"github.com/resonatecoop/id/oauth"
	//"github.com/resonatecoop/id/oauth"
	uuid "github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/util/routes"
	"github.com/stretchr/testify/mock"
	"github.com/uptrace/bun"
)

type ServiceInterface struct {
	mock.Mock
}

func (_m *ServiceInterface) GetConfig() *config.Config {
	ret := _m.Called()

	var r0 *config.Config
	if rf, ok := ret.Get(0).(func() *config.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*config.Config)
		}
	}

	return r0
}
func (_m *ServiceInterface) RestrictToRoles(allowedRoles ...int32) {
	_m.Called(allowedRoles)
}
func (_m *ServiceInterface) IsRoleAllowed(role model.AccessRole) bool {
	ret := _m.Called(role)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(string(role))
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
func (_m *ServiceInterface) GetRoutes() []routes.Route {
	ret := _m.Called()

	var r0 []routes.Route
	if rf, ok := ret.Get(0).(func() []routes.Route); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]routes.Route)
		}
	}

	return r0
}
func (_m *ServiceInterface) RegisterRoutes(router *mux.Router, prefix string) {
	_m.Called(router, prefix)
}
func (_m *ServiceInterface) ClientExists(clientID uuid.UUID) bool {
	ret := _m.Called(clientID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(clientID.String())
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
func (_m *ServiceInterface) FindClientByClientID(clientID uuid.UUID) (*model.Client, error) {
	ret := _m.Called(clientID)

	var r0 *model.Client
	if rf, ok := ret.Get(0).(func(string) *model.Client); ok {
		r0 = rf(clientID.String())
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(clientID.String())
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) CreateClient(clientID uuid.UUID, secret string, redirectURI string) (*model.Client, error) {
	ret := _m.Called(clientID.String(), secret, redirectURI)

	var r0 *model.Client
	if rf, ok := ret.Get(0).(func(string, string, string) *model.Client); ok {
		r0 = rf(clientID.String(), secret, redirectURI)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(clientID.String(), secret, redirectURI)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) CreateClientTx(tx *bun.DB, clientID uuid.UUID, secret string, redirectURI string) (*model.Client, error) {
	ret := _m.Called(tx, clientID.String(), secret, redirectURI)

	var r0 *model.Client
	if rf, ok := ret.Get(0).(func(*bun.DB, string, string, string) *model.Client); ok {
		r0 = rf(tx, clientID.String(), secret, redirectURI)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*bun.DB, string, string, string) error); ok {
		r1 = rf(tx, clientID.String(), secret, redirectURI)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) AuthClient(clientID uuid.UUID, secret string) (*model.Client, error) {
	ret := _m.Called(clientID.String(), secret)

	var r0 *model.Client
	if rf, ok := ret.Get(0).(func(string, string) *model.Client); ok {
		r0 = rf(clientID.String(), secret)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Client)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(clientID.String(), secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) UserExists(username string) bool {
	ret := _m.Called(username)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
func (_m *ServiceInterface) FindUserByUsername(username string) (*model.User, error) {
	ret := _m.Called(username)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string) *model.User); ok {
		r0 = rf(username)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) CreateUser(roleID int32, username string, password string) (*model.User, error) {
	ret := _m.Called(roleID, username, password)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string, string, string) *model.User); ok {
		r0 = rf(string(roleID), username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(string(roleID), username, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) CreateUserTx(tx *bun.DB, roleID int32, username string, password string) (*model.User, error) {
	ret := _m.Called(tx, roleID, username, password)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(*bun.DB, string, string, string) *model.User); ok {
		r0 = rf(tx, string(roleID), username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*bun.DB, string, string, string) error); ok {
		r1 = rf(tx, string(roleID), username, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) SetPassword(user *model.User, password string) error {
	ret := _m.Called(user, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.User, string) error); ok {
		r0 = rf(user, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *ServiceInterface) SetPasswordTx(tx *bun.DB, user *model.User, password string) error {
	ret := _m.Called(tx, user, password)

	var r0 error
	if rf, ok := ret.Get(0).(func(*bun.DB, *model.User, string) error); ok {
		r0 = rf(tx, user, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *ServiceInterface) UpdateUsername(user *model.User, username string) error {
	ret := _m.Called(user, username)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.User, string) error); ok {
		r0 = rf(user, username)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *ServiceInterface) UpdateUsernameTx(db *bun.DB, user *model.User, username string) error {
	ret := _m.Called(db, user, username)

	var r0 error
	if rf, ok := ret.Get(0).(func(*bun.DB, *model.User, string) error); ok {
		r0 = rf(db, user, username)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *ServiceInterface) AuthUser(username string, thePassword string) (*model.User, error) {
	ret := _m.Called(username, thePassword)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string, string) *model.User); ok {
		r0 = rf(username, thePassword)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(username, thePassword)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) GetScope(requestedScope string) (string, error) {
	ret := _m.Called(requestedScope)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(requestedScope)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(requestedScope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) Login(client *model.Client, user *model.User, scope string) (*model.AccessToken, *model.RefreshToken, error) {
	ret := _m.Called(client, user, scope)

	var r0 *model.AccessToken
	if rf, ok := ret.Get(0).(func(*model.Client, *model.User, string) *model.AccessToken); ok {
		r0 = rf(client, user, scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AccessToken)
		}
	}

	var r1 *model.RefreshToken
	if rf, ok := ret.Get(1).(func(*model.Client, *model.User, string) *model.RefreshToken); ok {
		r1 = rf(client, user, scope)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.RefreshToken)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*model.Client, *model.User, string) error); ok {
		r2 = rf(client, user, scope)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
func (_m *ServiceInterface) GrantAuthorizationCode(client *model.Client, user *model.User, expiresIn int, redirectURI string, scope string) (*model.AuthorizationCode, error) {
	ret := _m.Called(client, user, expiresIn, redirectURI, scope)

	var r0 *model.AuthorizationCode
	if rf, ok := ret.Get(0).(func(*model.Client, *model.User, int, string, string) *model.AuthorizationCode); ok {
		r0 = rf(client, user, expiresIn, redirectURI, scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AuthorizationCode)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Client, *model.User, int, string, string) error); ok {
		r1 = rf(client, user, expiresIn, redirectURI, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) GrantAccessToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.AccessToken, error) {
	ret := _m.Called(client, user, expiresIn, scope)

	var r0 *model.AccessToken
	if rf, ok := ret.Get(0).(func(*model.Client, *model.User, int, string) *model.AccessToken); ok {
		r0 = rf(client, user, expiresIn, scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Client, *model.User, int, string) error); ok {
		r1 = rf(client, user, expiresIn, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) GetOrCreateRefreshToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.RefreshToken, error) {
	ret := _m.Called(client, user, expiresIn, scope)

	var r0 *model.RefreshToken
	if rf, ok := ret.Get(0).(func(*model.Client, *model.User, int, string) *model.RefreshToken); ok {
		r0 = rf(client, user, expiresIn, scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RefreshToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Client, *model.User, int, string) error); ok {
		r1 = rf(client, user, expiresIn, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) GetValidRefreshToken(token string, client *model.Client) (*model.RefreshToken, error) {
	ret := _m.Called(token, client)

	var r0 *model.RefreshToken
	if rf, ok := ret.Get(0).(func(string, *model.Client) *model.RefreshToken); ok {
		r0 = rf(token, client)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RefreshToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *model.Client) error); ok {
		r1 = rf(token, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) Authenticate(token string) (*model.AccessToken, error) {
	ret := _m.Called(token)

	var r0 *model.AccessToken
	if rf, ok := ret.Get(0).(func(string) *model.AccessToken); ok {
		r0 = rf(token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AccessToken)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) NewIntrospectResponseFromAccessToken(accessToken *model.AccessToken) (*oauth.IntrospectResponse, error) {
	ret := _m.Called(accessToken)

	var r0 *oauth.IntrospectResponse
	if rf, ok := ret.Get(0).(func(*model.AccessToken) *oauth.IntrospectResponse); ok {
		r0 = rf(accessToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.IntrospectResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.AccessToken) error); ok {
		r1 = rf(accessToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
func (_m *ServiceInterface) NewIntrospectResponseFromRefreshToken(refreshToken *model.RefreshToken) (*oauth.IntrospectResponse, error) {
	ret := _m.Called(refreshToken)

	var r0 *oauth.IntrospectResponse
	if rf, ok := ret.Get(0).(func(*model.RefreshToken) *oauth.IntrospectResponse); ok {
		r0 = rf(refreshToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth.IntrospectResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.RefreshToken) error); ok {
		r1 = rf(refreshToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
