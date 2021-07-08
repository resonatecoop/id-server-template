package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/resonatecoop/id/oauth/tokentypes"
	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrInvalidRedirectURI ...
	ErrInvalidRedirectURI = errors.New("Invalid redirect URI")
)

func (s *Service) authorizationCodeGrant(r *http.Request, client *model.Client) (*AccessTokenResponse, error) {
	ctx := context.Background()
	// Fetch the authorization code
	authorizationCode, err := s.getValidAuthorizationCode(
		r.Form.Get("code"),
		r.Form.Get("redirect_uri"),
		client,
	)
	if err != nil {
		return nil, err
	}

	// Log in the user
	accessToken, refreshToken, err := s.Login(
		authorizationCode.Client,
		authorizationCode.User,
		authorizationCode.Scope,
	)
	if err != nil {
		return nil, err
	}

	// Delete the authorization code

	_, err = s.db.NewDelete().
		Model(authorizationCode).
		WherePK().
		ForceDelete().
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	// Create response
	accessTokenResponse, err := NewAccessTokenResponse(
		accessToken,
		refreshToken,
		s.cnf.Oauth.AccessTokenLifetime,
		tokentypes.Bearer,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenResponse, nil
}
