package oauth_test

import (
	"time"

	uuid "github.com/google/uuid"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/util"
	pass "github.com/resonatecoop/id/util/password"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestUserExistsFindsValidUser() {
	validUsername := suite.users[0].Username
	assert.True(suite.T(), suite.service.UserExists(validUsername))
}

func (suite *OauthTestSuite) TestUserExistsDoesntFindInvalidUser() {
	invalidUsername := "bogus_name"
	assert.False(suite.T(), suite.service.UserExists(invalidUsername))
}

func (suite *OauthTestSuite) TestUpdateUsernameWorksWithValidEntry() {
	user, err := suite.service.CreateUser(
		model.UserRole,  // role ID
		"test@newuser",  // username
		"test_password", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser", user.Username)

	newUsername := "mynew@email"

	err = suite.service.UpdateUsername(user, newUsername)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), newUsername, user.Username)
}

func (suite *OauthTestSuite) TestUpdateUsernameTxWorksWithValidEntry() {
	user, err := suite.service.CreateUser(
		roles.User,      // role ID
		"test@newuser",  // username
		"test_password", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser", user.Username)

	newUsername := "mynew@email"

	err = suite.service.UpdateUsernameTx(suite.db, user, newUsername)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), newUsername, user.Username)
}

func (suite *OauthTestSuite) TestUpdateUsernameFailsWithABlankEntry() {
	user, err := suite.service.CreateUser(
		roles.User,      // role ID
		"test@newuser",  // username
		"test_password", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser", user.Username)

	newUsername := ""

	err = suite.service.UpdateUsername(user, newUsername)

	assert.EqualError(suite.T(), err, oauth.ErrCannotSetEmptyUsername.Error())

	assert.NotEqual(suite.T(), newUsername, user.Username)
}

func (suite *OauthTestSuite) TestFindUserByUsername() {
	var (
		user *model.User
		err  error
	)

	// When we try to find a user with a bogus username
	user, err = suite.service.FindUserByUsername("bogus")

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUserNotFound, err)
	}

	// When we try to find a user with a valid username
	user, err = suite.service.FindUserByUsername("test@user")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user", user.Username)
	}

	// Test username case insensitiviness
	user, err = suite.service.FindUserByUsername("TeSt@UsEr")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user", user.Username)
	}
}

func (suite *OauthTestSuite) TestCreateUser() {
	var (
		user *model.User
		err  error
	)

	// We try to insert a non unique user
	user, err = suite.service.CreateUser(
		roles.User,      // role ID
		"test@user",     // username
		"test_password", // password
	)

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUsernameTaken.Error(), err.Error())
	}

	// We try to insert a unique user
	user, err = suite.service.CreateUser(
		roles.User,      // role ID
		"test@newuser",  // username
		"test_password", // password
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@newuser", user.Username)
	}

	// Test username case insensitivity
	user, err = suite.service.CreateUser(
		roles.User,      // role ID
		"TeSt@NeWuSeR2", // username
		"test_password", // password
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@newuser2", user.Username)
	}
}

func (suite *OauthTestSuite) TestSetPassword() {
	var (
		user *model.User
		err  error
	)

	// Insert a test user without a password
	user = &model.User{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		RoleID:   util.StringOrNull(roles.User),
		Username: "test@user_nopass",
		Password: util.StringOrNull(""),
	}
	err = suite.db.Create(user).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Try to set an empty password
	err = suite.service.SetPassword(user, "")

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrPasswordTooShort, err)
	}

	// Try changing the password
	err = suite.service.SetPassword(user, "test_password")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// User object should have been updated
	assert.Equal(suite.T(), "test@user_nopass", user.Username)
	assert.Nil(suite.T(), pass.VerifyPassword(user.Password.String, "test_password"))
}

func (suite *OauthTestSuite) TestAuthUser() {
	var (
		user *model.User
		err  error
	)

	// Insert a test user without a password
	err = suite.db.Create(&model.User{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		RoleID:   util.StringOrNull(roles.User),
		Username: "test@user_nopass",
		Password: util.StringOrNull(""),
	}).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// When we try to authenticate a user without a password
	user, err = suite.service.AuthUser("test@user_nopass", "bogus")

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUserPasswordNotSet, err)
	}

	// When we try to authenticate with a bogus username
	user, err = suite.service.AuthUser("bogus", "test_password")

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUserNotFound, err)
	}

	// When we try to authenticate with an invalid password
	user, err = suite.service.AuthUser("test@user", "bogus")

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrInvalidUserPassword, err)
	}

	// When we try to authenticate with valid username and password
	user, err = suite.service.AuthUser("test@user", "test_password")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user", user.Username)
	}

	// Test username case insensitivity
	user, err = suite.service.AuthUser("TeSt@UsEr", "test_password")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user", user.Username)
	}
}

func (suite *OauthTestSuite) TestBlankPassword() {
	var (
		user *model.User
		err  error
	)

	user, err = suite.service.CreateUser(
		roles.User,         // role ID
		"test@user_nopass", // username
		"",                 // password,
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user_nopass", user.Username)
	}

	// When we try to authenticate
	user, err = suite.service.AuthUser("test@user_nopass", "")

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUserPasswordNotSet, err)
	}
}
