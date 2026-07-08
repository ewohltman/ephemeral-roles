package mock_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	log := mock.NewLogger()

	require.NotNil(t, log)
}
