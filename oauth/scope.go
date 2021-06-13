package oauth

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/resonatecoop/user-api/model"
	"github.com/uptrace/bun"
)

var (
	// ErrInvalidScope ...
	ErrInvalidScope = errors.New("Invalid scope")
)

// GetScope takes a requested scope and, if it's empty, returns the default
// scope, if not empty, it validates the requested scope
func (s *Service) GetScope(requestedScope string) (string, error) {
	// Return the default scope if the requested scope is empty
	if requestedScope == "" {
		return s.GetDefaultScope(), nil
	}

	// If the requested scope exists in the database, return it
	if s.ScopeExists(requestedScope) {
		return requestedScope, nil
	}

	// Otherwise return error
	return "", ErrInvalidScope
}

// GetDefaultScope returns the default scope
func (s *Service) GetDefaultScope() string {
	ctx := context.Background()
	// Fetch default scopes
	var scopes []string

	rows, err := s.db.NewSelect().
		Model((*model.Scope)(nil)).
		Where("is_default = ?", true).
		Rows(ctx)

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		scope := new(model.Scope)
		if err := s.db.ScanRow(ctx, rows, scope); err != nil {
			panic(err)
		}
		scopes = append(scopes, scope.Name)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	// s.db.NewSelect().Model(new(model.Scope)).Where("is_default = ?", true).Pluck("scope", &scopes)

	// Sort the scopes alphabetically
	sort.Strings(scopes)

	// Return space delimited scope string
	return strings.Join(scopes, " ")
}

// ScopeExists checks if a scope exists
func (s *Service) ScopeExists(requestedScope string) bool {
	ctx := context.Background()
	// Split the requested scope string
	scopes := strings.Split(requestedScope, " ")

	var available_scopes []model.Scope

	// Count how many of requested scopes exist in the database
	count, _ := s.db.NewSelect().
		Model(&available_scopes).
		Where("name IN (?)", bun.In(scopes)).
		ScanAndCount(ctx)

	// Return true only if all requested scopes found
	return count == len(scopes)
}
