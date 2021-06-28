package oauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/oauth/tokentypes"
	testutil "github.com/resonatecoop/id/test-util"
	"github.com/resonatecoop/id/util"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestAuthorizationCodeGrantEmptyNotFound() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"authorization_code"},
		"code":       {""},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAuthorizationCodeNotFound.Error(),
		404,
	)
}

func (suite *OauthTestSuite) TestAuthorizationCodeGrantBogusNotFound() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type": {"authorization_code"},
		"code":       {"bogus"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAuthorizationCodeNotFound.Error(),
		404,
	)
}

func (suite *OauthTestSuite) TestAuthorizationCodeGrantExpired() {
	// Insert a test authorization code

	ctx := context.Background()

	authorizationcode := &model.AuthorizationCode{
		IDRecord:    model.IDRecord{CreatedAt: time.Now().UTC()},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(-10 * time.Second),
		ClientID:    suite.clients[0].ID,
		UserID:      suite.users[0].ID,
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}

	_, err := suite.db.NewInsert().
		Model(authorizationcode).
		Exec(ctx)

	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"test_code"},
		"redirect_uri": {"https://www.example.com"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAuthorizationCodeExpired.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestAuthorizationCodeGrantInvalidRedirectURI() {
	// Insert a test authorization code

	ctx := context.Background()

	authorizationCode := &model.AuthorizationCode{
		IDRecord:    model.IDRecord{CreatedAt: time.Now().UTC()},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(+10 * time.Second),
		ClientID:    suite.clients[0].ID,
		UserID:      suite.users[0].ID,
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}

	_, err := suite.db.NewInsert().
		Model(authorizationCode).
		Exec(ctx)

	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"test_code"},
		"redirect_uri": {"https://bogus"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrInvalidRedirectURI.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestAuthorizationCodeGrant() {
	ctx := context.Background()
	// Insert a test authorization code
	authorizationcode := &model.AuthorizationCode{
		IDRecord:    model.IDRecord{CreatedAt: time.Now().UTC()},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(+10 * time.Second),
		ClientID:    suite.clients[0].ID,
		UserID:      suite.users[0].ID,
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write tenantadmin",
	}

	_, err := suite.db.NewInsert().
		Model(authorizationcode).
		Exec(ctx)

	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {"test_code"},
		"redirect_uri": {"https://www.example.com"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	accessToken, refreshToken := new(model.AccessToken), new(model.RefreshToken)

	err = suite.db.NewSelect().
		Model(accessToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	assert.Nil(suite.T(), err)

	// Check the response
	expected := &oauth.AccessTokenResponse{
		UserID:       accessToken.UserID.String(),
		AccessToken:  accessToken.Token,
		ExpiresIn:    3600,
		TokenType:    tokentypes.Bearer,
		Scope:        "read_write tenantadmin",
		RefreshToken: refreshToken.Token,
	}
	testutil.TestResponseObject(suite.T(), w, expected, 200)

	// The authorization code should get deleted after use

	err = suite.db.NewSelect().
		Model(new(model.AuthorizationCode)).
		Limit(1).
		Scan(ctx)

	assert.NotNil(suite.T(), err)
}
