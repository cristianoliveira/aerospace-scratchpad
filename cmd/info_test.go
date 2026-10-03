package cmd_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/focus"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/layout"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/monitors"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/windows"
	"github.com/cristianoliveira/aerospace-ipc/pkg/aerospace/workspaces"
	"github.com/cristianoliveira/aerospace-ipc/pkg/client"
	"github.com/cristianoliveira/aerospace-scratchpad/cmd"
	client_mock "github.com/cristianoliveira/aerospace-scratchpad/internal/mocks/client"
	"github.com/cristianoliveira/aerospace-scratchpad/internal/testutils"
)

type infoAeroSpaceClient struct {
	conn client.AeroSpaceConnection
}

func (m *infoAeroSpaceClient) Windows() *windows.Service {
	return nil
}

func (m *infoAeroSpaceClient) Workspaces() *workspaces.Service {
	return nil
}

func (m *infoAeroSpaceClient) Focus() *focus.Service {
	return nil
}

func (m *infoAeroSpaceClient) Layout() *layout.Service {
	return nil
}

func (m *infoAeroSpaceClient) Monitors() *monitors.Service {
	return nil
}

func (m *infoAeroSpaceClient) Connection() client.AeroSpaceConnection {
	return m.conn
}

func TestInfoCmd(t *testing.T) {
	t.Run("reports compatibility information", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			CheckServerVersion().
			Return(nil).
			Times(1)
		socket.EXPECT().
			GetSocketPath().
			Return("/tmp/aerospace.sock", nil).
			Times(1)
		socket.EXPECT().
			GetServerVersion().
			Return("0.4.0", nil).
			Times(1)

		args := []string{"info"}
		command := cmd.RootCmd(&infoAeroSpaceClient{conn: socket})
		command.SetArgs(args)
		output := &bytes.Buffer{}
		command.SetOut(output)
		command.SetErr(output)

		err := command.Execute()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
		testutils.MatchSnapshot(t, nil, cmdAsString, output.String(), err)
	})

	t.Run("reports incompatibility when version check fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			CheckServerVersion().
			Return(errors.New("mocked incompatibility")).
			Times(1)
		socket.EXPECT().
			GetSocketPath().
			Return("/tmp/aerospace.sock", nil).
			Times(1)
		socket.EXPECT().
			GetServerVersion().
			Return("0.4.0", nil).
			Times(1)

		args := []string{"info"}
		command := cmd.RootCmd(&infoAeroSpaceClient{conn: socket})
		command.SetArgs(args)
		output := &bytes.Buffer{}
		command.SetOut(output)
		command.SetErr(output)

		err := command.Execute()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
		testutils.MatchSnapshot(t, nil, cmdAsString, output.String(), err)
	})

	t.Run("still prints when compatibility fails but other calls error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		socket := client_mock.NewMockAeroSpaceConnection(ctrl)
		socket.EXPECT().
			CheckServerVersion().
			Return(errors.New("mocked incompatibility")).
			Times(1)
		socket.EXPECT().
			GetSocketPath().
			Return("", errors.New("mocked socket path failure")).
			Times(1)
		socket.EXPECT().
			GetServerVersion().
			Return("", errors.New("mocked server version failure")).
			Times(1)

		args := []string{"info"}
		command := cmd.RootCmd(&infoAeroSpaceClient{conn: socket})
		command.SetArgs(args)
		output := &bytes.Buffer{}
		command.SetOut(output)
		command.SetErr(output)

		err := command.Execute()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		cmdAsString := "aerospace-scratchpad " + strings.Join(args, " ")
		testutils.MatchSnapshot(t, nil, cmdAsString, output.String(), err)
	})
}
