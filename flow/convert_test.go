package flow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/BaritoLog/go-boilerplate/testkit"
	"github.com/BaritoLog/go-boilerplate/timekit"
	"github.com/Shopify/sarama"
)

func TestConvertBytesToTimber_GenerateTimestamp(t *testing.T) {
	timekit.FreezeUTC("2018-06-06T12:12:12Z")
	defer timekit.Unfreeze()

	timber, err := ConvertBytesToTimber(sampleRawTimber())
	FatalIfError(t, err)
	FatalIf(t, timber.Timestamp() != "2018-06-06T12:12:12Z", "wrong timber.Timestamp()")
}

func TestConvertBytesToTimber_JsonParseError(t *testing.T) {
	_, err := ConvertBytesToTimber([]byte(`invalid_json`))
	FatalIfWrongError(t, err, string(JsonParseError))
}
func TestConvertBytesToTimberCollection_JsonParseError(t *testing.T) {
	_, err := ConvertBytesToTimberCollection([]byte(`invalid_json`))
	FatalIfWrongError(t, err, string(JsonParseError))
}

func TestConvertBytesToTimber_MissingContext(t *testing.T) {
	_, err := ConvertBytesToTimber([]byte(`{"hello":"world"}`))
	FatalIfWrongError(t, err, string(MissingContextError))
}

func TestConvertBytesToTimber_InvalidContext(t *testing.T) {
	_, err := ConvertBytesToTimber([]byte(`{"_ctx":{}}`))
	FatalIfWrongError(t, err, "Invalid Context Error: kafka_topic is missing")
}

func TestConvertRequestToTimber(t *testing.T) {

	req, err := http.NewRequest("POST", "/", bytes.NewReader(sampleRawTimber()))
	FatalIfError(t, err)

	timber, err := ConvertRequestToTimber(req)
	FatalIfError(t, err)

	FatalIf(t, timber["message"] != "some-message", "Wrong timber.message")
}

func TestConvertBatchRequestToTimberCollection(t *testing.T) {

	req, err := http.NewRequest("POST", "/produce_batch", bytes.NewReader(sampleRawTimberCollection()))
	FatalIfError(t, err)

	timberCollection, err := ConvertBatchRequestToTimberCollection(req)
	FatalIfError(t, err)

	FatalIf(t, timberCollection.Items[0]["message"] != "some-message-1", "Wrong timber.message")
}

func TestNewTimberFromKafka(t *testing.T) {

	message := &sarama.ConsumerMessage{
		Topic: "some-topic",
		Value: sampleRawTimber(),
	}

	timber, err := ConvertKafkaMessageToTimber(message)
	FatalIfError(t, err)
	FatalIf(t, timber["message"] != "some-message", "Wrong timber[message]")
}

func TestConvertToKafkaMessage(t *testing.T) {
	topic := "some-topic"

	timber := Timber{}
	timber.SetTimestamp("2018-03-10T23:00:00Z")

	message := ConvertTimberToKafkaMessage(timber, topic)
	FatalIf(t, message.Topic != topic, "%s != %s", message.Topic, topic)

	get, _ := message.Value.Encode()
	expected, _ := json.Marshal(timber)
	FatalIf(t, string(get) != string(expected), "Wrong message value")
}

func TestConvertTimberToElasticDocument(t *testing.T) {
	timber := NewTimber()
	timber["hello"] = "world"
	timber.SetContext(&TimberContext{
		KafkaTopic:     "some-topic",
		ESIndexPrefix:  "some-prefix",
		ESDocumentType: "some-type",
	})

	document := ConvertTimberToElasticDocument(timber)
	FatalIf(t, len(document) != 1, "wrong document size")
	FatalIf(t, timber["hello"] != "world", "wrong document.hello")
}

func sampleRawTimber() []byte {
	return []byte(`{
		"location": "some-location",
		"message":"some-message",
		"_ctx": {
			"kafka_topic": "some_topic",
			"kafka_partition": 3,
			"kafka_replication_factor": 1,
			"es_index_prefix": "some-type",
			"es_document_type": "some-type",
			"app_max_tps": 10,
			"app_secret": "some-secret-1234"
		}
	}`)

}

func sampleRawTimberCollection() []byte {
	return []byte(`{
		"items": [
			{
				"location": "some-location-1",
				"message":"some-message-1"
			},
			{
				"location": "some-location-2",
				"message":"some-message-2"
			}
		],
		"_ctx": {
			"kafka_topic": "some_topic",
			"kafka_partition": 3,
			"kafka_replication_factor": 1,
			"es_index_prefix": "some-type",
			"es_document_type": "some-type",
			"app_max_tps": 10,
			"app_secret": "some-secret-1234"
		}
	}`)

}
