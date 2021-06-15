package oauth_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/resonatecoop/id/util"

	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestGrantAccessToken() {
	var (
		ctx          context.Context
		accessToken  *model.AccessToken
		err          error
		accessTokens []model.AccessToken
	)

	ctx = context.Background()

	// Grant a client only access token
	accessToken, err = suite.service.GrantAccessToken(
		suite.clients[0],       // client
		nil,                    // user
		3600,                   // expires in
		"scope doesn't matter", // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct access token object should be returned
	if assert.NotNil(suite.T(), accessToken) {
		// Fetch all access tokens
		//	model.AccessTokenPreload(suite.db).Order("created_at").Find(&tokens)

		err = suite.db.NewSelect().Model(&accessTokens).
			Column("access_token.*").
			Relation("Client").
			OrderExpr("created_at").
			Scan(ctx)

		// There should be no error
		assert.Nil(suite.T(), err)

		// There should be just one right now
		assert.Equal(suite.T(), 1, len(accessTokens))

		// And the token should match the one returned by the grant method
		assert.Equal(suite.T(), accessTokens[0].Token, accessToken.Token)

		// Client id should be set
		assert.True(suite.T(), util.IsValidUUID(accessTokens[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID.String(), accessTokens[0].ClientID.String())

		// User id should be nil
		assert.Equal(suite.T(), accessTokens[0].UserID, uuid.Nil)
	}

	// Grant a user specific access token
	accessToken, err = suite.service.GrantAccessToken(
		suite.clients[0],       // client
		suite.users[0],         // user
		3600,                   // expires in
		"scope doesn't matter", // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct access token object should be returned
	if assert.NotNil(suite.T(), accessToken) {
		// Fetch all access tokens

		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM access_tokens ORDER BY created_at")
		if err != nil {
			panic(err)
		}
		err = suite.db.ScanRows(ctx, rows, &accessTokens)

		// There should be no error
		assert.Nil(suite.T(), err)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(accessTokens))

		// And the second token should match the one returned by the grant method
		assert.Equal(suite.T(), accessTokens[1].Token, accessToken.Token)

		// Client id should be set
		assert.True(suite.T(), util.IsValidUUID(accessTokens[1].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID.String(), accessTokens[1].ClientID.String())

		// User id should be set
		assert.True(suite.T(), util.IsValidUUID(accessTokens[1].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID.String(), accessTokens[1].UserID.String())
	}
}

func (suite *OauthTestSuite) TestGrantAccessTokenDeletesExpiredTokens() {
	var (
		ctx              context.Context
		accessToken      *model.AccessToken
		testAccessTokens = []*model.AccessToken{
			// Expired access token with a user
			{
				Token:     "test_token_1",
				ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
				ClientID:  suite.clients[0].ID,
				UserID:    suite.users[0].ID,
			},
			// Expired access token without a user
			{
				Token:     "test_token_2",
				ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
				ClientID:  suite.clients[0].ID,
			},
			// Access token with a user
			{
				Token:     "test_token_3",
				ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
				ClientID:  suite.clients[0].ID,
				UserID:    suite.users[0].ID,
			},
			// Access token without a user
			{
				Token:     "test_token_4",
				ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
				ClientID:  suite.clients[0].ID,
			},
		}
		err            error
		existingTokens []string
	)
	ctx = context.Background()
	// Insert test access tokens
	for _, testAccessToken := range testAccessTokens {
		_, err = suite.db.NewInsert().
			Model(testAccessToken).
			Exec(ctx)
		assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// This should only delete test_token_1
	_, err = suite.service.GrantAccessToken(
		suite.clients[0],       // client
		suite.users[0],         // user
		3600,                   // expires in
		"scope doesn't matter", // scope
	)
	assert.NoError(suite.T(), err)

	accessToken = new(model.AccessToken)
	// Check the test_token_1 was deleted
	err = suite.db.NewSelect().
		Model(accessToken).
		Where("token = ?", "test_token_1").
		Limit(1).
		Scan(ctx)

	assert.NotNil(suite.T(), err)

	// Check the other two tokens are still around
	existingTokens = []string{
		"test_token_2",
		//"test_token_3",
		"test_token_4",
	}

	for _, token := range existingTokens {

		err = suite.db.NewSelect().
			Model(accessToken).
			Where("token = ?", token).
			Limit(1).
			Scan(ctx)

		assert.Nil(suite.T(), err)
	}

	// This should only delete test_token_2
	_, err = suite.service.GrantAccessToken(
		suite.clients[0],       // client
		nil,                    // user
		3600,                   // expires in
		"scope doesn't matter", // scope
	)
	assert.NoError(suite.T(), err)

	// Check the test_token_2 was deleted
	err = suite.db.NewSelect().
		Model(accessToken).
		Where("token = ?", "test_token_2").
		Limit(1).
		Scan(ctx)

	assert.NotNil(suite.T(), err)

	// Check that last two tokens are still around
	existingTokens = []string{
		"test_token_3",
		"test_token_4",
	}
	for _, token := range existingTokens {
		err = suite.db.NewSelect().
			Model(accessToken).
			Where("token = ?", token).
			Limit(1).
			Scan(ctx)

		assert.Nil(suite.T(), err)
	}
}
