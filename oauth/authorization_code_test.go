package oauth_test

import (
	"context"

	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestGrantAuthorizationCode() {
	var (
		ctx               context.Context
		authorizationCode *model.AuthorizationCode
		err               error
		codes             []*model.AuthorizationCode
	)

	// Grant an authorization code
	authorizationCode, err = suite.service.GrantAuthorizationCode(
		suite.clients[0],              // client
		suite.users[0],                // user
		3600,                          // expires in
		"redirect URI doesn't matter", // redirect URI
		"scope doesn't matter",        // scope
	)

	ctx = context.Background()

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct authorization code object should be returned
	if assert.NotNil(suite.T(), authorizationCode) {
		// Fetch all auth codes
		rows, err := suite.db.QueryContext(ctx, "SELECT * FROM authorization_codes ORDER BY created_at")
		if err != nil {
			panic(err)
		}

		err = suite.db.ScanRows(ctx, rows, &codes)

		// There should be just one right now
		assert.Equal(suite.T(), 1, len(codes))

		// And the code should match the one returned by the grant method
		assert.Equal(suite.T(), codes[0].Code, authorizationCode.Code)

		// Client ID should be set
		assert.True(suite.T(), util.IsValidUUID(codes[0].ClientID.String()))
		assert.Equal(suite.T(), suite.clients[0].ID.String(), codes[0].ClientID.String())

		// User ID should be set
		assert.True(suite.T(), util.IsValidUUID(codes[0].UserID.String()))
		assert.Equal(suite.T(), suite.users[0].ID.String(), codes[0].UserID.String())
	}
}
