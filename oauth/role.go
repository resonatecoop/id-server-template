package oauth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrRoleNotFound ...
	ErrRoleNotFound = errors.New("Role not found")
)

// FindRoleByID looks up a role by ID and returns it
func (s *Service) FindRoleByID(id int32) (*model.AccessRole, error) {
	role := new(model.Role)
	err := s.db.NewSelect().Model(role).Where("id = ?", id).Scan(context.Background())

	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	return (*model.AccessRole)(&role.ID), nil
}
