package exporter

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/rs/zerolog/log"
)

// MultiExporter implements the Exporter interface to send logs to multiple destinations
type MultiExporter struct {
	BaseExporter
	exporters []Exporter
	mu        sync.RWMutex
}

// NewMultiExporter creates a new multi-destination exporter
func NewMultiExporter(exporters ...Exporter) *MultiExporter {
	return &MultiExporter{
		BaseExporter: NewBaseExporter("multi"),
		exporters:    exporters,
	}
}

// Export sends osquery data to all configured exporters
func (e *MultiExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	if !e.IsEnabled() {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var errors []error
	var wg sync.WaitGroup

	for _, exporter := range e.exporters {
		if !exporter.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(exp Exporter) {
			defer wg.Done()
			if err := exp.Export(ctx, logType, data, environment, uuid); err != nil {
				log.Error().
					Err(err).
					Str("exporter", exp.Name()).
					Msg("Failed to export data")
				errors = append(errors, fmt.Errorf("exporter %s: %w", exp.Name(), err))
			}
		}(exporter)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("export errors: %v", errors)
	}

	return nil
}

// ExportQuery sends osquery query results to all configured exporters
func (e *MultiExporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
	if !e.IsEnabled() {
		return nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var errors []error
	var wg sync.WaitGroup

	for _, exporter := range e.exporters {
		if !exporter.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(exp Exporter) {
			defer wg.Done()
			if err := exp.ExportQuery(ctx, data, environment, uuid, name, status); err != nil {
				log.Error().
					Err(err).
					Str("exporter", exp.Name()).
					Msg("Failed to export query data")
				errors = append(errors, fmt.Errorf("exporter %s: %w", exp.Name(), err))
			}
		}(exporter)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("export query errors: %v", errors)
	}

	return nil
}

// Configure sets up all exporters with necessary settings
func (e *MultiExporter) Configure(mgr *settings.Settings) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var errors []error

	for _, exporter := range e.exporters {
		if err := exporter.Configure(mgr); err != nil {
			errors = append(errors, fmt.Errorf("exporter %s: %w", exporter.Name(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration errors: %v", errors)
	}

	return nil
}

// AddExporter adds a new exporter to the multi-exporter
func (e *MultiExporter) AddExporter(exporter Exporter) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.exporters = append(e.exporters, exporter)
}
