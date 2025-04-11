package exporter

import (
	"context"
	"fmt"

	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/rs/zerolog/log"
)

// StdoutExporter implements the Exporter interface for stdout
type StdoutExporter struct {
	BaseExporter
}

// NewStdoutExporter creates a new stdout exporter
func NewStdoutExporter() *StdoutExporter {
	return &StdoutExporter{
		BaseExporter: NewBaseExporter("stdout"),
	}
}

// Export sends osquery data to stdout
func (e *StdoutExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	if !e.IsEnabled() {
		return nil
	}

	switch logType {
	case StatusLog:
		return e.exportStatus(ctx, data, environment, uuid)
	case ResultLog:
		return e.exportResult(ctx, data, environment, uuid)
	default:
		return fmt.Errorf("unknown log type: %s", logType)
	}
}

// exportStatus sends status logs to stdout
func (e *StdoutExporter) exportStatus(ctx context.Context, data []byte, environment, uuid string) error {
	log.Info().
		Str("type", "status").
		Str("environment", environment).
		Str("uuid", uuid).
		Int("size", len(data)).
		Msgf("Status: %s:%s - %d bytes [%s]", environment, uuid, len(data), string(data))
	return nil
}

// exportResult sends result logs to stdout
func (e *StdoutExporter) exportResult(ctx context.Context, data []byte, environment, uuid string) error {
	log.Info().
		Str("type", "result").
		Str("environment", environment).
		Str("uuid", uuid).
		Int("size", len(data)).
		Msgf("Result: %s:%s - %d bytes [%s]", environment, uuid, len(data), string(data))
	return nil
}

// ExportQuery sends osquery query results to stdout
func (e *StdoutExporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
	if !e.IsEnabled() {
		return nil
	}

	log.Info().
		Str("type", "query").
		Str("environment", environment).
		Str("uuid", uuid).
		Str("name", name).
		Int("status", status).
		Int("size", len(data)).
		Msgf("Query: %s:%d - %s:%s - %d bytes [%s]", name, status, environment, uuid, len(data), string(data))
	return nil
}

// Configure sets up the exporter with necessary settings
func (e *StdoutExporter) Configure(mgr *settings.Settings) error {
	log.Info().Msg("No configuration needed for stdout exporter")
	return nil
}
