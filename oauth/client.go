package oauth

import (
	"context"
	"errors"
	"strings"

	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

var (
	// ErrClientNotFound ...
	ErrClientNotFound = errors.New("Client not found")
	// ErrInvalidClientSecret ...
	ErrInvalidClientSecret = errors.New("Invalid client secret")
	// ErrClientIDTaken ...
	ErrClientIDTaken = errors.New("Client ID taken")
)

// ClientExists returns true if client exists
func (s *Service) ClientExists(clientID string) bool {
	_, err := s.FindClientByClientID(clientID)
	return err == nil
}

// FindClientByClientID looks up a client by client ID
func (s *Service) FindClientByClientID(clientID string) (*model.Client, error) {
	// Client IDs are case insensitive
	ctx := context.Background()
	client := new(model.Client)

	err := s.db.NewSelect().
		Model(client).
		Where("key = LOWER(?)", clientID).
		Limit(1).
		Scan(ctx)

	// Not Found!
	if err != nil {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// FindClientByRedirectURI looks up a client by redirect URI
func (s *Service) FindClientByApplicationURL(applicationURL string) (*model.Client, error) {
	ctx := context.Background()
	client := new(model.Client)

	err := s.db.NewSelect().
		Model(client).
		Where("application_url = ? AND application_hostname IN (?)", applicationURL, bun.In(s.cnf.Origins)).
		Limit(1).
		Scan(ctx)

	// Not Found!
	if err != nil {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// CreateClient saves a new client to database
func (s *Service) CreateClient(clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*model.Client, error) {
	return s.createClientCommon(s.db, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL)
}

// CreateClientTx saves a new client to database using injected db object
func (s *Service) CreateClientTx(tx *bun.DB, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*model.Client, error) {
	return s.createClientCommon(tx, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL)
}

// AuthClient authenticates client
func (s *Service) AuthClient(clientID, secret string) (*model.Client, error) {
	// Fetch the client
	client, err := s.FindClientByClientID(clientID)
	if err != nil {
		return nil, ErrClientNotFound
	}

	// Verify the secret
	if password.VerifyPassword(client.Secret, secret) != nil {
		return nil, ErrInvalidClientSecret
	}

	return client, nil
}

func (s *Service) createClientCommon(db *bun.DB, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*model.Client, error) {
	ctx := context.Background()
	// Check client ID
	if s.ClientExists(clientID) {
		return nil, ErrClientIDTaken
	}

	// Hash password
	secretHash, err := password.HashPassword(secret)
	if err != nil {
		return nil, err
	}

	client := &model.Client{
		Key:                 strings.ToLower(clientID),
		Secret:              string(secretHash),
		RedirectURI:         util.StringOrNull(redirectURI),
		ApplicationName:     util.StringOrNull(applicationName),
		ApplicationHostname: util.StringOrNull(strings.ToLower(applicationHostname)),
		ApplicationURL:      util.StringOrNull(strings.ToLower(applicationURL)),
	}

	_, err = s.db.NewInsert().Model(client).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}
