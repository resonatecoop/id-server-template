package oauth_test

import (
	"context"
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
		"username":   {"test@user.com"},
		"password":   {"test_password"},
		"scope":      {"read_write artist"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	accessToken, refreshToken := new(model.AccessToken), new(model.RefreshToken)

	ctx := context.Background()

	err = suite.db.NewSelect().
		Model(accessToken).
		Limit(1).
		Scan(ctx)

	// an access token is found
	assert.Nil(suite.T(), err)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Limit(1).
		Scan(ctx)

	// a refresh token is founds
	assert.Nil(suite.T(), err)

	// Check the response
	expected := &oauth.AccessTokenResponse{
		UserID:       accessToken.UserID.String(),
		AccessToken:  accessToken.Token,
		ExpiresIn:    3600,
		TokenType:    tokentypes.Bearer,
		Scope:        "read_write artist",
		RefreshToken: refreshToken.Token,
	}
	testutil.TestResponseObject(suite.T(), w, expected, 200)
}

func (suite *OauthTestSuite) TestPasswordGrantWithRoleRestriction() {
	suite.service.RestrictToRoles(int32(model.SuperAdminRole))

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"password"},
		"username":   {"test@user.com"},
		"password":   {"test_password"},
		"scope":      {"read_write artist"},
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

	suite.service.RestrictToRoles(int32(model.SuperAdminRole), int32(model.AdminRole), int32(model.TenantAdminRole), int32(model.LabelRole), int32(model.ArtistRole), int32(model.UserRole))
}
