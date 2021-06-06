package oauth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/resonatecoop/id/session"
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

	var result sql.Result
	var err error

	result, err = s.db.NewSelect().
		Model(accessToken).
		Where("token = ?", token).
		Limit(1).
		Exec(ctx)

	// Not found
	if result.RowsAffected() < 1 {
		return nil, ErrAccessTokenNotFound
	}

	// Check the access token hasn't expired
	if time.Now().UTC().After(accessToken.ExpiresAt) {
		return nil, ErrAccessTokenExpired
	}

	// Extend refresh token expiration database
	query := s.db.Model(new(model.RefreshToken)).Where("client_id = ?", accessToken.ClientID.String)
	if accessToken.UserID.Valid {
		query = query.Where("user_id = ?", accessToken.UserID.String)
	} else {
		query = query.Where("user_id IS NULL")
	}
	increasedExpiresAt := gorm.NowFunc().Add(
		time.Duration(s.cnf.Oauth.RefreshTokenLifetime) * time.Second,
	)
	if err := query.UpdateColumn("expires_at", increasedExpiresAt).Error; err != nil {
		return nil, err
	}

	return accessToken, nil
}

// ClearUserTokens deletes the user's access and refresh tokens associated with this client id
func (s *Service) ClearUserTokens(userSession *session.UserSession) {
	// Clear all refresh tokens with user_id and client_id
	refreshToken := new(model.RefreshToken)
	found := !model.RefreshTokenPreload(s.db).Where("token = ?", userSession.RefreshToken).First(refreshToken).RecordNotFound()
	if found {
		s.db.Unscoped().Where("client_id = ? AND user_id = ?", refreshToken.ClientID, refreshToken.UserID).Delete(model.RefreshToken{})
	}

	// Clear all access tokens with user_id and client_id
	accessToken := new(model.AccessToken)
	found = !model.AccessTokenPreload(s.db).Where("token = ?", userSession.AccessToken).First(accessToken).RecordNotFound()
	if found {
		s.db.Unscoped().Where("client_id = ? AND user_id = ?", accessToken.ClientID, accessToken.UserID).Delete(model.AccessToken{})
	}
}
