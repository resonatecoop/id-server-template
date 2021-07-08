package health

import (
	"github.com/uptrace/bun"
)

// Service struct keeps db object to avoid passing it around
type Service struct {
	db *bun.DB
}

// NewService returns a new Service instance
func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

// Close stops any running services
func (s *Service) Close() {}
