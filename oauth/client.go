package oauth

import (
	"errors"
	"strings"
	"time"

	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
	uuid "github.com/satori/go.uuid"
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
	client := new(model.Client)
	notFound := s.db.Where("key = LOWER(?)", clientID).
		First(client).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// FindClientByRedirectURI looks up a client by redirect URI
func (s *Service) FindClientByApplicationURL(applicationURL string) (*model.Client, error) {
	client := new(model.Client)
	notFound := s.db.Where("application_url = ? AND application_hostname IN (?)", applicationURL, s.cnf.Origins).
		First(client).RecordNotFound()

	// Not found
	if notFound {
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
		MybunModel: model.MybunModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Key:                 strings.ToLower(clientID),
		Secret:              string(secretHash),
		RedirectURI:         util.StringOrNull(redirectURI),
		ApplicationName:     util.StringOrNull(applicationName),
		ApplicationHostname: util.StringOrNull(strings.ToLower(applicationHostname)),
		ApplicationURL:      util.StringOrNull(strings.ToLower(applicationURL)),
	}
	if err := db.Create(client).Error; err != nil {
		return nil, err
	}
	return client, nil
}
