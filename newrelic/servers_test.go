package newrelic

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/matt3o12/termui-widgets/httpPlus"
	"github.com/stretchr/testify/assert"
)

func patchHTTPClient(base string) func() {
	if base[len(base)-1] != '/' {
		base += "/"
	}
	bck := NewRelicAPIBase
	deferFunc := func() {
		NewRelicAPIBase = bck
	}

	NewRelicAPIBase = base
	return deferFunc
}

func TestLoadServer(t *testing.T) {
	server := httpPlus.SetupTestServer(path.Join("..", "newrelic"))
	defer server.Close()

	defer patchHTTPClient(server.URL)()

	creds := APICredentials{}

	// Test 404
	newrelicServer, err := LoadServer(creds, 404)
	assert.Nil(t, newrelicServer)

	expectedErr := fmt.Errorf("API returned an invalid "+
		"status code: '404 Not Found' (url: %v/v2/servers/404.json)", server.URL)
	if assert.Error(t, err) {
		assert.Equal(t, err, expectedErr)
	}

	// Test normal
	newrelicServer, err = LoadServer(creds, 200)
	expected := Server{}
	expected.ID = 200
	expected.Reporting = true
	expected.Name = "server.example.com"
	expected.Host = "server.example.com"
	expected.ServerMetrics.CPU = 1
	expected.ServerMetrics.CPUStolen = .5
	expected.ServerMetrics.DiskIO = 1.5
	expected.ServerMetrics.MemoryUsed = 2502950912
	expected.ServerMetrics.MemoryTotal = 8266973184
	expected.ServerMetrics.FullestDiskFree = 1476798000000
	expected.ServerMetrics.FullestDiskPercent = 5.34

	lastReportedAt := newrelicServer.LastReportedAt
	newrelicServer.LastReportedAt = time.Time{}
	expectedTime := time.Date(2015, 4, 19, 18, 14, 5, 0, time.UTC)

	assert.Equal(t, expected, *newrelicServer)
	assert.NoError(t, err)

	if !lastReportedAt.Equal(expectedTime) {
		t.Errorf("%v != %v", expectedTime, lastReportedAt)
	}
}

func TestGetAPIEndpoint(t *testing.T) {
	assert.Equal(t, "https://api.newrelic.com/test", getAPIEndpoint("test"))
	assert.Equal(t, "https://api.newrelic.com/foobar", getAPIEndpoint("foo%v", "bar"))

	defer patchHTTPClient("http://google.com/")()
	assert.Equal(t, "http://google.com/test", getAPIEndpoint("test"))
	assert.Equal(t, "http://google.com/hello+world", getAPIEndpoint("hello+%v", "world"))
}
