package httpPlus

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tmpDir(t *testing.T) string {
	testDir, err := ioutil.TempDir("", "httpPlus_test_")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return testDir
}

func readResp(t *testing.T, resp *http.Response) string {
	result, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	return string(result)
}

// Test the "TestServer"
func TestServer(t *testing.T) {
	// Create tmpdir and remove on exit.
	testDir := tmpDir(t)
	defer os.RemoveAll(testDir)

	// Setup server and close on exit.
	server := SetupTestServer(testDir)
	defer server.Close()

	assert.NotNil(t, server)

	// Create folder hierarchy
	resourcePath := path.Join(testDir, "resources", "api", "foo", "bar")
	if err := os.MkdirAll(resourcePath, 0777); !assert.NoError(t, err) {
		t.FailNow()
	}

	// Create file and add content: "world :)"
	file := path.Join(resourcePath, "hello.txt")
	content := []byte("world :)")
	if err := ioutil.WriteFile(file, content, 0777); !assert.NoError(t, err) {
		t.FailNow()
	}

	// Load resource file.
	resp, err := http.Get(server.URL + "/foo/bar/hello.txt")
	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	assert.NoError(t, err)
	assert.Equal(t, "world :)", readResp(t, resp))
}

func TestServerNotFound(t *testing.T) {
	testDir := tmpDir(t)
	defer os.RemoveAll(testDir)

	server := SetupTestServer(testDir)

	resp, err := http.Get(server.URL + "/index.html")
	defer resp.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	expeted := "open %v/resources/api/index.html: no such file or directory\n"
	assert.Equal(t, fmt.Sprintf(expeted, testDir), readResp(t, resp))
}
