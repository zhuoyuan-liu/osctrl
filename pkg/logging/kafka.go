package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/jmpsec/osctrl/pkg/types"

	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/tlscfg"
)

type LoggerKafka struct {
	config   config.KafkaConfiguration
	Enabled  bool
	producer *kgo.Client
}

func CreateLoggerKafka(config config.KafkaConfiguration) (*LoggerKafka, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.BoostrapServer),
		kgo.ConsumeTopics(config.Topic),
	}

	if config.ConnectionTimeout > 0 {
		kgo.DialTimeout(config.ConnectionTimeout)
	}

	// if we rely on SASL then populate the options
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

	if config.SSLCALocation != "" {
		caCert, err := os.ReadFile(config.SSLCALocation)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA Cert from '%s'", config.SSLCALocation)
		}
		cfg, err := tlscfg.New(tlscfg.WithCA(caCert, tlscfg.ForClient))
		if err != nil {
			return nil, fmt.Errorf("failed to use CA Cert for client side: %w", err)
		}
		opts = append(opts, kgo.DialTLSConfig(cfg))
	}

	producer, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	return &LoggerKafka{
		config:   config,
		Enabled:  true,
		producer: producer,
	}, nil
}

func (l *LoggerKafka) Settings(mgr *settings.Settings) {
	log.Warn().Msg("No kafka logging settings")
}

func (l *LoggerKafka) parseLogs(logType string, data []byte) ([]any, error) {
	// QueryLogs = distributed query results are delivered from OSQuery as a single JSON object.
	if logType == types.QueryLog {
		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("error parsing data %s: %w", string(data), err)
		}
		return []any{result}, nil
	}

	// others (Result + Status) are delivered in an array
	var logs []any
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("error parsing log %s: %w", string(data), err)
	}

	return logs, nil
}

func (l *LoggerKafka) Send(logType string, data []byte, environment, uuid string, debug bool) {
	if debug {
		log.Info().Msgf(
			"Sending %d bytes to Kafka topic %s for %s - %s",
			len(data), l.config.Topic, environment, uuid)
	}

	logs, err := l.parseLogs(logType, data)
	if err != nil {
		log.Err(err).Msg("failed to parse logs")
		return
	}

	ctx := context.Background()
	key := []byte(uuid) // uuid is the unique id of the os-query agent host that sent this data

	// Prepare all kafka records, so they can be sent in one batch
	var records []*kgo.Record
	for _, logEntry := range logs {
		jsonEvent, err := json.Marshal(logEntry)
		if err != nil {
			log.Err(err).Msg("Error parsing data")
			continue
		}

		r := &kgo.Record{Topic: l.config.Topic, Key: key, Value: jsonEvent}
		records = append(records, r)
	}

	if len(records) == 0 {
		log.Warn().Msgf("unexpected record count of 0 from %s:%s", uuid, environment)
		return
	}

	results := l.producer.ProduceSync(ctx, records...)
	if err := results.FirstErr(); err != nil {
		log.Err(err).Msgf("failed to produce messages to kafka topic '%s'", l.config.Topic)
	} else if debug {
		log.Info().Msgf("successfully sent %d %s messages to kafka topic '%s' from %s:%s",
			len(records), logType, l.config.Topic, uuid, environment)
	}
}
