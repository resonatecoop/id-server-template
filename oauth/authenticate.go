package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrAccessTokenNotFound ...
	ErrAccessTokenNotFound = errors.New("Access token not found")
	// ErrAccessTokenExpired ...
	ErrAccessTokenExpired = errors.New("Access token expired")
)

// Authenticate checks the access token is valid
func (s *Service) Authenticate(token string) (*model.AccessToken, error) {
	// Fetch the access token from the database
	ctx := context.Background()
	accessToken := new(model.AccessToken)

	err := s.db.NewSelect().
		Model(accessToken).
		Where("token = ?", token).
		Limit(1).
		Scan(ctx)

	// Not found
	if err != nil {
		return nil, ErrAccessTokenNotFound
	}

	// Check the access token hasn't expired
	if time.Now().UTC().After(accessToken.ExpiresAt) {
		return nil, ErrAccessTokenExpired
	}

	// Extend refresh token expiration database

	increasedExpiresAt := time.Now().Add(
		time.Duration(s.cnf.Oauth.RefreshTokenLifetime) * time.Second,
	)

	//var res sql.Result

	//	err = GetOrCreateRefreshToken

	if util.IsValidUUID(accessToken.UserID.String()) && accessToken.UserID != uuid.Nil {
		_, err = s.db.NewUpdate().
			Model(new(model.RefreshToken)).
			Set("expires_at = ?", increasedExpiresAt).
			Set("updated_at = ?", time.Now().UTC()).
			Where("client_id = ?", accessToken.ClientID.String()).
			Where("user_id = ?", accessToken.UserID.String()).
			Exec(ctx)
	} else {
		_, err = s.db.NewUpdate().
			Model(new(model.RefreshToken)).
			Set("expires_at = ?", increasedExpiresAt).
			Set("updated_at = ?", time.Now().UTC()).
			Where("client_id = ?", accessToken.ClientID.String()).
			Where("user_id = uuid_nil()").
			Exec(ctx)
	}

	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

// ClearUserTokens deletes the user's access and refresh tokens associated with this client id
func (s *Service) ClearUserTokens(userSession *session.UserSession) {
	// Clear all refresh tokens with user_id and client_id
	ctx := context.Background()
	refreshToken := new(model.RefreshToken)

	err := s.db.NewSelect().
		Model(refreshToken).
		Where("token = ?", userSession.RefreshToken).
		Limit(1).
		Scan(ctx)

	// Found
	if err == nil {
		_, err = s.db.NewDelete().
			Model(refreshToken).
			Where("client_id = ? AND user_id = ?", refreshToken.ClientID, refreshToken.UserID).
			Exec(ctx)
	}

	// Clear all access tokens with user_id and client_id
	accessToken := new(model.AccessToken)

	err = s.db.NewSelect().
		Model(accessToken).
		Where("token = ?", userSession.AccessToken).
		Limit(1).
		Scan(ctx)

	// Found
	if err == nil {
		_, err = s.db.NewDelete().
			Model(accessToken).
			Where("client_id = ? AND user_id = ?", accessToken.ClientID, accessToken.UserID).
			Exec(ctx)
	}
}
