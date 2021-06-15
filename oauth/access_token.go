package oauth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resonatecoop/user-api/model"
)

// GrantAccessToken deletes old tokens and grants a new access token
func (s *Service) GrantAccessToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.AccessToken, error) {
	// Begin a transaction
	tx, err := s.db.Begin()
	ctx := context.Background()

	//var result Sql.result

	if err != nil {
		return nil, err
	}

	accessToken := new(model.AccessToken)

	// Delete expired access tokens
	if user != nil && user.ID != uuid.Nil {
		_, err = tx.NewDelete().
			Model(accessToken).
			Where("user_id = ?", user.ID).
			Where("client_id = ?", client.ID).
			Where("expires_at <= ?", time.Now()).
			Exec(ctx)
	} else {
		_, err = tx.NewDelete().
			Model(accessToken).
			Where("user_id = uuid_nil()").
			Where("client_id = ?", client.ID).
			Where("expires_at <= ?", time.Now()).
			Exec(ctx)
	}

	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}

	// Create a new access token
	accessToken = model.NewOauthAccessToken(client, user, expiresIn, scope)

	_, err = tx.NewInsert().
		Model(accessToken).
		Exec(ctx)

	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}
	accessToken.ClientID = client.ID

	if user == nil {
		accessToken.UserID = uuid.Nil
	} else {
		accessToken.UserID = user.ID
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}

	return accessToken, nil
}
