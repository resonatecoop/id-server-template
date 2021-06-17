package oauth_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenCreatesNew() {
	var (
		ctx          context.Context
		refreshToken *model.RefreshToken
		err          error
		tokens       []*model.RefreshToken
	)

	// Since there is no user specific token,
	// a new one should be created and returned
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	ctx = context.Background()

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be just one token now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].ClientID)

		// User ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID, tokens[0].UserID)
	}

	// Valid user specific token exists, new one should NOT be created
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be just one token now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].ClientID)

		// User ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID, tokens[0].UserID)
	}

	// Since there is no client only token,
	// a new one should be created and returned
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be 2 tokens
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].ClientID)

		// User ID should be nil
		assert.Equal(suite.T(), tokens[1].UserID, uuid.Nil)
	}

	// Valid client only token exists, new one should NOT be created
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be 2 tokens
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].ClientID)

		// User ID should be nil
		assert.Equal(suite.T(), tokens[1].UserID, uuid.Nil)
	}
}

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenReturnsExisting() {
	var (
		ctx          context.Context
		refreshToken *model.RefreshToken
		err          error
		tokens       []*model.RefreshToken
	)

	// Insert an access token without a user
	refreshToken = &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
	}

	ctx = context.Background()

	_, err = suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	assert.Nil(suite.T(), err)

	// Since the current client only token is valid, this should just return it
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct refresh token should be returned
	if assert.NotNil(suite.T(), refreshToken) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be just one token right now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)
		assert.Equal(suite.T(), "test_token", refreshToken.Token)
		assert.Equal(suite.T(), "test_token", tokens[0].Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].ClientID)

		// User ID should be nil
		assert.Equal(suite.T(), tokens[0].UserID, uuid.Nil)
	}

	// Insert an access token with a user

	refreshToken = &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token2",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
	}

	_, err = suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	//Insert succeeded
	assert.Nil(suite.T(), err)

	// Since the current user specific only token is valid,
	// this should just return it
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct refresh token should be returned
	if assert.NotNil(suite.T(), refreshToken) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)
		assert.Equal(suite.T(), "test_token2", refreshToken.Token)
		assert.Equal(suite.T(), "test_token2", tokens[1].Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].ClientID)

		// User ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID, tokens[1].UserID)
	}
}

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenDeletesExpired() {
	var (
		ctx          context.Context
		refreshToken *model.RefreshToken
		err          error
		tokens       []*model.RefreshToken
	)

	// Insert an expired client only test refresh token
	refreshToken = &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
		ClientID:  suite.clients[0].ID,
	}

	ctx = context.Background()

	_, err = suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	// No issue inserting
	assert.Nil(suite.T(), err)

	// Since the current client only token is expired,
	// this should delete it and create and return a new one
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be just one token right now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", tokens[0].Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].ClientID)

		// User ID should be nil
		assert.Equal(suite.T(), tokens[0].UserID, uuid.Nil)
	}

	// Insert an expired user specific test refresh token
	refreshToken = &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
	}

	suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	assert.Nil(suite.T(), err)

	// Since the current user specific token is expired,
	// this should delete it and create and return a new one
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &tokens)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", tokens[1].Token)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].ClientID)

		// User ID should be set
		assert.True(suite.T(), util.IsValidUUID(tokens[1].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID, tokens[1].UserID)
	}
}

func (suite *OauthTestSuite) TestGetValidRefreshToken() {
	var (
		ctx          context.Context
		refreshToken *model.RefreshToken
		err          error
	)

	// Insert some test refresh tokens
	testRefreshTokens := []*model.RefreshToken{
		// Expired test refresh token
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_expired_token",
			ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		// Refresh token
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
	}

	ctx = context.Background()

	for _, testRefreshToken := range testRefreshTokens {
		_, err = suite.db.NewInsert().
			Model(testRefreshToken).
			Exec(ctx)

		assert.Nil(suite.T(), err)
	}

	// Test passing an empty token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"",               // refresh token
		suite.clients[0], // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenNotFound, err)
	}

	// Test passing a bogus token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"bogus",          // refresh token
		suite.clients[0], // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenNotFound, err)
	}

	// Test passing an expired token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"test_expired_token", // refresh token
		suite.clients[0],     // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenExpired, err)
	}

	// Test passing a valid token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"test_token",     // refresh token
		suite.clients[0], // client
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct refresh token object should be returned
	assert.NotNil(suite.T(), refreshToken)
	assert.Equal(suite.T(), "test_token", refreshToken.Token)
}
