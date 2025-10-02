package logging

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmpsec/osctrl/pkg/config"
	"github.com/jmpsec/osctrl/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

type MockKafkaProducer struct {
	SentRecords []*kgo.Record
	Results     kgo.ProduceResults
}

func CreateTestableLoggerKafka(config config.KafkaConfiguration, mockProducer KafkaProducer) *LoggerKafka {
	return &LoggerKafka{
		config:   config,
		Enabled:  true,
		producer: mockProducer,
	}
}

func (m *MockKafkaProducer) ProduceSync(ctx context.Context, records ...*kgo.Record) kgo.ProduceResults {
	m.SentRecords = append(m.SentRecords, records...)
	return m.Results
}

func (m *MockKafkaProducer) GetSentValues() [][]byte {
	values := make([][]byte, len(m.SentRecords))
	for i, record := range m.SentRecords {
		values[i] = record.Value
	}
	return values
}

func (m *MockKafkaProducer) GetSentKeys() [][]byte {
	keys := make([][]byte, len(m.SentRecords))
	for i, record := range m.SentRecords {
		keys[i] = record.Key
	}
	return keys
}

func (m *MockKafkaProducer) GetSentTopics() []string {
	topics := make([]string, len(m.SentRecords))
	for i, record := range m.SentRecords {
		topics[i] = record.Topic
	}
	return topics
}

func TestLoggerKafka_Send_QueryLog(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	logger := CreateTestableLoggerKafka(config.KafkaConfiguration{
		Topic: "test-topic",
	}, mockProducer)

	queryData := map[string]interface{}{
		"query":  "SELECT * FROM processes",
		"status": 0,
		"results": []map[string]interface{}{
			{"pid": "1234", "name": "test-process"},
		},
	}
	queryJSON, err := json.Marshal(queryData)
	require.NoError(t, err)

	logger.Send(types.QueryLog, queryJSON, "test-env", "test-uuid", false)

	// Verify exactly 1 record was sent
	assert.Len(t, mockProducer.SentRecords, 1, "QueryLog should produce exactly 1 Kafka record")

	// Verify record content
	record := mockProducer.SentRecords[0]
	assert.Equal(t, "test-topic", record.Topic)
	assert.Equal(t, []byte("test-uuid"), record.Key)

	// Verify the value contains the complete query data
	var sentData map[string]interface{}
	err = json.Unmarshal(record.Value, &sentData)
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM processes", sentData["query"])
	assert.Equal(t, float64(0), sentData["status"])
}

func TestLoggerKafka_Send_ResultLog_MultipleSplit(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	logger := CreateTestableLoggerKafka(config.KafkaConfiguration{
		Topic: "test-topic",
	}, mockProducer)

	resultData := []map[string]interface{}{
		{"pid": "1234", "name": "process1", "cpu_time": "100"},
		{"pid": "5678", "name": "process2", "cpu_time": "200"},
		{"pid": "9012", "name": "process3", "cpu_time": "300"},
	}
	resultJSON, err := json.Marshal(resultData)
	require.NoError(t, err)

	logger.Send(types.ResultLog, resultJSON, "test-env", "test-uuid", false)

	// Verify 3 separate records were sent
	assert.Len(t, mockProducer.SentRecords, 3, "ResultLog array should produce 3 separate Kafka records")

	// Verify each record
	expectedData := []map[string]interface{}{
		{"pid": "1234", "name": "process1", "cpu_time": "100"},
		{"pid": "5678", "name": "process2", "cpu_time": "200"},
		{"pid": "9012", "name": "process3", "cpu_time": "300"},
	}

	for i, record := range mockProducer.SentRecords {
		// Verify topic and key are consistent
		assert.Equal(t, "test-topic", record.Topic)
		assert.Equal(t, []byte("test-uuid"), record.Key)

		// Verify each record contains the expected individual entry
		var sentData map[string]interface{}
		err = json.Unmarshal(record.Value, &sentData)
		require.NoError(t, err)

		assert.Equal(t, expectedData[i]["pid"], sentData["pid"])
		assert.Equal(t, expectedData[i]["name"], sentData["name"])
		assert.Equal(t, expectedData[i]["cpu_time"], sentData["cpu_time"])
	}
}

func TestLoggerKafka_Send_StatusLog_MultipleSplit(t *testing.T) {
	mockProducer := &MockKafkaProducer{}

	logger := CreateTestableLoggerKafka(config.KafkaConfiguration{
		Topic: "status-topic",
	}, mockProducer)

	// Test data - multiple status entries
	statusData := []map[string]interface{}{
		{"severity": "INFO", "message": "Query executed successfully"},
		{"severity": "WARNING", "message": "Config file outdated"},
	}
	statusJSON, err := json.Marshal(statusData)
	require.NoError(t, err)

	logger.Send(types.StatusLog, statusJSON, "prod-env", "host-123", true) // debug=true

	// Verify 2 separate records were sent
	assert.Len(t, mockProducer.SentRecords, 2, "StatusLog should produce 2 separate Kafka records")

	// Verify topics are all correct
	topics := mockProducer.GetSentTopics()
	for _, topic := range topics {
		assert.Equal(t, "status-topic", topic)
	}

	// Verify keys are all the same
	keys := mockProducer.GetSentKeys()
	for _, key := range keys {
		assert.Equal(t, []byte("host-123"), key)
	}

	// Verify individual message content
	values := mockProducer.GetSentValues()

	var firstMessage map[string]interface{}
	err = json.Unmarshal(values[0], &firstMessage)
	require.NoError(t, err)
	assert.Equal(t, "INFO", firstMessage["severity"])
	assert.Equal(t, "Query executed successfully", firstMessage["message"])

	var secondMessage map[string]interface{}
	err = json.Unmarshal(values[1], &secondMessage)
	require.NoError(t, err)
	assert.Equal(t, "WARNING", secondMessage["severity"])
	assert.Equal(t, "Config file outdated", secondMessage["message"])
}
