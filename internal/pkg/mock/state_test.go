package mock_test

import (
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewState(t *testing.T) {
	state, err := mock.NewState()
	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(state.Guilds) == 0 {
		t.Fatal("Empty guilds state")
	}
}
