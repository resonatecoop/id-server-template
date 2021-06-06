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
func (s *Service) FindRoleByID(id int8) (*model.Role, error) {
	role := new(model.Role)
	_, err := s.db.NewSelect().Model(role).Where("id = ?", id).Scan(context.TODO())

	if err == sql.ErrNoRows {
		return nil, ErrRoleNotFound
	}
	return role, nil
}
