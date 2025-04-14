package exporter

import (
	"context"
	"fmt"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/nodes"
	"github.com/jmpsec/osctrl/pkg/queries"
	"github.com/jmpsec/osctrl/pkg/settings"
)

// TLSExporter handles export operations for the TLS service
type TLSExporter struct {
	Exporter Exporter
	Nodes    *nodes.NodeManager
	Queries  *queries.Queries
}

// NewTLSExporter creates a new exporter for the TLS service
func NewTLSExporter(cfg config.ServiceFlagParams, mgr *settings.Settings, nodes *nodes.NodeManager, queries *queries.Queries) (*TLSExporter, error) {
	var mainExporter Exporter

	switch cfg.ConfigValues.Logger {
	case config.LoggingStdout:
		mainExporter = NewStdoutExporter()
	case config.LoggingKafka:
		kafkaExporter, err := NewKafkaExporter(cfg.KafkaConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kafka exporter: %w", err)
		}
		mainExporter = kafkaExporter
	// Other exporters would be added here as cases
	default:
		// Default to stdout if the specified logger is not supported yet
		mainExporter = NewStdoutExporter()
	}

	// Configure the exporter
	if err := mainExporter.Configure(mgr); err != nil {
		return nil, fmt.Errorf("failed to configure primary exporter: %w", err)
	}

	tls := &TLSExporter{
		Exporter: mainExporter,
		Nodes:    nodes,
		Queries:  queries,
	}

	return tls, nil
}

// Export sends data to configured exporters
func (e *TLSExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	var errs []error

	if e.Exporter != nil && e.Exporter.IsEnabled() {
		if err := e.Exporter.Export(ctx, logType, data, environment, uuid); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("export errors: %v", errs)
	}
	return nil
}

// ExportQuery sends query result to configured exporters
func (e *TLSExporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
	var errs []error

	if e.Exporter != nil && e.Exporter.IsEnabled() {
		if err := e.Exporter.ExportQuery(ctx, data, environment, uuid, name, status); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("export query errors: %v", errs)
	}
	return nil
}
