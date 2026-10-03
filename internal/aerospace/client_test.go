package aerospace_test

import (
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/aerospace"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

func TestAeroSpaceClientUsesSDKValidationForProvisioningCommands(t *testing.T) {
	t.Run("rejects an invalid monitor ordinal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		client := aerospace.NewAeroSpaceClient(testutils.NewMockAeroSpaceWM(ctrl))

		err := client.FocusMonitor(0)

		if err == nil || !strings.Contains(err.Error(), "monitor ordinal must be at least 1") {
			t.Fatalf("expected SDK monitor ordinal validation, got %v", err)
		}
	})

	t.Run("rejects a blank workspace name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		client := aerospace.NewAeroSpaceClient(testutils.NewMockAeroSpaceWM(ctrl))

		err := client.SummonWorkspace(" ")

		if err == nil || !strings.Contains(err.Error(), "workspace name must not be blank") {
			t.Fatalf("expected SDK workspace validation, got %v", err)
		}
	})
}
