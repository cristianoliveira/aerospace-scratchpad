package cli_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cristianoliveira/aerospace-scratchpad/internal/cli"
)

func TestOutputFormatter_Text(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter, err := cli.NewOutputFormatter(buf, "text")
	if err != nil {
		t.Fatalf("unexpected error creating formatter: %v", err)
	}

	event := cli.OutputEvent{
		Command:         "move",
		Action:          "to-scratchpad",
		WindowID:        1234,
		AppName:         "Finder",
		Workspace:       "ws1",
		TargetWorkspace: ".scratchpad",
		Result:          "ok",
		Message:         "done",
	}

	if err = formatter.Print(event); err != nil {
		t.Fatalf("unexpected error printing event: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	expected := `command=move action=to-scratchpad window_id=1234 app_name=Finder workspace=ws1 target_workspace=.scratchpad result=ok message=done`
	if got != expected {
		t.Fatalf("text output mismatch:\nwant: %s\ngot:  %s", expected, got)
	}
}

func TestOutputFormatter_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter, err := cli.NewOutputFormatter(buf, "json")
	if err != nil {
		t.Fatalf("unexpected error creating formatter: %v", err)
	}

	event := cli.OutputEvent{
		Command:         "show",
		Action:          "focus",
		WindowID:        42,
		AppName:         "Terminal",
		Workspace:       "ws2",
		TargetWorkspace: "",
		Result:          "ok",
		Message:         "",
	}

	if err = formatter.Print(event); err != nil {
		t.Fatalf("unexpected error printing event: %v", err)
	}

	var decoded cli.OutputEvent
	if err = json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to unmarshal json output: %v", err)
	}

	if decoded != event {
		t.Fatalf("json output mismatch:\nwant: %+v\ngot:  %+v", event, decoded)
	}
}

func TestOutputFormatter_TSVAndCSV(t *testing.T) {
	event := cli.OutputEvent{
		Command:         "summon",
		Action:          "to-workspace",
		WindowID:        99,
		AppName:         "Notes",
		Workspace:       "ws3",
		TargetWorkspace: "ws4",
		Result:          "ok",
		Message:         "focused",
	}

	tests := []struct {
		name   string
		format string
		sep    rune
	}{
		{name: "tsv", format: "tsv", sep: '\t'},
		{name: "csv", format: "csv", sep: ','},
	}

	expectedRow := []string{
		"summon",
		"to-workspace",
		"99",
		"Notes",
		"ws3",
		"ws4",
		"ok",
		"focused",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter, err := cli.NewOutputFormatter(buf, tc.format)
			if err != nil {
				t.Fatalf("unexpected error creating formatter: %v", err)
			}

			if err = formatter.Print(event); err != nil {
				t.Fatalf("unexpected error printing event: %v", err)
			}

			reader := csv.NewReader(strings.NewReader(buf.String()))
			reader.Comma = tc.sep
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("failed to parse %s output: %v", tc.name, err)
			}

			if len(rows) != 2 {
				t.Fatalf("expected header + one row, got %d rows", len(rows))
			}

			expectedHeader := []string{
				"command",
				"action",
				"window_id",
				"app_name",
				"workspace",
				"target_workspace",
				"result",
				"message",
			}
			if !equalStringSlices(rows[0], expectedHeader) {
				t.Fatalf("header mismatch:\nwant: %v\ngot:  %v", expectedHeader, rows[0])
			}

			if !equalStringSlices(rows[1], expectedRow) {
				t.Fatalf("row mismatch:\nwant: %v\ngot:  %v", expectedRow, rows[1])
			}
		})
	}
}

func TestOutputFormatter_EmptyFieldsQuotedInText(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter, err := cli.NewOutputFormatter(buf, "text")
	if err != nil {
		t.Fatalf("unexpected error creating formatter: %v", err)
	}

	event := cli.OutputEvent{
		Command:         "move",
		Action:          "to-scratchpad",
		WindowID:        0,
		AppName:         "",
		Workspace:       "",
		TargetWorkspace: "",
		Result:          "ok",
		Message:         "",
	}

	if err = formatter.Print(event); err != nil {
		t.Fatalf("unexpected error printing event: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	expected := `command=move action=to-scratchpad window_id=0 app_name="" workspace="" target_workspace="" result=ok message=""`
	if got != expected {
		t.Fatalf("text output mismatch for empty fields:\nwant: %s\ngot:  %s", expected, got)
	}
}

func TestOutputFormatter_EmptyFieldsInSeparatedOutputs(t *testing.T) {
	event := cli.OutputEvent{
		Command:         "move",
		Action:          "to-scratchpad",
		WindowID:        0,
		AppName:         "",
		Workspace:       "",
		TargetWorkspace: "",
		Result:          "ok",
		Message:         "",
	}

	tests := []struct {
		name   string
		format string
		sep    rune
	}{
		{name: "tsv_empty_fields", format: "tsv", sep: '\t'},
		{name: "csv_empty_fields", format: "csv", sep: ','},
	}

	expectedRow := []string{
		"move",
		"to-scratchpad",
		"0",
		"",
		"",
		"",
		"ok",
		"",
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter, err := cli.NewOutputFormatter(buf, tc.format)
			if err != nil {
				t.Fatalf("unexpected error creating formatter: %v", err)
			}

			if err = formatter.Print(event); err != nil {
				t.Fatalf("unexpected error printing event: %v", err)
			}

			reader := csv.NewReader(strings.NewReader(buf.String()))
			reader.Comma = tc.sep
			rows, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("failed to parse %s output: %v", tc.name, err)
			}

			if len(rows) != 2 {
				t.Fatalf("expected header + one row, got %d rows", len(rows))
			}

			if !equalStringSlices(rows[1], expectedRow) {
				t.Fatalf("row mismatch:\nwant: %v\ngot:  %v", expectedRow, rows[1])
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
