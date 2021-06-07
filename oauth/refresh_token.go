package oauth

import (
	"context"
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
	ctx := context.Background()
	// Try to fetch an existing refresh token first
	refreshToken := new(model.RefreshToken)
	// query := model.RefreshTokenPreload(s.db).Where("client_id = ?", client.ID)

	var err error

	if user != nil && len(user.ID.String()) > 0 {
		err = s.db.NewSelect().
			Model(refreshToken).
			Where("client_id = ?", client.ID).
			Where("user_id = ?", user.ID).
			Limit(1).
			Scan(ctx)
	} else {
		err = s.db.NewSelect().
			Model(refreshToken).
			Where("client_id = ?", client.ID).
			Where("user_id IS NULL").
			Limit(1).
			Scan(ctx)
	}

	// Check if the token is expired, if found
	var expired bool
	if err == nil {
		expired = time.Now().UTC().After(refreshToken.ExpiresAt)
	}

	// If the refresh token has expired, delete it
	if expired {
		_, dberr := s.db.NewDelete().
			Model(refreshToken).
			Exec(ctx)
		//		s.db.Unscoped().Delete(refreshToken)
	}

	// Create a new refresh token if it expired or was not found
	if expired || (err != nil) {
		refreshToken = model.NewOauthRefreshToken(client, user, expiresIn, scope)

		_, err = s.db.NewInsert().
			Model(refreshToken).
			Exec(ctx)

		if err != nil {
			return nil, err
		}

		refreshToken.Client = client
		refreshToken.User = user
	}

	return refreshToken, nil
}

// GetValidRefreshToken returns a valid non expired refresh token
func (s *Service) GetValidRefreshToken(token string, client *model.Client) (*model.RefreshToken, error) {
	ctx := context.Background()
	// Fetch the refresh token from the database
	refreshToken := new(model.RefreshToken)

	err := s.db.NewSelect().
		Model(refreshToken).
		Where("client_id = ?", client.ID).
		Where("token = ?", token).
		Limit(1).
		Scan(ctx)

	// Not found
	if err != nil {
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
