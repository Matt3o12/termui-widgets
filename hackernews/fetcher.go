package hackernews

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// EntryAPIPoint points the the entry api point of the hackernews API.
	EntryAPIPoint = "%s/v0/item/%v.json"

	// TopStoriesAPIPoint points to the top stories on HN.
	TopStoriesAPIPoint = "%v/v0/topstories.json"

	// NewStoriesAPIPoint points to the most recent stories on HN.
	NewStoriesAPIPoint = "%v/v0/newstories.json"

	apiStatusError = "API endpoint returned invalid status: %v"
	maxRetries     = 5
)

// APIBaseURL is the url of the hacker news website.
// This is not a constant so that it can be patched by the tests.
var APIBaseURL = "https://hacker-news.firebaseio.com"

var client = http.Client{}

// JSONDecodeErr return when the json file couldn't be parsed.
var ErrJSONDecode = errors.New("Error while decoding the JSON response " +
	"from the server (invalid type/unexpected nil type).")

// GetEntryPoint returns the full URL for the given entry point.
func GetEntryPoint(point string, args ...interface{}) string {
	args = append([]interface{}{APIBaseURL}, args...)
	return fmt.Sprintf(point, args...)
}

// Requests a resource from the URL and retries up to n-times.
// If retryCount is not grather then 0, this function will panic.
func doRequest(retryCount int, url string) (*http.Response, error) {
	if retryCount <= 0 {
		panic("retryCount must be grather then 0")
	}

	var response *http.Response
	var err error

	for i := 0; i < retryCount; i++ {
		response, err = client.Get(url)

		if err == nil && response != nil {
			return response, nil
		}
	}

	if response == nil && err == nil {
		err = errors.New("http.Get returned a nil response")
	}

	return response, err
}

// LoadEntry loads an `Entry` from the server.
func LoadEntry(id int) (Entry, error) {
	url := GetEntryPoint(EntryAPIPoint, id)
	resp, err := doRequest(maxRetries, url)
	if err != nil {
		return Entry{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Entry{}, fmt.Errorf(apiStatusError, resp.Status)
	}

	reader := json.NewDecoder(resp.Body)
	var values interface{}
	if err := reader.Decode(&values); err != nil {
		return Entry{}, err
	}

	mappedValues, ok := values.(map[string]interface{})
	if !ok {
		return Entry{}, ErrJSONDecode
	}

	title, ok1 := mappedValues["title"].(string)
	timestamp, ok2 := mappedValues["time"].(float64)
	if !ok1 || !ok2 {
		return Entry{}, ErrJSONDecode
	}

	entryTime := time.Unix(int64(timestamp), 0)
	return Entry{title, entryTime, id}, nil
}

func loadIDs(url string) ([]int, error) {
	resp, err := doRequest(maxRetries, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(apiStatusError, resp.Status)
	}

	reader := json.NewDecoder(resp.Body)
	var ids []int
	if err := reader.Decode(&ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// LoadTopIDs fetches the top IDs fromt the HN database.
func LoadTopIDs() ([]int, error) {
	return loadIDs(GetEntryPoint(TopStoriesAPIPoint))
}

// LoadMostRecentIDs loads the most recent IDs from the HN database.
func LoadMostRecentIDs() ([]int, error) {
	return loadIDs(GetEntryPoint(NewStoriesAPIPoint))
}
