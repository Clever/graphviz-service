package kayvee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"testing"

	"github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Tests struct {
	Version        string     `json:"version"`
	FormatTests    []TestSpec `json:"format"`
	FormatLogTests []TestSpec `json:"formatLog"`
}

type TestSpec struct {
	Title  string                 `json:"title"`
	Input  map[string]interface{} `json:"input"`
	Output string                 `json:"output"`
}

type keyVal map[string]interface{}

type sentryMock struct {
	Packets []*raven.Packet
}

func (s *sentryMock) Capture(packet *raven.Packet, captureTags map[string]string) (eventID string, ch chan error) {
	s.Packets = append(s.Packets, packet)
	return "12345", nil
}

// takes two strings (which are assumed to be JSON)
func compareJSONStrings(t *testing.T, expected string, actual string) {
	actualJSON := map[string]interface{}{}
	expectedJSON := map[string]interface{}{}
	err := json.Unmarshal([]byte(actual), &actualJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `actual`:", actual))
	}
	err = json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		panic(fmt.Sprint("failed to json unmarshal `expected`:", expected))
	}

	assert.Equal(t, expectedJSON, actualJSON)
}

func Test_KayveeSpecs(t *testing.T) {
	file, err := ioutil.ReadFile("tests.json")
	assert.NoError(t, err, "failed to open test specs (tests.json)")
	var tests Tests
	json.Unmarshal(file, &tests)
	t.Logf("spec (tests.json) version: %s\n", string(tests.Version))

	for _, spec := range tests.FormatTests {
		expected := spec.Output
		actual := Format(spec.Input["data"].(map[string]interface{}))

		compareJSONStrings(t, expected, actual)
	}

	for _, spec := range tests.FormatLogTests {
		expected := spec.Output

		// Ensure correct type is passed to FormatLog, even if value undefined in tests.json
		source, _ := spec.Input["source"].(string)
		level, _ := spec.Input["level"].(string)
		title, _ := spec.Input["title"].(string)
		data, _ := spec.Input["data"].(map[string]interface{})
		loglevel := LogLevel(level)
		actual := FormatLog(source, loglevel, title, data)

		compareJSONStrings(t, expected, actual)
	}

}

func assertLogFormatAndCompareContent(t *testing.T, logline, expected string) {
	rx := regexp.MustCompile(`kayvee_test\.go.*({.*})`)
	require.Regexp(t, rx, logline)
	actual := rx.FindStringSubmatch(logline)[1]
	compareJSONStrings(t, expected, actual)
}

func TestLogInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSentryLogger("testkayvee", log.New(buf, "", defaultFlags), nil)
	logger.Info("testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Info, "testloginfo", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSentryLogger("testkayvee", log.New(buf, "", defaultFlags), nil)
	logger.Warning("testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"})
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Warning, "testlogwarning", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogErrorNoSentryClient(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSentryLogger("testkayvee", log.New(buf, "", defaultFlags), nil)
	logger.Error("testlogerrornosentryclient", map[string]interface{}{"key1": "val1", "key2": "val2"}, fmt.Errorf("testerror"))
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Error, "testlogerrornosentryclient", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogErrorNoSentryClientNoError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewSentryLogger("testkayvee", log.New(buf, "", defaultFlags), nil)
	logger.Error("testlogerrornosentryclientnoerror", map[string]interface{}{"key1": "val1", "key2": "val2"}, nil)
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Error, "testlogerrornosentryclientnoerror", map[string]interface{}{"key1": "val1", "key2": "val2"}))
}

func TestLogErrorSentryClientNoError(t *testing.T) {
	buf := &bytes.Buffer{}
	sentry := &sentryMock{}
	logger := &SentryLogger{source: "testkayvee", logger: log.New(buf, "", defaultFlags), sentryClient: sentry}
	logger.Error("testlogerrorsentryclientnoerror", map[string]interface{}{"key1": "val1", "key2": "val2"}, nil)
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Error, "testlogerrorsentryclientnoerror", map[string]interface{}{"key1": "val1", "key2": "val2"}))
	assert.Empty(t, sentry.Packets)
}

func TestLogErrorSentryClient(t *testing.T) {
	buf := &bytes.Buffer{}
	sentry := &sentryMock{}
	logger := &SentryLogger{source: "testkayvee", logger: log.New(buf, "", defaultFlags), sentryClient: sentry}
	logger.Error("testlogerrorsentryclient", map[string]interface{}{"key1": "val1", "key2": "val2"}, fmt.Errorf("testerror"))
	assertLogFormatAndCompareContent(t, string(buf.Bytes()), FormatLog(
		"testkayvee", Error, "testlogerrorsentryclient", map[string]interface{}{"key1": "val1", "key2": "val2", "sentry_event_id": "12345"}))
	require.Equal(t, 1, len(sentry.Packets))
	require.Equal(t, 1, len(sentry.Packets[0].Interfaces))
	exception, ok := sentry.Packets[0].Interfaces[0].(*raven.Exception)
	require.True(t, ok)
	assert.Equal(t, "testerror", exception.Value)
	require.NotEmpty(t, exception.Stacktrace.Frames)
	lastFrame := exception.Stacktrace.Frames[len(exception.Stacktrace.Frames)-1]
	assert.Equal(t, "github.com/Clever/kayvee-go/kayvee_test.go", lastFrame.Filename)
	assert.Equal(t, "TestLogErrorSentryClient", lastFrame.Function)
	assert.Equal(t, 3, len(lastFrame.PreContext))
	assert.Equal(t, 3, len(lastFrame.PostContext))
}
