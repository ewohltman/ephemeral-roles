package mock

import "testing"

func TestNewState(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(state.Guilds) == 0 {
		t.Fatal("Empty guilds state")
	}
}
