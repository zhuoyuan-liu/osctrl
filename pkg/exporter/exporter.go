// Package exporter provides interfaces and implementations for exporting osquery data
// to various destinations like S3, Elasticsearch, file storage, etc.
package exporter

import (
	"context"

	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/jmpsec/osctrl/pkg/types"
)

// LogType constants for the different types of logs we export
const (
	StatusLog = "status"
	ResultLog = "result"
)

// Exporter defines the interface for exporting osquery data
type Exporter interface {
	// Export sends osquery data to the destination
	Export(ctx context.Context, logType string, data []byte, environment, uuid string) error

	// ExportQuery sends osquery query results to the destination
	ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error

	// Configure sets up the exporter with necessary settings
	Configure(mgr *settings.Settings) error

	// IsEnabled returns whether the exporter is enabled
	IsEnabled() bool

	// SetEnabled enables or disables the exporter
	SetEnabled(enabled bool)

	// Name returns the name of the exporter
	Name() string
}

// BaseExporter provides common functionality for all exporters
type BaseExporter struct {
	enabled bool
	name    string
}

// NewBaseExporter creates a new base exporter
func NewBaseExporter(name string) BaseExporter {
	return BaseExporter{
		enabled: true,
		name:    name,
	}
}

// IsEnabled returns whether the exporter is enabled
func (e *BaseExporter) IsEnabled() bool {
	return e.enabled
}

// SetEnabled enables or disables the exporter
func (e *BaseExporter) SetEnabled(enabled bool) {
	e.enabled = enabled
}

// Name returns the name of the exporter
func (e *BaseExporter) Name() string {
	return e.name
}

// StatusData represents the data to be exported for status logs
type StatusData struct {
	Environment string
	UUID        string
	Data        []byte
}

// ResultData represents the data to be exported for result logs
type ResultData struct {
	Environment string
	UUID        string
	Data        []byte
}

// QueryData represents the data to be exported for query logs
type QueryData struct {
	Environment string
	UUID        string
	Name        string
	Status      int
	Data        []byte
}

// GetLogTypeString returns the string representation of a log type
func GetLogTypeString(logType string) string {
	switch logType {
	case types.StatusLog:
		return StatusLog
	case types.ResultLog:
		return ResultLog
	default:
		return logType
	}
}
