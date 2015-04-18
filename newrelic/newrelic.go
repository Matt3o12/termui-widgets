package newrelic

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

// APICredentials for the relic API.
type APICredentials struct {
	APIKey string `json:"api_key"`
}

// The default path of the newrelic credentionals.
var DefaultAPICredentialsPath = path.Join(
	"~", ".termui", "newrelic", "creds.json")

func getCredsFile() (string, error) {
	getter := func() string {
		if credPath := os.Getenv("NEWRELIC_CREDS"); credPath != "" {
			return credPath
		}

		if termuiHome := os.Getenv("TERMUI_HOME"); termuiHome != "" {
			return path.Join(termuiHome, "newrelic", "creds.json")
		}

		return DefaultAPICredentialsPath
	}

	path, err := homedir.Expand(getter())
	if err != nil {
		return "", err
	}

	return path, nil
}

// LoadAPICredentials loads the API credentials from env variables or
// the default path. It may return IO errors.
func LoadAPICredentials() (APICredentials, error) {
	filename, err := getCredsFile()
	if err != nil {
		return APICredentials{}, err
	}

	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0006)
	if err != nil {
		return APICredentials{}, err
	}

	var creds APICredentials
	if err := json.NewDecoder(file).Decode(&creds); err != nil {
		return creds, err
	}

	if creds.APIKey == "" {
		return creds, errors.New("Error while loading newrelic " +
			"credentials. Missing API key")
	}

	return creds, nil
}
