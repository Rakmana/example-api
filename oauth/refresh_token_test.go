package oauth_test

import (
	"time"

	"github.com/RichardKnop/example-api/oauth"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenCreatesNew() {
	var (
		refreshToken *oauth.RefreshToken
		err          error
		tokens       []*oauth.RefreshToken
	)

	// Since there is no user specific token,
	// a new one should be created and returned
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be just one token now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[0].ClientID.Valid)
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].Client.ID)

		// User ID should be set
		assert.True(suite.T(), tokens[0].UserID.Valid)
		assert.Equal(suite.T(), suite.users[0].ID, tokens[0].User.ID)
	}

	// Valid user specific token exists, new one should NOT be created
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be just one token now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[0].ClientID.Valid)
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[0].Client.ID)

		// User ID should be set
		assert.True(suite.T(), tokens[0].UserID.Valid)
		assert.Equal(suite.T(), suite.users[0].ID, tokens[0].User.ID)
	}

	// Since there is no client only token,
	// a new one should be created and returned
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be 2 tokens
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[1].ClientID.Valid)
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].Client.ID)

		// User ID should be nil
		assert.False(suite.T(), tokens[1].UserID.Valid)
	}

	// Valid client only token exists, new one should NOT be created
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be 2 tokens
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[1].ClientID.Valid)
		assert.Equal(suite.T(), suite.clients[0].ID, tokens[1].Client.ID)

		// User ID should be nil
		assert.False(suite.T(), tokens[1].UserID.Valid)
	}
}

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenReturnsExisting() {
	var (
		refreshToken *oauth.RefreshToken
		err          error
		tokens       []*oauth.RefreshToken
	)

	// Insert an access token without a user
	err = suite.db.Create(&oauth.RefreshToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.clients[0],
	}).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Since the current client only token is valid, this should just return it
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct refresh token should be returned
	if assert.NotNil(suite.T(), refreshToken) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be just one token right now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)
		assert.Equal(suite.T(), "test_token", refreshToken.Token)
		assert.Equal(suite.T(), "test_token", tokens[0].Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[0].ClientID.Valid)
		assert.Equal(suite.T(), int64(suite.clients[0].ID), tokens[0].ClientID.Int64)

		// User ID should be nil
		assert.False(suite.T(), tokens[0].UserID.Valid)
	}

	// Insert an access token with a user
	err = suite.db.Create(&oauth.RefreshToken{
		Token:     "test_token2",
		ExpiresAt: time.Now().Add(+10 * time.Second),
		Client:    suite.clients[0],
		User:      suite.users[0],
	}).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Since the current user specific only token is valid,
	// this should just return it
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be Nil
	assert.Nil(suite.T(), err)

	// Correct refresh token should be returned
	if assert.NotNil(suite.T(), refreshToken) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db).Order("id").Find(&tokens)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)
		assert.Equal(suite.T(), "test_token2", refreshToken.Token)
		assert.Equal(suite.T(), "test_token2", tokens[1].Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[1].ClientID.Valid)
		assert.Equal(suite.T(), int64(suite.clients[0].ID), tokens[1].ClientID.Int64)

		// User ID should be set
		assert.True(suite.T(), tokens[1].UserID.Valid)
		assert.Equal(suite.T(), int64(suite.users[0].ID), tokens[1].UserID.Int64)
	}
}

func (suite *OauthTestSuite) TestGetOrCreateRefreshTokenDeletesExpired() {
	var (
		refreshToken *oauth.RefreshToken
		err          error
		tokens       []*oauth.RefreshToken
	)

	// Insert an expired client only test refresh token
	err = suite.db.Create(&oauth.RefreshToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(-10 * time.Second),
		Client:    suite.clients[0],
	}).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Since the current client only token is expired,
	// this should delete it and create and return a new one
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		nil,              // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db.Unscoped()).Order("id").Find(&tokens)

		// There should be just one token right now
		assert.Equal(suite.T(), 1, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[0].Token, refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", tokens[0].Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[0].ClientID.Valid)
		assert.Equal(suite.T(), int64(suite.clients[0].ID), tokens[0].ClientID.Int64)

		// User ID should be nil
		assert.False(suite.T(), tokens[0].UserID.Valid)
	}

	// Insert an expired user specific test refresh token
	err = suite.db.Create(&oauth.RefreshToken{
		Token:     "test_token",
		ExpiresAt: time.Now().Add(-10 * time.Second),
		Client:    suite.clients[0],
		User:      suite.users[0],
	}).Error
	assert.NoError(suite.T(), err, "Inserting test data failed")

	// Since the current user specific token is expired,
	// this should delete it and create and return a new one
	refreshToken, err = suite.service.GetOrCreateRefreshToken(
		suite.clients[0], // client
		suite.users[0],   // user
		3600,             // expires in
		"read_write",     // scope
	)

	// Error should be nil
	if assert.Nil(suite.T(), err) {
		// Fetch all refresh tokens
		oauth.RefreshTokenPreload(suite.db.Unscoped()).Order("id").Find(&tokens)

		// There should be 2 tokens now
		assert.Equal(suite.T(), 2, len(tokens))

		// Correct refresh token object should be returned
		assert.NotNil(suite.T(), refreshToken)
		assert.Equal(suite.T(), tokens[1].Token, refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", refreshToken.Token)
		assert.NotEqual(suite.T(), "test_token", tokens[1].Token)

		// Client ID should be set
		assert.True(suite.T(), tokens[1].ClientID.Valid)
		assert.Equal(suite.T(), int64(suite.clients[0].ID), tokens[1].ClientID.Int64)

		// User ID should be set
		assert.True(suite.T(), tokens[1].UserID.Valid)
		assert.Equal(suite.T(), int64(suite.users[0].ID), tokens[1].UserID.Int64)
	}
}

func (suite *OauthTestSuite) TestGetValidRefreshToken() {
	var (
		refreshToken *oauth.RefreshToken
		err          error
	)

	// Insert some test refresh tokens
	testRefreshTokens := []*oauth.RefreshToken{
		// Expired test refresh token
		&oauth.RefreshToken{
			Token:     "test_expired_token",
			ExpiresAt: time.Now().Add(-10 * time.Second),
			Client:    suite.clients[0],
			User:      suite.users[0],
		},
		// Refresh token
		&oauth.RefreshToken{
			Token:     "test_token",
			ExpiresAt: time.Now().Add(+10 * time.Second),
			Client:    suite.clients[0],
			User:      suite.users[0],
		},
	}
	for _, testRefreshToken := range testRefreshTokens {
		err := suite.db.Create(testRefreshToken).Error
		assert.NoError(suite.T(), err, "Inserting test data failed")
	}

	// Test passing an empty token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"",               // refresh token
		suite.clients[0], // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenNotFound, err)
	}

	// Test passing a bogus token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"bogus",          // refresh token
		suite.clients[0], // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenNotFound, err)
	}

	// Test passing an expired token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"test_expired_token", // refresh token
		suite.clients[0],     // client
	)

	// Refresh token should be nil
	assert.Nil(suite.T(), refreshToken)

	// Correct error should be returned
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), oauth.ErrRefreshTokenExpired, err)
	}

	// Test passing a valid token
	refreshToken, err = suite.service.GetValidRefreshToken(
		"test_token",     // refresh token
		suite.clients[0], // client
	)

	// Error should be nil
	assert.Nil(suite.T(), err)

	// Correct refresh token object should be returned
	assert.NotNil(suite.T(), refreshToken)
	assert.Equal(suite.T(), "test_token", refreshToken.Token)
}
