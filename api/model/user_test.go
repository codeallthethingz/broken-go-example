package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModel(t *testing.T) {
	accessToken := "1234"
	email := "email@example.com"
	user := &User{
		AccessToken: accessToken,
		Email:       email,
	}
	bytes, err := json.Marshal(user)
	if err != nil {
		t.Error(err)
	}
	require.False(t, strings.Contains(string(bytes), accessToken))
	require.True(t, strings.Contains(string(bytes), email))
}
