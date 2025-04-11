# Migration Guide: From `logging` to `exporter` Package

This guide explains how to migrate from the old `logging` package to the new `exporter` package in osctrl.

## Overview

The osctrl project is moving from the `logging` package to the new `exporter` package to:

1. Clarify the purpose of the functionality (exporting osquery data rather than traditional application logging)
2. Improve code maintainability by removing type switching and type assertions
3. Support multiple export destinations through a clean interface-based design
4. Add proper context support for better cancellation and deadline handling

## Key Changes

| Old (logging)                   | New (exporter)                      | Notes                                              |
|--------------------------------|--------------------------------------|---------------------------------------------------|
| `LoggerTLS`                     | `TLSExporter`                       | Main entry point for exporting data                |
| `LoggerDB`, `LoggerS3`, etc.    | `*Exporter` implementing `Exporter` | Each exporter now follows a consistent interface   |
| `Log()` and `QueryLog()`        | `Export()` and `ExportQuery()`      | Methods now return errors and accept context       |
| Type switching with assertions  | Interface-based polymorphism        | More maintainable and easier to extend             |
| No multi-destination support    | `MultiExporter` for multiple outputs| Send data to multiple destinations concurrently    |

## Migration Steps

### Step 1: Replace Logger Creation

**Before:**
```go
logger, err := logging.CreateLoggerTLS(cfg, mgr, nodes, queries)
```

**After:**
```go
exporter, err := exporter.NewTLSExporter(cfg, mgr, nodes, queries)
```

### Step 2: Update Export Calls

**Before:**
```go
logger.Log(types.StatusLog, data, environment, uuid, debug)
logger.QueryLog(types.ResultLog, data, environment, uuid, name, status, debug)
```

**After:**
```go
// Create a context
ctx := context.Background() // or use an existing context

// Handle potential errors
err := exporter.Export(ctx, exporter.StatusLog, data, environment, uuid)
if err != nil {
    log.Error().Err(err).Msg("Failed to export status log")
}

err = exporter.ExportQuery(ctx, data, environment, uuid, name, status)
if err != nil {
    log.Error().Err(err).Msg("Failed to export query data")
}
```

### Step 3: Handling Multiple Export Destinations

If you need to export to multiple destinations:

```go
// Create individual exporters
stdoutExporter := exporter.NewStdoutExporter()
s3Exporter, err := exporter.NewS3Exporter(s3Config)
if err != nil {
    return err
}

// Combine them with MultiExporter
multiExporter := exporter.NewMultiExporter(stdoutExporter, s3Exporter)

// Use as a normal exporter
err = multiExporter.Export(ctx, exporter.StatusLog, data, environment, uuid)
```

## Implementation Examples

### Example: Replacing a TLS Handler

**Before:**
```go
func (t *TLSHandler) HandleEnrollRequest() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... existing code ...
        
        // Log enrollment
        t.Logger.Log(types.StatusLog, jsonStatus, envName, nodeKey, t.DebugService)
        
        // ... existing code ...
    }
}
```

**After:**
```go
func (t *TLSHandler) HandleEnrollRequest() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... existing code ...
        
        // Export enrollment data
        ctx := r.Context()
        if err := t.Exporter.Export(ctx, exporter.StatusLog, jsonStatus, envName, nodeKey); err != nil {
            log.Error().Err(err).Msg("Failed to export enrollment status")
        }
        
        // ... existing code ...
    }
}
```

### Example: Implementing a Custom Exporter

If you have a custom logger, implement the `Exporter` interface:

```go
type CustomExporter struct {
    exporter.BaseExporter
    client *CustomClient
}

func NewCustomExporter(config CustomConfig) *CustomExporter {
    return &CustomExporter{
        BaseExporter: exporter.NewBaseExporter("custom"),
        client: NewCustomClient(config),
    }
}

func (e *CustomExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
    if !e.IsEnabled() {
        return nil
    }
    
    // Your custom export logic here
    return e.client.Send(ctx, data)
}

func (e *CustomExporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
    if !e.IsEnabled() {
        return nil
    }
    
    // Your custom query export logic here
    return e.client.SendQuery(ctx, data, name, status)
}

func (e *CustomExporter) Configure(mgr *settings.Settings) error {
    // Your configuration logic here
    return nil
}
```

## Service Integration

For each service using the logger, update the dependencies:

1. Replace `Logger` fields with `Exporter`
2. Update function signatures to pass contexts
3. Handle returned errors

## Testing

Test both individual exporters and combined exporters:

```go
func TestS3Exporter(t *testing.T) {
    s3Config := config.S3Configuration{/* config */}
    exporter, err := exporter.NewS3Exporter(s3Config)
    require.NoError(t, err)
    
    ctx := context.Background()
    err = exporter.Export(ctx, exporter.StatusLog, []byte("test"), "test-env", "test-uuid")
    require.NoError(t, err)
}
```

## Checklist

- [ ] Replace `Logger` creation with `Exporter` creation
- [ ] Update all export calls to use new method signatures
- [ ] Add error handling for export operations
- [ ] Pass contexts to export methods
- [ ] Update documentation and comments
- [ ] Update tests to use new exporter interfaces
- [ ] Consider using `MultiExporter` for multiple destinations

## Advantages of the New System

1. **Cleaner Code**: Interface-based design eliminates type switching and assertions
2. **Multiple Destinations**: Send data to multiple exporters concurrently
3. **Better Error Handling**: All methods return errors for proper handling
4. **Context Support**: Proper cancellation and deadline handling
5. **Testability**: Easier to mock and test with interfaces
6. **Extensibility**: Implement the interface to add new export destinations

## Getting Help

If you encounter issues during migration, please:
1. Check this guide for common migration patterns
2. Look at the implementations in the `exporter` package for examples
3. File an issue in the GitHub repository with specific questions