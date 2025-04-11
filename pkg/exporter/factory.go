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
	Exporter    Exporter
	AlwaysStore Exporter
	Nodes       *nodes.NodeManager
	Queries     *queries.Queries
}

// NewTLSExporter creates a new exporter for the TLS service
func NewTLSExporter(cfg config.ServiceFlagParams, mgr *settings.Settings, nodes *nodes.NodeManager, queries *queries.Queries) (*TLSExporter, error) {
	var mainExporter Exporter
	var err error

	// Create the primary exporter based on the configuration
	switch cfg.ConfigValues.Logger {
	case config.LoggingStdout:
		mainExporter = NewStdoutExporter()
	case config.LoggingS3:
		var s3Exporter *S3Exporter
		if cfg.S3LogConfig.Bucket != "" {
			// Use provided S3 config
			s3Exporter, err = NewS3Exporter(cfg.S3LogConfig)
		} else {
			// Load from file
			s3Exporter, err = NewS3ExporterFromFile(cfg.LoggerFile)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create S3 exporter: %w", err)
		}
		mainExporter = s3Exporter
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

	// Set up always store exporter if needed
	if cfg.AlwaysLog {
		// This would be a DB exporter in the actual implementation
		// For now we'll just use stdout as a placeholder
		tls.AlwaysStore = NewStdoutExporter()
		if err := tls.AlwaysStore.Configure(mgr); err != nil {
			return nil, fmt.Errorf("failed to configure always-store exporter: %w", err)
		}
	}

	return tls, nil
}

// Export sends data to configured exporters
func (e *TLSExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	var errs []error

	// Primary exporter
	if e.Exporter != nil && e.Exporter.IsEnabled() {
		if err := e.Exporter.Export(ctx, logType, data, environment, uuid); err != nil {
			errs = append(errs, err)
		}
	}

	// For status logs, also send to always-store if configured
	if e.AlwaysStore != nil && e.AlwaysStore.IsEnabled() && logType == StatusLog {
		if err := e.AlwaysStore.Export(ctx, logType, data, environment, uuid); err != nil {
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

	// Primary exporter
	if e.Exporter != nil && e.Exporter.IsEnabled() {
		if err := e.Exporter.ExportQuery(ctx, data, environment, uuid, name, status); err != nil {
			errs = append(errs, err)
		}
	}

	// Also send to always-store if configured
	if e.AlwaysStore != nil && e.AlwaysStore.IsEnabled() {
		if err := e.AlwaysStore.ExportQuery(ctx, data, environment, uuid, name, status); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("export query errors: %v", errs)
	}
	return nil
}
