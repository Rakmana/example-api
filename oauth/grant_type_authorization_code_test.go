package oauth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/RichardKnop/recall/util"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestAuthorizationCodeGrant() {
	// Insert a test authorization code
	if err := suite.db.Create(&AuthorizationCode{
		Code:        "test_code",
		ExpiresAt:   time.Now().Add(+10 * time.Second),
		Client:      suite.clients[0],
		User:        suite.users[0],
		RedirectURI: util.StringOrNull("https://www.example.com"),
		Scope:       "read_write",
	}).Error; err != nil {
		log.Fatal(err)
	}

	// Prepare a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/something", nil)
	if err != nil {
		log.Fatal(err)
	}
	r.Form = url.Values{
		"grant_type": {"authorization_code"},
		"code":       {"test_code"},
	}

	var w *httptest.ResponseRecorder

	// First we will test an invalid redirect URI error
	w = httptest.NewRecorder()
	suite.service.authorizationCodeGrant(w, r, suite.clients[0])

	// Check the status code
	assert.Equal(suite.T(), 400, w.Code)

	// Check the response body
	assert.Equal(
		suite.T(),
		fmt.Sprintf("{\"error\":\"%s\"}", errInvalidRedirectURI.Error()),
		strings.TrimSpace(w.Body.String()),
	)

	// Now add the redirect URI parameter
	r.Form.Set("redirect_uri", "https://www.example.com")

	// And test a successful case
	w = httptest.NewRecorder()
	suite.service.authorizationCodeGrant(w, r, suite.clients[0])

	// Check the status code
	assert.Equal(suite.T(), 200, w.Code)

	// Check the correct data was inserted
	accessToken := new(AccessToken)
	assert.False(suite.T(), suite.db.First(accessToken).RecordNotFound())
	refreshToken := new(RefreshToken)
	assert.False(suite.T(), suite.db.First(refreshToken).RecordNotFound())

	// Check the response body
	expected, err := json.Marshal(&AccessTokenResponse{
		ID:           accessToken.ID,
		AccessToken:  accessToken.Token,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Scope:        "read_write",
		RefreshToken: refreshToken.Token,
	})
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(suite.T(), string(expected), strings.TrimSpace(w.Body.String()))

	// Check the authorization code was deleted
	notFound := suite.db.Unscoped().First(new(AuthorizationCode)).RecordNotFound()
	assert.True(suite.T(), notFound)
}
