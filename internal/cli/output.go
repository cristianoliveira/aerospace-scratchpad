package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatTSV  OutputFormat = "tsv"
	OutputFormatCSV  OutputFormat = "csv"
)

// OutputEvent describes a single command result in a structured way.
type OutputEvent struct {
	Command         string `json:"command"`
	Action          string `json:"action"`
	WindowID        int    `json:"window_id"`
	AppName         string `json:"app_name"`
	Workspace       string `json:"workspace"`
	TargetWorkspace string `json:"target_workspace"`
	Result          string `json:"result"`
	Message         string `json:"message"`
}

// OutputFormatter writes events in a script-friendly format.
type OutputFormatter struct {
	format        OutputFormat
	writer        io.Writer
	headerWritten bool
}

func NewOutputFormatter(w io.Writer, format string) (*OutputFormatter, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case string(OutputFormatText):
		return &OutputFormatter{format: OutputFormatText, writer: w}, nil
	case string(OutputFormatJSON):
		return &OutputFormatter{format: OutputFormatJSON, writer: w}, nil
	case string(OutputFormatTSV):
		return &OutputFormatter{format: OutputFormatTSV, writer: w}, nil
	case string(OutputFormatCSV):
		return &OutputFormatter{format: OutputFormatCSV, writer: w}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}
}

func (f *OutputFormatter) Print(event OutputEvent) error {
	switch f.format {
	case OutputFormatJSON:
		return f.printJSON(event)
	case OutputFormatTSV:
		return f.printSeparated(event, '\t')
	case OutputFormatCSV:
		return f.printSeparated(event, ',')
	case OutputFormatText:
		return f.printText(event)
	default:
		return fmt.Errorf("unsupported output format: %s", f.format)
	}
}

func (f *OutputFormatter) printText(event OutputEvent) error {
	values := f.rowValues(event)
	parts := make([]string, 0, len(outputHeaders))
	for i, header := range outputHeaders {
		parts = append(parts, fmt.Sprintf("%s=%s", header, quoteIfNeeded(values[i])))
	}

	_, err := fmt.Fprintln(f.writer, strings.Join(parts, " "))
	return err
}

func (f *OutputFormatter) printJSON(event OutputEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.writer, string(data))
	return err
}

func (f *OutputFormatter) printSeparated(event OutputEvent, sep rune) error {
	writer := csv.NewWriter(f.writer)
	writer.Comma = sep

	if !f.headerWritten {
		if err := writer.Write(outputHeaders); err != nil {
			return err
		}
		f.headerWritten = true
	}

	if err := writer.Write(f.rowValues(event)); err != nil {
		return err
	}
	writer.Flush()

	return writer.Error()
}

func (f *OutputFormatter) rowValues(event OutputEvent) []string {
	return []string{
		event.Command,
		event.Action,
		strconv.Itoa(event.WindowID),
		event.AppName,
		event.Workspace,
		event.TargetWorkspace,
		event.Result,
		event.Message,
	}
}

func quoteIfNeeded(value string) string {
	if value == "" {
		return "\"\""
	}

	if strings.ContainsAny(value, " \t\"") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}

	return value
}

var outputHeaders = []string{ //nolint:gochecknoglobals // shared header ordering for all formats
	"command",
	"action",
	"window_id",
	"app_name",
	"workspace",
	"target_workspace",
	"result",
	"message",
}
