package webhook

import (
	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/oauth"
	"github.com/uptrace/bun"
)

// Service struct keeps variables for reuse
type Service struct {
	cnf          *config.Config
	db           *bun.DB
	oauthService oauth.ServiceInterface
}

// NewService returns a new Service instance
func NewService(cnf *config.Config, db *bun.DB, oauthService oauth.ServiceInterface) *Service {
	return &Service{
		cnf:          cnf,
		db:           db,
		oauthService: oauthService,
	}
}

// GetConfig returns config.Config instance
func (s *Service) GetConfig() *config.Config {
	return s.cnf
}

// GetOauthService returns oauth.Service instance
func (s *Service) GetOauthService() oauth.ServiceInterface {
	return s.oauthService
}

// Close stops any running services
func (s *Service) Close() {}
