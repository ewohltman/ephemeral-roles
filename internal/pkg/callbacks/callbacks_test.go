package callbacks_test

import (
	"fmt"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestHandler_RoleNameFromChannel(t *testing.T) {
	const boyKeyword = "{eph}"

	expected := fmt.Sprintf("%s %s", boyKeyword, mock.TestChannel)
	handler := &callbacks.Handler{RolePrefix: "{eph}"}

	actual := handler.RoleNameFromChannel(mock.TestChannel)

	if actual != expected {
		t.Errorf("unexpected role name: %s", actual)
	}
}
