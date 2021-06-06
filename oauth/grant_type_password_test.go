package oauth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/resonatecoop/id/oauth"

	"github.com/resonatecoop/id/oauth/tokentypes"
	testutil "github.com/resonatecoop/id/test-util"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestPasswordGrant() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"password"},
		"username":   {"test@user"},
		"password":   {"test_password"},
		"scope":      {"read_write"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	accessToken, refreshToken := new(model.AccessToken), new(model.RefreshToken)
	assert.False(suite.T(), model.AccessTokenPreload(suite.db).
		Last(accessToken).RecordNotFound())
	assert.False(suite.T(), model.RefreshTokenPreload(suite.db).
		Last(refreshToken).RecordNotFound())

	// Check the response
	expected := &oauth.AccessTokenResponse{
		UserID:       accessToken.UserID.String,
		AccessToken:  accessToken.Token,
		ExpiresIn:    3600,
		TokenType:    tokentypes.Bearer,
		Scope:        "read_write",
		RefreshToken: refreshToken.Token,
	}
	testutil.TestResponseObject(suite.T(), w, expected, 200)
}

func (suite *OauthTestSuite) TestPasswordGrantWithRoleRestriction() {
	suite.service.RestrictToRoles(model.)

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"password"},
		"username":   {"test@user"},
		"password":   {"test_password"},
		"scope":      {"read_write"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrInvalidUsernameOrPassword.Error(),
		401,
	)

	suite.service.RestrictToRoles(roles.Superuser, roles.User)
}
