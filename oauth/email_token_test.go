package oauth_test

import (
	"errors"

	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

var (
	// ErrEmailValidAPIKeyNotProvided ...
	ErrEmailValidAPIKeyNotProvided = errors.New("you must provide a valid api-key before calling Send()")
)

func (suite *OauthTestSuite) TestPasswordReset() {
	var (
		err error
	)

	_, err = suite.service.SendEmailToken(model.NewOauthEmail(
		"test@user.com",
		"Reset your password",
		"password-reset",
	), "https://id.resonate.localhost/password-reset")

	assert.Equal(suite.T(), ErrEmailValidAPIKeyNotProvided, err)

	//assert.Equal(suite.T(), true, (err == nil || err == ErrEmailValidAPIKeyNotProvided))
}
