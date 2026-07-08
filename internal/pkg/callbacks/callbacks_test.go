package callbacks_test

import (
	"fmt"
	"testing"

	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/stretchr/testify/assert"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
)

const (
	rolePrefix  = "{eph}"
	testBotName = "testBot"
)

func TestHandler_RoleNameFromChannel(t *testing.T) {
	t.Parallel()

	handler := &callbacks.Handler{RolePrefix: rolePrefix}
	expected := fmt.Sprintf("%s %s", rolePrefix, mockconstants.TestChannel)
	actual := handler.RoleNameFromChannel(mockconstants.TestChannel)

	assert.Equal(t, expected, actual)
}
