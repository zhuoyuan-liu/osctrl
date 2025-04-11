# Exporter Package

The `exporter` package provides interfaces and implementations for exporting osquery data to various destinations in osctrl.

## Overview

This package replaces the original `logging` package to better reflect its purpose: exporting osquery data to various destinations rather than traditional application logging. It introduces a cleaner, interface-based design that eliminates type switching and allows for multiple export destinations.

## Key Components

- **Exporter Interface**: The core interface that all exporters implement
- **BaseExporter**: Common functionality for all exporters
- **StdoutExporter**: Example exporter that outputs to standard output
- **S3Exporter**: Exporter that sends data to AWS S3
- **MultiExporter**: Composite exporter that sends data to multiple destinations
- **TLSExporter**: Main entry point for the TLS service

## Using the Exporter

### Basic Usage

```go
// Create an exporter
stdoutExporter := exporter.NewStdoutExporter()

// Export data
ctx := context.Background()
err := stdoutExporter.Export(ctx, exporter.StatusLog, data, environment, uuid)
if err != nil {
    log.Error().Err(err).Msg("Failed to export data")
}

// Export query results
err = stdoutExporter.ExportQuery(ctx, data, environment, uuid, queryName, status)
if err != nil {
    log.Error().Err(err).Msg("Failed to export query results")
}
```

### Multiple Destinations

```go
// Create individual exporters
stdoutExporter := exporter.NewStdoutExporter()
s3Exporter, err := exporter.NewS3Exporter(s3Config)

// Create a multi-exporter
multiExporter := exporter.NewMultiExporter(stdoutExporter, s3Exporter)

// Use like a normal exporter
err = multiExporter.Export(ctx, exporter.StatusLog, data, environment, uuid)
```

### Service Integration

The TLS service should create a TLSExporter:

```go
exporter, err := exporter.NewTLSExporter(cfg, mgr, nodes, queries)
if err != nil {
    return err
}
```

## Creating a Custom Exporter

To create a new exporter, implement the `Exporter` interface:

```go
type CustomExporter struct {
    exporter.BaseExporter
    // Additional fields
}

func NewCustomExporter(config CustomConfig) *CustomExporter {
    return &CustomExporter{
        BaseExporter: exporter.NewBaseExporter("custom"),
        // Initialize additional fields
    }
}

// Implement Export, ExportQuery, and Configure methods
```

See the StdoutExporter and S3Exporter implementations for examples.

## Migration

If you're migrating from the old `logging` package, please see the migration guide at `/docs/migration-guide-exporter.md`.