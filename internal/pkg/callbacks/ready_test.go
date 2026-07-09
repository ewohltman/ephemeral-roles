package callbacks_test

import (
	"testing"

	"github.com/disgoorg/disgo/events"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestHandler_Ready(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:          log,
		RolePrefix:   "testRolePrefix",
		ReadyCounter: monitor.ReadyCounter(&monitor.Config{Log: log}),
	}

	handler.Ready(&events.Ready{
		GenericEvent: events.NewGenericEvent(session, 0, 0),
	})
}
