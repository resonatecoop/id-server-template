package oauth

import (
	"github.com/resonatecoop/id/config"
	"github.com/uptrace/bun"

	"github.com/resonatecoop/user-api/model"
)

// Service struct keeps objects to avoid passing them around
type Service struct {
	cnf          *config.Config
	db           *bun.DB
	allowedRoles []model.AccessRole
}

// NewService returns a new Service instance
func NewService(cnf *config.Config, db *bun.DB) *Service {
	return &Service{
		cnf:          cnf,
		db:           db,
		allowedRoles: []model.AccessRole{model.AdminRole, model.UserRole},
	}
}

// GetConfig returns config.Config instance
func (s *Service) GetConfig() *config.Config {
	return s.cnf
}

// RestrictToRoles restricts this service to only specified roles
func (s *Service) RestrictToRoles(allowedRoles ...model.AccessRole) {
	s.allowedRoles = allowedRoles
}

// IsRoleAllowed returns true if the role is allowed to use this service
func (s *Service) IsRoleAllowed(role model.AccessRole) bool {
	for _, allowedRole := range s.allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// Close stops any running services
func (s *Service) Close() {}
