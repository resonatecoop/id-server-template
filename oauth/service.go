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
	allowedRoles []int32
}

// NewService returns a new Service instance
func NewService(cnf *config.Config, db *bun.DB) *Service {
	return &Service{
		cnf:          cnf,
		db:           db,
		allowedRoles: []int32{int32(model.SuperAdminRole), int32(model.AdminRole), int32(model.TenantAdminRole), int32(model.LabelRole), int32(model.ArtistRole), int32(model.UserRole)},
	}
}

// GetConfig returns config.Config instance
func (s *Service) GetConfig() *config.Config {
	return s.cnf
}

// RestrictToRoles restricts this service to only specified roles
func (s *Service) RestrictToRoles(allowedRoles ...int32) {
	s.allowedRoles = allowedRoles
	// for i := range s.allowedRoles {
	// 	s.allowedRoles[i] = model.AccessRole(allowedRoles[i])
	// }
}

// IsRoleAllowed returns true if the role is allowed to use this service
func (s *Service) IsRoleAllowed(role int32) bool {
	for _, allowedRole := range s.allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// Close stops any running services
func (s *Service) Close() {}
