package oauth

import (
	"context"
	"time"

	"github.com/resonatecoop/user-api/model"
)

// GrantAccessToken deletes old tokens and grants a new access token
func (s *Service) GrantAccessToken(client *model.Client, user *model.User, expiresIn int, scope string) (*model.AccessToken, error) {
	// Begin a transaction
	tx, err := s.db.Begin()
	ctx := context.TODO()

	//var result Sql.result

	if err != nil {
		return nil, err
	}

	access_token := new(model.AccessToken)

	// Delete expired access tokens
	if user != nil && len(user.ID.String()) > 0 {
		_, err = tx.NewDelete().
			Model(access_token).
			Where("user_id = ?", user.ID).
			Where("client_id = ?", client.ID).
			Where("expires_at <= ?", time.Now()).
			Exec(ctx)
	} else {
		_, err = tx.NewDelete().
			Model(access_token).
			Where("user_id IS NULL").
			Where("client_id = ?", client.ID).
			Where("expires_at <= ?", time.Now()).
			Exec(ctx)
	}

	_, err = tx.NewDelete().
		Model(access_token).
		Where("expires_at <= ?", time.Now()).
		Exec(ctx)
	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}

	// Create a new access token
	accessToken := model.NewOauthAccessToken(client, user, expiresIn, scope)

	_, err = tx.NewInsert().
		Model(accessToken).
		Exec(ctx)

	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}
	accessToken.Client = client
	accessToken.User = user

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback() // rollback the transaction
		return nil, err
	}

	return accessToken, nil
}
