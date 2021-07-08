package oauth_test

import (
	"context"
	"time"

	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api/model"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestAuthenticate() {
	var (
		ctx         context.Context
		accessToken *model.AccessToken
		err         error
	)

	// Insert some test access tokens
	testAccessTokens := []*model.AccessToken{
		// Expired access token
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_expired_token",
			ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		// Access token without a user
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_client_token_2",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
		},
		// Access token with a user
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_user_token_3",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
	}

	ctx = context.Background()
	for _, testAccessToken := range testAccessTokens {

		_, err := suite.db.NewInsert().
			Model(testAccessToken).
			Exec(ctx)

		assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// Test passing an empty token
	accessToken, err = suite.service.Authenticate("")

	// Access token should be nil
	assert.Nil(suite.T(), accessToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrAccessTokenNotFound, err)
	}

	// Test passing a bogus token
	accessToken, err = suite.service.Authenticate("bogus")

	// Access token should be nil
	assert.Nil(suite.T(), accessToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrAccessTokenNotFound, err)
	}

	// Test passing an expired token
	accessToken, err = suite.service.Authenticate("test_expired_token")

	// Access token should be nil
	assert.Nil(suite.T(), accessToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrAccessTokenExpired, err)
	}

	// Test passing a valid client token
	accessToken, err = suite.service.Authenticate("test_client_token_2")

	// Correct access token should be returned
	if assert.NotNil(suite.T(), accessToken) {
		assert.Equal(suite.T(), "test_client_token_2", accessToken.Token)
		assert.EqualValues(suite.T(), suite.clients[0].ID, accessToken.ClientID)
		assert.EqualValues(suite.T(), accessToken.UserID, uuid.Nil)
	}

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Test passing a valid user token
	accessToken, err = suite.service.Authenticate("test_user_token_3")

	// Correct access token should be returned
	if assert.NotNil(suite.T(), accessToken) {
		assert.Equal(suite.T(), "test_user_token_3", accessToken.Token)
		assert.EqualValues(suite.T(), suite.clients[0].ID, accessToken.ClientID)
		assert.EqualValues(suite.T(), suite.users[0].ID, accessToken.UserID)
	}

	// Error should be nil
	assert.Nil(suite.T(), err)
}

func (suite *OauthTestSuite) TestAuthenticateRollingRefreshToken() {
	var (
		ctx               context.Context
		testAccessTokens  []*model.AccessToken
		testRefreshTokens []*model.RefreshToken
		accessToken       *model.AccessToken
		err               error
		refreshTokens     []*model.RefreshToken
	)

	// Insert some test access tokens
	testAccessTokens = []*model.AccessToken{
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_1",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_2",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_3",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[1].ID,
		},
	}

	ctx = context.Background()

	for _, testAccessToken := range testAccessTokens {

		_, err := suite.db.NewInsert().
			Model(testAccessToken).
			Exec(ctx)

		assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// Insert some test access tokens
	testRefreshTokens = []*model.RefreshToken{
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC().Add(+1 * time.Second)},
			Token:     "test_token_1",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC().Add(+2 * time.Second)},
			Token:     "test_token_2",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC().Add(+3 * time.Second)},
			Token:     "test_token_3",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[1].ID,
		},
	}
	for _, testRefreshToken := range testRefreshTokens {

		_, err := suite.db.NewInsert().
			Model(testRefreshToken).
			Exec(ctx)

		assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// Authenticate with the first access token
	//	now1 := time.Now().UTC()
	// gorm.NowFunc = func() time.Time {
	// 	return now1
	// }
	accessToken, err = suite.service.Authenticate("test_token_1")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_1", accessToken.Token)
	assert.EqualValues(suite.T(), suite.clients[0].ID, accessToken.ClientID)
	assert.EqualValues(suite.T(), suite.users[0].ID, accessToken.UserID)

	// First refresh token expiration date should be extended
	refreshTokens = make([]*model.RefreshToken, len(testRefreshTokens))
	// err = suite.db.NewSelect().
	// 	Model

	rows, err := suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens WHERE token IN ('test_token_1', 'test_token_2', 'test_token_3') ORDER BY created_at")
	if err != nil {
		panic(err)
	}

	err = suite.db.ScanRows(ctx, rows, &refreshTokens)

	// err = suite.db.Where(
	// 	"token IN ('test_token_1', 'test_token_2', 'test_token_3')",
	// ).Order("created_at").Find(&refreshTokens).Error
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_1", refreshTokens[0].Token)
	assert.Equal(
		suite.T(),
		//now1.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_2", refreshTokens[1].Token)
	assert.Equal(
		suite.T(),
		testRefreshTokens[1].ExpiresAt.Unix(),
		refreshTokens[1].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_3", refreshTokens[2].Token)
	assert.Equal(
		suite.T(),
		testRefreshTokens[2].ExpiresAt.Unix(),
		refreshTokens[2].ExpiresAt.Unix(),
	)

	// Authenticate with the second access token
	// now2 := time.Now().UTC()
	// gorm.NowFunc = func() time.Time {
	// 	return now2
	// }
	accessToken, err = suite.service.Authenticate("test_token_2")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_2", accessToken.Token)
	assert.EqualValues(suite.T(), suite.clients[0].ID, accessToken.ClientID)
	assert.EqualValues(suite.T(), accessToken.UserID, uuid.Nil)

	// Second refresh token expiration date should be extended
	refreshTokens = make([]*model.RefreshToken, len(testRefreshTokens))

	rows, err = suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens WHERE token IN ('test_token_1', 'test_token_2', 'test_token_3') ORDER BY created_at")
	if err != nil {
		panic(err)
	}

	err = suite.db.ScanRows(ctx, rows, &refreshTokens)

	// err = suite.db.Where(
	// 	"token IN ('test_token_1', 'test_token_2', 'test_token_3')",
	// ).Order("created_at").Find(&refreshTokens).Error
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_1", refreshTokens[0].Token)
	assert.Equal(
		suite.T(),
		//now1.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_2", refreshTokens[1].Token)
	assert.Equal(
		suite.T(),
		//now2.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[1].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[1].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_3", refreshTokens[2].Token)
	assert.Equal(
		suite.T(),
		testRefreshTokens[2].ExpiresAt.Unix(),
		refreshTokens[2].ExpiresAt.Unix(),
	)

	// Authenticate with the third access token
	//now3 := time.Now().UTC()
	// gorm.NowFunc = func() time.Time {
	// 	return now3
	// }
	accessToken, err = suite.service.Authenticate("test_token_3")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_3", accessToken.Token)
	assert.EqualValues(suite.T(), suite.clients[0].ID, accessToken.ClientID)
	assert.EqualValues(suite.T(), suite.users[1].ID, accessToken.UserID)

	// First refresh token expiration date should be extended
	refreshTokens = make([]*model.RefreshToken, len(testRefreshTokens))
	rows, err = suite.db.QueryContext(ctx, "SELECT * FROM refresh_tokens WHERE token IN ('test_token_1', 'test_token_2', 'test_token_3') ORDER BY created_at")
	if err != nil {
		panic(err)
	}

	err = suite.db.ScanRows(ctx, rows, &refreshTokens)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "test_token_1", refreshTokens[0].Token)
	assert.Equal(
		suite.T(),
		//now1.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[0].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_2", refreshTokens[1].Token)
	assert.Equal(
		suite.T(),
		//now2.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[1].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[1].ExpiresAt.Unix(),
	)
	assert.Equal(suite.T(), "test_token_3", refreshTokens[2].Token)
	assert.Equal(
		suite.T(),
		//now3.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[2].UpdatedAt.Unix()+int64(suite.cnf.Oauth.RefreshTokenLifetime),
		refreshTokens[2].ExpiresAt.Unix(),
	)
}

func (suite *OauthTestSuite) TestClearUserTokens() {
	var (
		ctx               context.Context
		testAccessTokens  []*model.AccessToken
		testRefreshTokens []*model.RefreshToken
		err               error
		testUserSession   *session.UserSession
	)

	// Insert some test access tokens
	testAccessTokens = []*model.AccessToken{
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_1",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_2",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[1].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_3",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[1].ID,
		},
	}

	ctx = context.Background()

	for _, testAccessToken := range testAccessTokens {
		_, err = suite.db.NewInsert().
			Model(testAccessToken).
			Exec(ctx)
		assert.Nil(suite.T(), err)
		//TODO IMPROVE EQUIVALENCE: assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// Insert some test access tokens
	testRefreshTokens = []*model.RefreshToken{
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_1",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_2",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[1].ID,
			UserID:    suite.users[0].ID,
		},
		{
			IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
			Token:     "test_token_3",
			ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
			ClientID:  suite.clients[0].ID,
			UserID:    suite.users[1].ID,
		},
	}

	for _, testRefreshToken := range testRefreshTokens {
		_, err = suite.db.NewInsert().
			Model(testRefreshToken).
			Exec(ctx)

		assert.Nil(suite.T(), err)
		//IMPROVE EQUIVALENCE assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	testUserSession = &session.UserSession{
		ClientID:     suite.clients[0].Key,
		Username:     suite.users[0].Username,
		AccessToken:  "test_token_1",
		RefreshToken: "test_token_1",
	}

	// Remove test_token_1 from accress and refresh token tables
	suite.service.ClearUserTokens(testUserSession)

	// Assert that the refresh token was removed

	refreshToken := new(model.RefreshToken)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Where("token = ?", testUserSession.RefreshToken).
		Limit(1).
		Scan(ctx)

	assert.NotNil(suite.T(), err)
	//TODO Improve assertion to be specific about scan error?

	// Assert that the access token was removed
	accessToken := new(model.AccessToken)

	err = suite.db.NewSelect().
		Model(accessToken).
		Where("token = ?", testUserSession.AccessToken).
		Limit(1).
		Scan(ctx)

	assert.NotNil(suite.T(), err)
	//TODO Improve assertion to be specific about scan error?

	// Assert that the other two tokens are still there
	// Refresh tokens
	err = suite.db.NewSelect().
		Model(refreshToken).
		Where("token = ?", "test_token_2").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Where("token = ?", "test_token_3").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)

	// Access tokens
	err = suite.db.NewSelect().
		Model(accessToken).
		Where("token = ?", "test_token_2").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)

	err = suite.db.NewSelect().
		Model(accessToken).
		Where("token = ?", "test_token_3").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)
}
