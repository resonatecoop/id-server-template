package oauth

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/resonatecoop/id/oauth/tokentypes"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
)

const (
	// AccessTokenHint ...
	AccessTokenHint = "access_token"
	// RefreshTokenHint ...
	RefreshTokenHint = "refresh_token"
)

var (
	// ErrTokenMissing ...
	ErrTokenMissing = errors.New("Token missing")
	// ErrTokenHintInvalid ...
	ErrTokenHintInvalid = errors.New("Invalid token hint")
)

func (s *Service) introspectToken(r *http.Request, client *model.Client) (*IntrospectResponse, error) {
	// Parse the form so r.Form becomes available
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	// Get token from the query
	token := r.Form.Get("token")
	if token == "" {
		return nil, ErrTokenMissing
	}

	// Get token type hint from the query
	tokenTypeHint := r.Form.Get("token_type_hint")

	// Default to access token hint
	if tokenTypeHint == "" {
		tokenTypeHint = AccessTokenHint
	}

	switch tokenTypeHint {
	case AccessTokenHint:
		accessToken, err := s.Authenticate(token)
		if err != nil {
			return nil, err
		}
		return s.NewIntrospectResponseFromAccessToken(accessToken)
	case RefreshTokenHint:
		refreshToken, err := s.GetValidRefreshToken(token, client)
		if err != nil {
			return nil, err
		}
		return s.NewIntrospectResponseFromRefreshToken(refreshToken)
	default:
		return nil, ErrTokenHintInvalid
	}
}

// NewIntrospectResponseFromAccessToken ...
func (s *Service) NewIntrospectResponseFromAccessToken(accessToken *model.AccessToken) (*IntrospectResponse, error) {
	ctx := context.Background()
	var introspectResponse = &IntrospectResponse{
		Active:    true,
		Scope:     accessToken.Scope,
		TokenType: tokentypes.Bearer,
		ExpiresAt: int(accessToken.ExpiresAt.Unix()),
	}

	if util.IsValidUUID(accessToken.ClientID.String()) && accessToken.ClientID != uuid.Nil {
		client := new(model.Client)
		err := s.db.NewSelect().
			Model(client).
			Column("key").
			Where("id = ?", accessToken.ClientID.String()).
			Limit(1).
			Scan(ctx)
		if err != nil {
			return nil, ErrClientNotFound
		}
		introspectResponse.ClientID = client.Key
	}

	if util.IsValidUUID(accessToken.UserID.String()) && accessToken.UserID != uuid.Nil {
		user := new(model.User)
		err := s.db.NewSelect().
			Model(user).
			Column("username").
			Where("id = ?", accessToken.UserID.String()).
			Limit(1).
			Scan(ctx)
		if err != nil {
			return nil, ErrUserNotFound
		}

		introspectResponse.Username = user.Username
		introspectResponse.UserID = accessToken.UserID.String()
	}

	return introspectResponse, nil
}

// NewIntrospectResponseFromRefreshToken ...
func (s *Service) NewIntrospectResponseFromRefreshToken(refreshToken *model.RefreshToken) (*IntrospectResponse, error) {
	ctx := context.Background()
	var introspectResponse = &IntrospectResponse{
		Active:    true,
		Scope:     refreshToken.Scope,
		TokenType: tokentypes.Bearer,
		ExpiresAt: int(refreshToken.ExpiresAt.Unix()),
	}

	if util.IsValidUUID(refreshToken.ClientID.String()) && refreshToken.ClientID != uuid.Nil {
		client := new(model.Client)
		err := s.db.NewSelect().
			Model(client).
			Column("key").
			Where("id = ?", refreshToken.ClientID.String()).
			Limit(1).
			Scan(ctx)
		if err != nil {
			return nil, ErrClientNotFound
		}
		introspectResponse.ClientID = client.Key
	}

	if util.IsValidUUID(refreshToken.UserID.String()) && refreshToken.UserID != uuid.Nil {
		user := new(model.User)
		err := s.db.NewSelect().
			Model(user).
			Column("username").
			Where("id = ?", refreshToken.UserID.String()).
			Limit(1).
			Scan(ctx)
		if err != nil {
			return nil, ErrUserNotFound
		}

		introspectResponse.Username = user.Username
		introspectResponse.UserID = refreshToken.UserID.String()
	}

	return introspectResponse, nil
}
