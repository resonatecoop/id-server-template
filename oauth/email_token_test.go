package oauth_test

import (
	"github.com/RichardKnop/go-oauth2-server/models"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestPasswordReset() {
	var (
		err error
	)

	_, err = suite.service.SendEmailToken(models.NewOauthEmail(
		"test@localhost",
		"Reset your password",
		"password-reset",
	), "https://id.resonate.localhost/password-reset")

	assert.Nil(suite.T(), err)
}
