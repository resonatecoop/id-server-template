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
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestRefreshTokenGrantEmptyNotFound() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {""},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRefreshTokenNotFound.Error(),
		404,
	)
}

func (suite *OauthTestSuite) TestRefreshTokenGrantBogusNotFound() {
	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"bogus_token"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRefreshTokenNotFound.Error(),
		404,
	)
}

func (suite *OauthTestSuite) TestRefreshTokenGrantExipired() {

	ctx := context.Background()
	// Insert a test refresh token
	refreshtoken := &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}

	_, err := suite.db.NewInsert().
		Model(refreshtoken).
		Exec(ctx)

	// confirm there is no error
	assert.Nil(suite.T(), err)

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
		"scope":         {"read read_write"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRefreshTokenExpired.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestRefreshTokenGrantScopeCannotBeGreater() {
	// Insert a test refresh token
	refreshtoken := &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}

	ctx := context.Background()
	// Insert a test refresh token
	_, err := suite.db.NewInsert().
		Model(refreshtoken).
		Exec(ctx)

	// confirm there is no error
	assert.Nil(suite.T(), err)

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
		"scope":         {"read read_write"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRequestedScopeCannotBeGreater.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestRefreshTokenGrantDefaultsToOriginalScope() {

	refreshtoken := &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write tenantadmin",
	}

	ctx := context.Background()
	// Insert a test refresh token
	_, err := suite.db.NewInsert().
		Model(refreshtoken).
		Exec(ctx)

	// confirm there is no error
	assert.Nil(suite.T(), err)

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	accessToken := new(model.AccessToken)

	err = suite.db.NewSelect().
		Model(accessToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	// record found
	assert.Nil(suite.T(), err)

	refreshToken := new(model.RefreshToken)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	// record found
	assert.Nil(suite.T(), err)

	// Check the response body
	expected := &oauth.AccessTokenResponse{
		UserID:       accessToken.UserID.String(),
		AccessToken:  accessToken.Token,
		ExpiresIn:    3600,
		TokenType:    tokentypes.Bearer,
		Scope:        "read_write tenantadmin",
		RefreshToken: refreshToken.Token,
	}
	testutil.TestResponseObject(suite.T(), w, expected, 200)
}

func (suite *OauthTestSuite) TestRefreshTokenGrant() {
	// Insert a test refresh token
	refreshToken := &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write tenantadmin",
	}

	ctx := context.Background()

	_, err := suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	assert.Nil(suite.T(), err)

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/tokens", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"test_token"},
		"scope":         {"read_write tenantadmin"},
	}

	// Serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Fetch data
	accessToken := new(model.AccessToken)

	err = suite.db.NewSelect().
		Model(accessToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	// record found
	assert.Nil(suite.T(), err)

	refreshToken = new(model.RefreshToken)

	err = suite.db.NewSelect().
		Model(refreshToken).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	// record found
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
}
