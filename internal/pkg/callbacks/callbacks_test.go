package callbacks_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	rolePrefix  = "{eph}"
	testBotName = "testBot"
)

func TestHandler_RoleNameFromChannel(t *testing.T) {
	t.Parallel()

	handler := &callbacks.Handler{RolePrefix: rolePrefix}
	expected := fmt.Sprintf("%s %s", rolePrefix, mock.TestChannelName)
	actual := handler.RoleNameFromChannel(mock.TestChannelName)

	assert.Equal(t, expected, actual)
}
