package oauth_test

import (
	"context"
	"database/sql"

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

/*
func (suite *OauthTestSuite) TestUpdateUsernameWorksWithValidEntry() {
	ctx := context.Background()

	user, err := suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"test@newuser.com",     // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser.com", user.Username)

	newUsername := "mynew@email.com"

	err = suite.service.UpdateUsername(user, newUsername)

	assert.NoError(suite.T(), err)

	err = suite.db.NewSelect().
		Model(user).
		WherePK().
		Scan(ctx)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), newUsername, user.Username)
}
*/

/*
func (suite *OauthTestSuite) TestUpdateUsernameTxWorksWithValidEntry() {
	ctx := context.Background()

	user, err := suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"test@newuser.com",     // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser.com", user.Username)

	newUsername := "mynew@email.com"

	err = suite.service.UpdateUsernameTx(suite.db, user, newUsername)

	assert.NoError(suite.T(), err)

	err = suite.db.NewSelect().
		Model(user).
		WherePK().
		Scan(ctx)

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), newUsername, user.Username)
}
*/

/*
func (suite *OauthTestSuite) TestUpdateUsernameFailsWithABlankEntry() {
	user, err := suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"test@newuser.com",     // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@newuser.com", user.Username)

	newUsername := ""

	err = suite.service.UpdateUsername(user, newUsername)

	assert.EqualError(suite.T(), err, oauth.ErrCannotSetEmptyUsername.Error())

	assert.NotEqual(suite.T(), newUsername, user.Username)
}
*/

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
	user, err = suite.service.FindUserByUsername("test@user.com")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user.com", user.Username)
	}

	// Test username case insensitiviness
	user, err = suite.service.FindUserByUsername("TeSt@UsEr.CoM")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@user.com", user.Username)
	}
}

/*
func (suite *OauthTestSuite) TestCreateUser() {
	var (
		user *model.User
		err  error
	)

	// We try to insert a non unique user
	user, err = suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"test@user.com",        // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	// User object should be nil
	assert.Nil(suite.T(), user)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrUsernameTaken.Error(), err.Error())
	}

	// We try to insert a unique user
	user, err = suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"test@newuser.com",     // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "test@newuser.com", user.Username)
	}

	// Test username case insensitivity
	user, err = suite.service.CreateUser(
		int32(model.UserRole),  // role ID
		"TeStinG@hOtMaIl.com",  // username
		"C0mpl3xPa$$w0rdAr3U5", // password
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "testing@hotmail.com", user.Username)
	}
}
*/

func (suite *OauthTestSuite) TestSetPassword() {
	var (
		ctx  context.Context
		user *model.User
		err  error
	)

	// Insert a test user without a password
	user = &model.User{
		RoleID:   int32(model.UserRole),
		Username: "test@user_nopass.com",
		Password: util.StringOrNull(""),
	}

	ctx = context.Background()

	_, err = suite.db.NewInsert().
		Model(user).
		Exec(ctx)

	assert.Nil(suite.T(), err)

	// Try changing the password
	err = suite.service.SetPassword(user, "C0mpl3xPa$$w0rdAr3U5")

	// Error should be nil
	assert.Nil(suite.T(), err)

	suite.db.NewSelect().
		Model(user).
		WherePK().
		Scan(ctx)

	// User object should have been updated
	assert.Equal(suite.T(), "test@user_nopass.com", user.Username)
	assert.Nil(suite.T(), pass.VerifyPassword(user.Password.String, "C0mpl3xPa$$w0rdAr3U5"))
}

func (suite *OauthTestSuite) TestAuthUser() {
	var (
		ctx  context.Context
		user *model.User
		err  error
	)

	// Insert a test user without a password
	user = &model.User{
		RoleID:   int32(model.UserRole),
		Username: "test@user_nopass",
		Password: util.StringOrNull(""),
	}

	ctx = context.Background()

	_, err = suite.db.NewInsert().
		Model(user).
		Exec(ctx)

	assert.Nil(suite.T(), err)

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

	// Insert a test user without a password
	user = &model.User{
		RoleID:   int32(model.UserRole),
		Username: "test@user",
		Password: sql.NullString{String: "$2a$10$4J4t9xuWhOKhfjN0bOKNReS9sL3BVSN9zxIr2.VaWWQfRBWh1dQIS", Valid: true},
	}

	ctx = context.Background()

	_, err = suite.db.NewInsert().
		Model(user).
		Exec(ctx)

	assert.Nil(suite.T(), err)

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

/*
func (suite *OauthTestSuite) TestBlankPassword() {
	var (
		//user *model.User
		err error
	)

	_, err = suite.service.CreateUser(
		int32(model.UserRole), // role ID
		"test@user_nopass",    // username
		"",                    // password,  "Password is required" ErrPasswordRequired
	)

	// Error should be nil
	//assert.Nil(suite.T(), err)

	// Actually, password is required
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrPasswordRequired, err)
	}

	// user, _ = suite.service.CreateUser(
	// 	int32(model.UserRole), // role ID
	// 	"test@user_somepass",  // username
	// 	"somepassword",        // password,
	// )

	// // Correct user object should be returned
	// if assert.NotNil(suite.T(), user) {
	// 	assert.Equal(suite.T(), "test@user_somepass", user.Username)
	// }

	// // When we try to authenticate
	// user, err = suite.service.AuthUser("test@user_nopass", "")

	// // User object should be nil
	// assert.Nil(suite.T(), user)

	// // Correct error should be returned
	// if assert.NotNil(suite.T(), err) {
	// 	assert.Equal(suite.T(), oauth.ErrUserPasswordNotSet, err)
	// }
}
*/

/*
func (suite *OauthTestSuite) TestDeleteUser() {
	var (
		user *model.User
		err  error
	)

	// We try to insert a unique user
	user, err = suite.service.CreateUser(
		int32(model.UserRole),   // role ID
		"temporary@newuser.com", // username
		"C0mpl3xPa$$w0rdAr3U5",  // password
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct user object should be returned
	if assert.NotNil(suite.T(), user) {
		assert.Equal(suite.T(), "temporary@newuser.com", user.Username)
	}

	// Delete the user
	err = suite.service.DeleteUser(user, "C0mpl3xPa$$w0rdAr3U5")

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Check it exists but has been soft deleted
	exists := suite.service.UserExists("temporary@newuser.com")

	// Error should be nil
	assert.False(suite.T(), exists)
}
*/
