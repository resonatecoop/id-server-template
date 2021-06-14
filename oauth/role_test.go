package oauth_test

import (
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestFindRoleByID() {
	var (
		role *model.AccessRole
		err  error
	)

	// Let's try to find a role by a bogus ID
	role, err = suite.service.FindRoleByID(99)

	// Role should be nil
	assert.Nil(suite.T(), role)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRoleNotFound, err)
	}

	// Now let's pass a valid ID
	role, err = suite.service.FindRoleByID(int32(model.UserRole))

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct role should be returned
	if assert.NotNil(suite.T(), role) {
		assert.Equal(suite.T(), model.UserRole, *role)
	}
}
