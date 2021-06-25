package oauth_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
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

func (suite *OauthTestSuite) TestCreateToken() {
	ctx := context.Background()

	newEmail := &model.Email{
		Subject:   "Greatly over exaggerated",
		Recipient: "joan@waters.com",
		Template:  "Dear Madam, about your claim in the newspaper:",
	}

	emailToken, err := suite.service.CreateEmailToken(newEmail.Recipient)

	// No error
	assert.Nil(suite.T(), err)

	// Reference is not nil
	assert.NotEqual(suite.T(), emailToken.Reference.String(), uuid.Nil.String())

	myNewEmailToken := new(model.EmailToken)

	err = suite.db.NewSelect().
		Model(myNewEmailToken).
		Where("id = ?", emailToken.ID).
		Scan(ctx)

	// No error
	assert.Nil(suite.T(), err)

	assert.Equal(suite.T(), myNewEmailToken.ID.String(), emailToken.ID.String())

}
