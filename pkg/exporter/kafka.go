package exporter

import (
	"context"
	"fmt"
	"os"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/tlscfg"
)

// KafkaExporter implements the Exporter interface for Apache Kafka
type KafkaExporter struct {
	BaseExporter
	config   config.KafkaConfiguration
	producer *kgo.Client
	debug    bool
}

// NewKafkaExporter creates a new Kafka exporter from config
func NewKafkaExporter(config config.KafkaConfiguration) (*KafkaExporter, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.BoostrapServer),
		kgo.ConsumeTopics(config.Topic),
	}

	if config.ConnectionTimeout > 0 {
		opts = append(opts, kgo.DialTimeout(config.ConnectionTimeout))
	}

	// Configure SASL authentication if enabled
	if config.SASL.Mechanism != "" {
		if config.SASL.Username == "" {
			return nil, fmt.Errorf("SASL mechanism requires a username")
		}

		if config.SASL.Password == "" {
			return nil, fmt.Errorf("SASL mechanism requires a password")
		}

		auth := scram.Auth{
			User: config.SASL.Username,
			Pass: config.SASL.Password,
		}

		var mechanism sasl.Mechanism
		switch config.SASL.Mechanism {
		case "SCRAM-SHA-512":
			mechanism = auth.AsSha512Mechanism()
		case "SCRAM-SHA-256":
			mechanism = auth.AsSha256Mechanism()
		default:
			return nil, fmt.Errorf("unknown SASL mechanism '%s'", config.SASL.Mechanism)
		}

		opts = append(opts, kgo.SASL(mechanism))
	}

	// Configure SSL if enabled
	if config.SSLCALocation != "" {
		caCert, err := os.ReadFile(config.SSLCALocation)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA Cert from '%s': %w", config.SSLCALocation, err)
		}

		cfg, err := tlscfg.New(tlscfg.WithCA(caCert, tlscfg.ForClient))
		if err != nil {
			return nil, fmt.Errorf("failed to use CA Cert for client side: %w", err)
		}

		opts = append(opts, kgo.DialTLSConfig(cfg))
	}

	// Create Kafka client
	producer, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &KafkaExporter{
		BaseExporter: NewBaseExporter("kafka"),
		config:       config,
		producer:     producer,
		debug:        false,
	}, nil
}

// SetDebug enables or disables debug logging
func (e *KafkaExporter) SetDebug(debug bool) {
	e.debug = debug
}

// Configure sets up the exporter with necessary settings
func (e *KafkaExporter) Configure(mgr *settings.Settings) error {
	log.Info().Msg("Configuring Kafka exporter")
	// Add any additional configuration logic here if needed
	return nil
}

// Close closes the Kafka producer client
func (e *KafkaExporter) Close() error {
	e.producer.Close()
	return nil
}

// Export sends osquery data to Kafka
func (e *KafkaExporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	if !e.IsEnabled() {
		return nil
	}

	if e.debug {
		log.Debug().
			Str("type", logType).
			Str("environment", environment).
			Str("uuid", uuid).
			Str("topic", e.config.Topic).
			Int("size", len(data)).
			Msg("Sending data to Kafka")
	}

	// Use UUID as key for consistent partitioning
	key := []byte(uuid)

	// Add metadata to record headers if needed
	headers := []kgo.RecordHeader{
		{Key: "logType", Value: []byte(logType)},
		{Key: "environment", Value: []byte(environment)},
	}

	// Create and send record
	record := kgo.Record{
		Topic:   e.config.Topic,
		Key:     key,
		Value:   data,
		Headers: headers,
	}

	// Use a channel to wait for the produce result
	resultChan := make(chan error, 1)

	e.producer.Produce(ctx, &record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Error().
				Err(err).
				Str("topic", e.config.Topic).
				Str("uuid", uuid).
				Msg("Failed to produce message to Kafka")
			resultChan <- fmt.Errorf("failed to produce message to Kafka: %w", err)
		} else {
			if e.debug {
				log.Debug().
					Str("topic", e.config.Topic).
					Str("uuid", uuid).
					Str("partition", fmt.Sprintf("%d", r.Partition)).
					Str("offset", fmt.Sprintf("%d", r.Offset)).
					Msg("Successfully sent message to Kafka")
			}
			resultChan <- nil
		}
	})

	// Wait for the callback or context cancellation
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ExportQuery sends osquery query results to Kafka
func (e *KafkaExporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
	if !e.IsEnabled() {
		return nil
	}

	if e.debug {
		log.Debug().
			Str("environment", environment).
			Str("uuid", uuid).
			Str("name", name).
			Int("status", status).
			Str("topic", e.config.Topic).
			Int("size", len(data)).
			Msg("Sending query result to Kafka")
	}

	// Use UUID as key for consistent partitioning
	key := []byte(uuid)

	// Add metadata to record headers
	headers := []kgo.RecordHeader{
		{Key: "logType", Value: []byte("query")},
		{Key: "environment", Value: []byte(environment)},
		{Key: "queryName", Value: []byte(name)},
		{Key: "status", Value: []byte(fmt.Sprintf("%d", status))},
	}

	// Create and send record
	record := kgo.Record{
		Topic:   e.config.Topic,
		Key:     key,
		Value:   data,
		Headers: headers,
	}

	// Use a channel to wait for the produce result
	resultChan := make(chan error, 1)

	e.producer.Produce(ctx, &record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Error().
				Err(err).
				Str("topic", e.config.Topic).
				Str("uuid", uuid).
				Str("name", name).
				Msg("Failed to produce query result to Kafka")
			resultChan <- fmt.Errorf("failed to produce query result to Kafka: %w", err)
		} else {
			if e.debug {
				log.Debug().
					Str("topic", e.config.Topic).
					Str("uuid", uuid).
					Str("name", name).
					Str("partition", fmt.Sprintf("%d", r.Partition)).
					Str("offset", fmt.Sprintf("%d", r.Offset)).
					Msg("Successfully sent query result to Kafka")
			}
			resultChan <- nil
		}
	})

	// Wait for the callback or context cancellation
	select {
	case err := <-resultChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
