package oauth_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	uuid "github.com/google/uuid"
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
	err := suite.db.Create(&model.AuthorizationCode{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(-10 * time.Second),
		Client:      suite.clients[0],
		User:        suite.users[0],
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}).Error
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
	err := suite.db.Create(&model.AuthorizationCode{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(+10 * time.Second),
		Client:      suite.clients[0],
		User:        suite.users[0],
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}).Error
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
	// Insert a test authorization code
	err := suite.db.Create(&model.AuthorizationCode{
		MyGormModel: model.MyGormModel{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
		},
		Code:        "test_code",
		ExpiresAt:   time.Now().UTC().Add(+10 * time.Second),
		Client:      suite.clients[0],
		User:        suite.users[0],
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}).Error
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

	// The authorization code should get deleted after use
	assert.True(suite.T(), suite.db.Unscoped().
		First(new(model.AuthorizationCode)).RecordNotFound())
}
