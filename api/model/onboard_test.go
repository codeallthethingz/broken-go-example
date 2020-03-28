package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOnboard(t *testing.T) {
	onboard := NewOnboard()
	require.NotNil(t, onboard.Systems)
	require.NotNil(t, onboard.BeforeJoining)
	require.NotNil(t, onboard.AfterJoining)
}
