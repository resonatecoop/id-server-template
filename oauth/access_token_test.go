package oauth

import (
	"database/sql/driver"
	"log"
	"time"

	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestGrantAccessToken() {
	var accessToken *AccessToken
	var err error
	var tokens []*AccessToken
	var v driver.Value

	// Grant a client only access token
	accessToken, err = s.GrantAccessToken(
		suite.client,
		nil,
		"doesn't matter",
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct access token object should be returned
	if assert.NotNil(suite.T(), accessToken) {
		// Fetch all access tokens
		s.db.Preload("Client").Preload("User").Find(&tokens)

		// There should be just one right now
		assert.Equal(suite.T(), 1, len(tokens))

		// And the token should match the one returned by the grant method
		assert.Equal(suite.T(), tokens[0].Token, accessToken.Token)

		// Client id should be set
		assert.True(suite.T(), tokens[0].ClientID.Valid)
		v, err = tokens[0].ClientID.Value()
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), int64(suite.client.ID), v)

		// User id should be nil
		assert.False(suite.T(), tokens[0].UserID.Valid)
		v, err = tokens[0].UserID.Value()
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), nil, v)
	}

	// Grant a user specific access token
	accessToken, err = s.GrantAccessToken(
		suite.client,
		suite.user,
		"doesn't matter",
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct access token object should be returned
	if assert.NotNil(suite.T(), accessToken) {
		// Fetch all access tokens
		s.db.Preload("Client").Preload("User").Find(&tokens)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(tokens))

		// And the second token should match the one returned by the grant method
		assert.Equal(suite.T(), tokens[1].Token, accessToken.Token)

		// Client id should be set
		assert.True(suite.T(), tokens[1].ClientID.Valid)
		v, err = tokens[1].ClientID.Value()
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), int64(suite.client.ID), v)

		// User id should be set
		assert.True(suite.T(), tokens[1].UserID.Valid)
		v, err = tokens[1].UserID.Value()
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), int64(suite.user.ID), v)
	}
}

func (suite *OauthTestSuite) TestDeleteExpiredAccessTokensClient() {
	// Insert an expired test access token with a user
	if err := suite.db.Create(&AccessToken{
		Token:     "test_token_1",
		ExpiresAt: time.Now().Add(-10 * time.Second),
		Client:    suite.client,
		User:      suite.user,
		Scope:     "doesn't matter",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Insert an expired test access token without a user
	if err := suite.db.Create(&AccessToken{
		Token:     "test_token_2",
		ExpiresAt: time.Now().Add(-10 * time.Second),
		Client:    suite.client,
		Scope:     "doesn't matter",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Insert a test access token with a user
	if err := suite.db.Create(&AccessToken{
		Token:     "test_token_3",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.client,
		User:      suite.user,
		Scope:     "doesn't matter",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Insert a test access token without a user
	if err := suite.db.Create(&AccessToken{
		Token:     "test_token_4",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.client,
		Scope:     "doesn't matter",
	}).Error; err != nil {
		log.Fatal(err)
	}

	var notFound bool
	var existingTokens []string

	// This should only delete test_token_1
	suite.service.deleteExpiredAccessTokens(suite.client, suite.user)

	// Check the test_token_1 was deleted
	notFound = suite.db.Where(AccessToken{
		Token: "test_token_1",
	}).First(&AccessToken{}).RecordNotFound()
	assert.True(suite.T(), notFound)

	// Check the other three tokens are still around
	existingTokens = []string{
		"test_token_2",
		"test_token_3",
		"test_token_4",
	}
	for _, token := range existingTokens {
		notFound = suite.db.Where(AccessToken{
			Token: token,
		}).First(new(AccessToken)).RecordNotFound()
		assert.False(suite.T(), notFound)
	}

	// This should only delete test_token_2
	suite.service.deleteExpiredAccessTokens(suite.client, nil)

	// Check the test_token_2 was deleted
	notFound = suite.db.Where(AccessToken{
		Token: "test_token_2",
	}).First(new(AccessToken)).RecordNotFound()
	assert.True(suite.T(), notFound)

	// Check that last two tokens are still around
	existingTokens = []string{
		"test_token_3",
		"test_token_4",
	}
	for _, token := range existingTokens {
		notFound := suite.db.Where(AccessToken{
			Token: token,
		}).First(new(AccessToken)).RecordNotFound()
		assert.False(suite.T(), notFound)
	}
}