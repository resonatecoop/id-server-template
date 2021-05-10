package oauth

import (
	"errors"
	"strings"

	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/RichardKnop/go-oauth2-server/util/password"
	"github.com/jinzhu/gorm"
)

var (
	// ErrClientNotFound ...
	ErrClientNotFound = errors.New("Client not found")
	// ErrInvalidClientSecret ...
	ErrInvalidClientSecret = errors.New("Invalid client secret")
	// ErrClientIDTaken ...
	ErrClientIDTaken = errors.New("Client ID taken")
	// ErrApplicationHostnameTaken ...
	ErrApplicationHostnameTaken = errors.New("Application hostname is taken")
)

var apps []models.OauthClient

// ClientExists returns true if client exists
func (s *Service) ClientExists(clientID string) bool {
	_, err := s.FindClientByClientID(clientID)
	return err == nil
}

func (s *Service) HostnameTaken(hostname string) bool {
	_, err := s.FindClientByHostname(hostname)
	return err == nil
}

// FindClientByHostname looks up a client by application hostname
func (s *Service) FindClientByHostname(applicationHostname string) (*models.OauthClient, error) {
	// Client IDs are case insensitive
	client := new(models.OauthClient)
	notFound := s.db.Where("application_hostname = LOWER(?)", applicationHostname).
		First(client).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// FindClientByClientID looks up a client by client ID
func (s *Service) FindClientByClientID(clientID string) (*models.OauthClient, error) {
	// Client IDs are case insensitive
	client := new(models.OauthClient)
	notFound := s.db.Where("key = LOWER(?)", clientID).
		First(client).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// DeleteClient
func (s *Service) DeleteClient(clientID string, user *models.OauthUser) error {
	client, err := s.FindClientByClientID(clientID)

	if err != nil {
		return err
	}

	return s.db.Unscoped().Where("user_id = ?", user.ID).Delete(client).Error
}

func (s *Service) FindClientsByUserId(oauthUser *models.OauthUser) ([]models.OauthClient, error) {
	notFound := s.db.
		Where("user_id = ?", oauthUser.ID).
		Select("ID,created_at,updated_at,key,user_id,redirect_uri,application_hostname,application_url,application_name,active").
		Order("created_at desc").
		Find(&apps).
		RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrClientNotFound
	}

	return apps, nil
}

// FindClientByRedirectURI looks up a client by redirect URI
func (s *Service) FindClientByApplicationURL(applicationURL string) (*models.OauthClient, error) {
	client := new(models.OauthClient)
	notFound := s.db.Where("application_url = ? AND application_hostname IN (?)", applicationURL, s.cnf.Origins).
		First(client).RecordNotFound()

	// Not found
	if notFound {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// CreateClient saves a new client to database
func (s *Service) CreateClient(oauthUser *models.OauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*models.OauthClient, error) {
	return s.createClientCommon(s.db, oauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL)
}

// CreateClientTx saves a new client to database using injected db object
func (s *Service) CreateClientTx(tx *gorm.DB, oauthUser *models.OauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*models.OauthClient, error) {
	return s.createClientCommon(tx, oauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL)
}

// AuthClient authenticates client
func (s *Service) AuthClient(clientID, secret string) (*models.OauthClient, error) {
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

func (s *Service) createClientCommon(db *gorm.DB, oauthUser *models.OauthUser, clientID, secret, redirectURI, applicationName, applicationHostname, applicationURL string) (*models.OauthClient, error) {
	// Check client ID
	if s.ClientExists(clientID) {
		return nil, ErrClientIDTaken
	}

	// Check the application hostname
	// may have to allow this
	if s.HostnameTaken(applicationHostname) {
		return nil, ErrApplicationHostnameTaken
	}

	// Hash password
	secretHash, err := password.HashPassword(secret)
	if err != nil {
		return nil, err
	}

	client := models.NewOauthClient(
		oauthUser,
		strings.ToLower(clientID),
		string(secretHash),
		redirectURI,
		applicationName,
		applicationHostname,
		applicationURL,
		false, // active
	)

	if err := db.Create(client).Error; err != nil {
		return nil, err
	}
	return client, nil
}
