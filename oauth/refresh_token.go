package oauth

import (
	"errors"
	"time"

	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrRefreshTokenNotFound ...
	ErrRefreshTokenNotFound = errors.New("Refresh token not found")
	// ErrRefreshTokenExpired ...
	ErrRefreshTokenExpired = errors.New("Refresh token expired")
	// ErrRequestedScopeCannotBeGreater ...
	ErrRequestedScopeCannotBeGreater = errors.New("Requested scope cannot be greater")
)

// GetOrCreateRefreshToken retrieves an existing refresh token, if expired,
// the token gets deleted and new refresh token is created
func (s *Service) GetOrCreateRefreshToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.RefreshToken, error) {
	// Try to fetch an existing refresh token first
	refreshToken := new(model.RefreshToken)
	query := model.RefreshTokenPreload(s.db).Where("client_id = ?", client.ID)
	if user != nil && len([]rune(user.ID)) > 0 {
		query = query.Where("user_id = ?", user.ID)
	} else {
		query = query.Where("user_id IS NULL")
	}
	found := !query.First(refreshToken).RecordNotFound()

	// Check if the token is expired, if found
	var expired bool
	if found {
		expired = time.Now().UTC().After(refreshToken.ExpiresAt)
	}

	// If the refresh token has expired, delete it
	if expired {
		s.db.Unscoped().Delete(refreshToken)
	}

	// Create a new refresh token if it expired or was not found
	if expired || !found {
		refreshToken = model.NewOauthRefreshToken(client, user, expiresIn, scope)
		if err := s.db.Create(refreshToken).Error; err != nil {
			return nil, err
		}
		refreshToken.Client = client
		refreshToken.User = user
	}

	return refreshToken, nil
}

// GetValidRefreshToken returns a valid non expired refresh token
func (s *Service) GetValidRefreshToken(token string, client *model.Client) (*model.RefreshToken, error) {
	// Fetch the refresh token from the database
	refreshToken := new(model.RefreshToken)
	notFound := model.RefreshTokenPreload(s.db).Where("client_id = ?", client.ID).
		Where("token = ?", token).First(refreshToken).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrRefreshTokenNotFound
	}

	// Check the refresh token hasn't expired
	if time.Now().UTC().After(refreshToken.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	return refreshToken, nil
}

// getRefreshTokenScope returns scope for a new refresh token
func (s *Service) getRefreshTokenScope(refreshToken *model.RefreshToken, requestedScope string) (string, error) {
	var (
		scope = refreshToken.Scope // default to the scope originally granted by the resource owner
		err   error
	)

	// If the scope is specified in the request, get the scope string
	if requestedScope != "" {
		scope, err = s.GetScope(requestedScope)
		if err != nil {
			return "", err
		}
	}

	// Requested scope CANNOT include any scope not originally granted
	if !util.SpaceDelimitedStringNotGreater(scope, refreshToken.Scope) {
		return "", ErrRequestedScopeCannotBeGreater
	}

	return scope, nil
}
