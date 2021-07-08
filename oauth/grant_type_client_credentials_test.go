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

func (suite *OauthTestSuite) TestClientCredentialsGrant() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"read_write"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	ctx := context.Background()
	accessToken := new(model.AccessToken)

	err = suite.db.NewSelect().
		Model(accessToken).
		Limit(1).
		Scan(ctx)

	// A record is found
	assert.Nil(suite.T(), err)

	// Check the response
	expected := &oauth.AccessTokenResponse{
		AccessToken: accessToken.Token,
		ExpiresIn:   3600,
		TokenType:   tokentypes.Bearer,
		Scope:       "read_write",
	}
	testutil.TestResponseObject(suite.T(), w, expected, 200)

	// Client credentials grant does not produce refresh token
	err = suite.db.NewSelect().
		Model(new(model.RefreshToken)).
		Limit(1).
		Scan(ctx)

	// Error raised as no record found
	assert.NotNil(suite.T(), err)
}
