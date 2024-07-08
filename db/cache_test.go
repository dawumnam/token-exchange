package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlackListToken(t *testing.T) {
	Init()
	token := "test_token"
	expiry := time.Now().Add(time.Hour)

	err := BlackListToken(token, expiry)
	assert.NoError(t, err)

	blacklisted, err := IsTokenBlacklisted(token)
	assert.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestIsTokenBlacklisted(t *testing.T) {
	nonBlacklistedToken := "non_blacklisted_token"

	blacklisted, err := IsTokenBlacklisted(nonBlacklistedToken)
	assert.NoError(t, err)
	assert.False(t, blacklisted)

	blacklistedToken := "blacklisted_token"
	err = BlackListToken(blacklistedToken, time.Now().Add(time.Hour))
	assert.NoError(t, err)

	blacklisted, err = IsTokenBlacklisted(blacklistedToken)
	assert.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestBlackListTokenExpiry(t *testing.T) {
	token := "expiring_token"
	expiry := time.Now().Add(2 * time.Second)

	err := BlackListToken(token, expiry)
	assert.NoError(t, err)

	blacklisted, err := IsTokenBlacklisted(token)
	assert.NoError(t, err)
	assert.True(t, blacklisted)

	time.Sleep(3 * time.Second)

	blacklisted, err = IsTokenBlacklisted(token)
	assert.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestCleanup(t *testing.T) {
	client, _ := GetRedisClient()
	err := client.FlushDB(ctx).Err()
	assert.NoError(t, err)
}
