package oauth

import (
	"context"
	"errors"
	"strings"

	"github.com/resonatecoop/user-api/model"
)

// Login creates an access token and refresh token for a user (logs him/her in)
func (s *Service) Login(client *model.Client, user *model.User, scope string) (*model.AccessToken, *model.RefreshToken, error) {

	if user == nil {
		return nil, nil, errors.New("valid user must be supplied")
	}

	// Return error if user's role is not allowed to use this service
	if !s.IsRoleAllowed(user.RoleID) {
		// For security reasons, return a general error message
		return nil, nil, ErrInvalidUsernameOrPassword
	}

	scope, err := s.updateUserScopeWithRole(user, scope)

	if err != nil {
		return nil, nil, err
	}

	// Create a new access token
	accessToken, err := s.GrantAccessToken(
		client,
		user,
		s.cnf.Oauth.AccessTokenLifetime, // expires in
		scope,
	)
	if err != nil {
		return nil, nil, err
	}

	// Create or retrieve a refresh token
	refreshToken, err := s.GetOrCreateRefreshToken(
		client,
		user,
		s.cnf.Oauth.RefreshTokenLifetime, // expires in
		scope,
	)
	if err != nil {
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) updateUserScopeWithRole(user *model.User, scope string) (string, error) {

	ctx := context.Background()

	scopes := strings.Split(scope, " ")

	if scopes[0] != "read" && scopes[0] != "read_write" {
		return "", errors.New("invalid scope format")
	}

	scopeRole := new(model.Role)

	err := s.db.NewSelect().
		Model(scopeRole).
		Where("id = ?", user.RoleID).
		Scan(ctx)

	if err != nil {
		return "", errors.New("problem determining role from user record")
	}

	scope = scopes[0] + " " + scopeRole.Name

	return scope, nil
}
