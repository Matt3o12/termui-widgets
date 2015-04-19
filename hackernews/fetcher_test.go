package hackernews

import (
	"fmt"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/matt3o12/termui-widgets/httpPlus"
	"github.com/stretchr/testify/assert"
)

var HTTPServer *httptest.Server

var TestTimestamp = time.Unix(1257894000, 0)

func PatchClient() func() {
	if HTTPServer == nil {
		HTTPServer = httpPlus.SetupTestServer(path.Join("..", "hackernews"))
	}

	return PatchAPIBaseURL(HTTPServer.URL)
}

func PatchAPIBaseURL(newURL string) func() {
	backup := APIBaseURL
	APIBaseURL = newURL

	return func() {
		APIBaseURL = backup
	}
}

func IntegrationTest(t *testing.T) {
	if testing.Short() {
		m := "This is a integration tests but short mode is turned on – skipping."
		t.Skip(m)
		t.SkipNow()
	}
}

func AssertEntry(t *testing.T, id int, name string, timestamp time.Time) {
	entry, err := LoadEntry(id)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, entry.Title, name)
	assert.Equal(t, entry.ID, id)
	msg := fmt.Sprintf("%v != %v", entry.Time, timestamp)
	assert.True(t, entry.Time.Equal(timestamp), msg)
}

func AssertEntryError(t *testing.T, id int, expectedErr error) {
	defer PatchClient()()
	entry, err := LoadEntry(id)
	assert.Equal(t, Entry{}, entry)
	assert.Equal(t, expectedErr, err)
}

func TestLoadEntry(t *testing.T) {
	defer PatchClient()()
	AssertEntry(t, 123, "Hello world. This is a mocked entry", TestTimestamp)
}

func TestLoadEntry_MissingTitle(t *testing.T) {
	AssertEntryError(t, 503, ErrJSONDecode)
}

func TestLoadEntry_InvalidResponse(t *testing.T) {
	AssertEntryError(t, 1503, ErrJSONDecode)
}

func TestLoadEntry_InvalidID(t *testing.T) {
	defer PatchClient()()
	entry, err := LoadEntry(503)
	assert.Error(t, err)
	assert.Equal(t, Entry{}, entry)
}

func TestLoadEntry_Integartion(t *testing.T) {
	IntegrationTest(t)
	eTime := time.Unix(1172394646, 0)
	AssertEntry(t, 1000, "How Important is the .com TLD?", eTime)
}

func TestLoadTopIDs(t *testing.T) {
	defer PatchClient()()
	if ids, err := LoadTopIDs(); assert.NoError(t, err) {
		assert.Equal(t, []int{123, 456, 789, 123}, ids)
	}
}

func TestLoadTopIDs_Integartion(t *testing.T) {
	IntegrationTest(t)
	if ids, err := LoadTopIDs(); assert.NoError(t, err) {
		assert.Equal(t, 500, len(ids))
	}
}

func TestLoadMostRecentIDs(t *testing.T) {
	defer PatchClient()()
	if ids, err := LoadMostRecentIDs(); assert.NoError(t, err) {
		assert.Equal(t, []int{105, 103, 99, 5, 1}, ids)
	}
}

func TestLoadMostRecentIDs_Integration(t *testing.T) {
	IntegrationTest(t)
	if ids, err := LoadMostRecentIDs(); assert.NoError(t, err) {
		assert.Equal(t, 500, len(ids))
	}
}

func TestGetEntryPoint(t *testing.T) {
	defer PatchAPIBaseURL("http://google.com")()

	uri := "%v/hello/%v/test?foo=%v"
	expected := "http://google.com/hello/world/test?foo=bar"

	assert.Equal(t, expected, GetEntryPoint(uri, "world", "bar"))
}
