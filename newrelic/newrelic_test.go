package newrelic

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func patchEnv(key, value string) func() {
	bck := os.Getenv(key)
	deferFunc := func() {
		os.Setenv(key, bck)
	}

	os.Setenv(key, value)
	return deferFunc
}

func TestGetCredsFile(t *testing.T) {
	assert.Equal(t, path.Join("foo/", "bar"), "foo/bar")

	deferFunc := patchEnv("NEWRELIC_CREDS", "/usr/test.json")
	if file, err := getCredsFile(); assert.NoError(t, err) {
		assert.Equal(t, file, "/usr/test.json")
	}
	deferFunc()

	deferFunc = patchEnv("TERMUI_HOME", "/usr/.myTermui")
	if file, err := getCredsFile(); assert.NoError(t, err) {
		assert.Equal(t, file, "/usr/.myTermui/newrelic/creds.json")
	}
	deferFunc()

	defer patchEnv("HOME", "/usr/matt3o12/")()
	if file, err := getCredsFile(); assert.NoError(t, err) {
		assert.Equal(t, file, "/usr/matt3o12/.termui/newrelic/creds.json")
	}
}

func TestLoadAPICredentials(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	credsFile := path.Join(tempDir, "newrelic_creds.json")
	defer patchEnv("NEWRELIC_CREDS", credsFile)()
	write := func(content string) {
		err := ioutil.WriteFile(credsFile, []byte(content), os.ModePerm)
		require.NoError(t, err)
	}

	write(`{"api_key": "hello_world"}`)
	creds, err := LoadAPICredentials()
	if assert.NoError(t, err) {
		assert.Equal(t, APICredentials{"hello_world"}, creds)
	}

	write(`{"foo": "bar"}`)
	creds, err = LoadAPICredentials()
	assert.EqualError(t, err, "Error while loading newrelic credentials. "+
		"Missing API key")
	assert.Equal(t, creds, APICredentials{})

	defer patchEnv("NEWRELIC_CREDS", "/I/do/not/exist")()
	creds, err = LoadAPICredentials()
	assert.EqualError(t, err, "open /I/do/not/exist: no such file or directory")
	assert.Equal(t, creds, APICredentials{})

}
