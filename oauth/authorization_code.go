package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrAuthorizationCodeNotFound ...
	ErrAuthorizationCodeNotFound = errors.New("Authorization code not found")
	// ErrAuthorizationCodeExpired ...
	ErrAuthorizationCodeExpired = errors.New("Authorization code expired")
)

// GrantAuthorizationCode grants a new authorization code
func (s *Service) GrantAuthorizationCode(client *model.Client, user *model.User, expiresIn int, redirectURI, scope string) (*model.AuthorizationCode, error) {
	// Create a new authorization code
	authorizationCode := model.NewOauthAuthorizationCode(client, user, expiresIn, redirectURI, scope)

	ctx := context.Background()

	_, err := s.db.NewInsert().Model(authorizationCode).Exec(ctx)
	if err != nil {
		return nil, err
	}
	authorizationCode.Client = client
	authorizationCode.User = user

	return authorizationCode, nil
}

// getValidAuthorizationCode returns a valid non expired authorization code
func (s *Service) getValidAuthorizationCode(code, redirectURI string, client *model.Client) (*model.AuthorizationCode, error) {
	// Fetch the auth code from the database
	ctx := context.Background()
	authorizationCode := new(model.AuthorizationCode)

	err := s.db.NewSelect().
		Model(authorizationCode).
		Where("client_id = ?", client.ID).
		Where("code = ?", code).
		Limit(1).
		Scan(ctx)

	// Not Found!
	if err != nil {
		return nil, ErrAuthorizationCodeNotFound
	}

	authorizationCode.Client = client

	user := new(model.User)

	err = s.db.NewSelect().
		Model(user).
		Where("id = ?", authorizationCode.UserID).
		Limit(1).
		Scan(ctx)

	// Not Found!
	if err != nil {
		return nil, errors.New("corresponding user for authorization code not found")
	}

	authorizationCode.User = user

	// Redirect URI must match if it was used to obtain the authorization code
	if redirectURI != authorizationCode.RedirectURI.String {
		return nil, ErrInvalidRedirectURI
	}

	// Check the authorization code hasn't expired
	if time.Now().After(authorizationCode.ExpiresAt) {
		return nil, ErrAuthorizationCodeExpired
	}

	return authorizationCode, nil
}
