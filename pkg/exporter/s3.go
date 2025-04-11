package exporter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/settings"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Exporter implements the Exporter interface for AWS S3
type S3Exporter struct {
	BaseExporter
	s3Config  config.S3Configuration
	awsConfig aws.Config
	client    *s3.Client
	uploader  *manager.Uploader
	debug     bool
}

// NewS3Exporter creates a new S3 exporter from config
func NewS3Exporter(s3Config config.S3Configuration) (*S3Exporter, error) {
	ctx := context.Background()
	creds := credentials.NewStaticCredentialsProvider(s3Config.AccessKey, s3Config.SecretAccessKey, "")
	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithRegion(s3Config.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	return &S3Exporter{
		BaseExporter: NewBaseExporter("s3"),
		s3Config:     s3Config,
		awsConfig:    cfg,
		client:       client,
		uploader:     uploader,
		debug:        false,
	}, nil
}

// NewS3ExporterFromFile creates a new S3 exporter from config file
func NewS3ExporterFromFile(s3File string) (*S3Exporter, error) {
	s3Config, err := loadS3Config(s3File)
	if err != nil {
		return nil, err
	}
	return NewS3Exporter(s3Config)
}

// loadS3Config loads S3 configuration from a file
func loadS3Config(file string) (config.S3Configuration, error) {
	var s3Cfg config.S3Configuration
	log.Info().Msgf("Loading S3 config from %s", file)

	// Load file and read config
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return s3Cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	cfgRaw := viper.Sub(config.LoggingS3)
	if cfgRaw == nil {
		return s3Cfg, fmt.Errorf("JSON key %s not found in %s", config.LoggingS3, file)
	}

	if err := cfgRaw.Unmarshal(&s3Cfg); err != nil {
		return s3Cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return s3Cfg, nil
}

// Configure sets up the exporter with necessary settings
func (e *S3Exporter) Configure(mgr *settings.Settings) error {
	log.Info().Msg("Configuring S3 exporter")
	// Can add more configuration logic here if needed
	return nil
}

// SetDebug enables or disables debug logging
func (e *S3Exporter) SetDebug(debug bool) {
	e.debug = debug
}

// Export sends osquery data to S3
func (e *S3Exporter) Export(ctx context.Context, logType string, data []byte, environment, uuid string) error {
	if !e.IsEnabled() {
		return nil
	}

	if e.debug {
		log.Debug().
			Str("type", logType).
			Str("environment", environment).
			Str("uuid", uuid).
			Int("size", len(data)).
			Msg("Sending data to S3")
	}

	// Create key path with timestamp to ensure uniqueness
	key := fmt.Sprintf("%s/%s/%s:%s.json",
		environment,
		logType,
		uuid,
		strconv.FormatInt(time.Now().UnixMilli(), 10),
	)

	contentLength := int64(len(data))

	result, err := e.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(e.s3Config.Bucket),
		Key:           aws.String(key),
		Body:          bytes.NewBuffer(data),
		ContentLength: &contentLength,
		ContentType:   aws.String(http.DetectContentType(data)),
	})

	if err != nil {
		return fmt.Errorf("failed to upload data to S3: %w", err)
	}

	if e.debug {
		log.Debug().
			Str("location", result.Location).
			Str("key", key).
			Msg("Successfully uploaded to S3")
	}

	return nil
}

// ExportQuery sends osquery query results to S3
func (e *S3Exporter) ExportQuery(ctx context.Context, data []byte, environment, uuid, name string, status int) error {
	if !e.IsEnabled() {
		return nil
	}

	if e.debug {
		log.Debug().
			Str("environment", environment).
			Str("uuid", uuid).
			Str("name", name).
			Int("status", status).
			Int("size", len(data)).
			Msg("Sending query result to S3")
	}

	// For query results, include query name and status in the key path
	key := fmt.Sprintf("%s/query/%s:%s:%d:%s.json",
		environment,
		name,
		uuid,
		status,
		strconv.FormatInt(time.Now().UnixMilli(), 10),
	)

	contentLength := int64(len(data))

	result, err := e.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(e.s3Config.Bucket),
		Key:           aws.String(key),
		Body:          bytes.NewBuffer(data),
		ContentLength: &contentLength,
		ContentType:   aws.String(http.DetectContentType(data)),
	})

	if err != nil {
		return fmt.Errorf("failed to upload query result to S3: %w", err)
	}

	if e.debug {
		log.Debug().
			Str("location", result.Location).
			Str("key", key).
			Msg("Successfully uploaded query result to S3")
	}

	return nil
}
