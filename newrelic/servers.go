package newrelic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lafikl/fluent"
)

// HTTPClient for fetching resources from the new server.
// Used across all connections because of HttpKeepalive.
var HTTPClient http.Client

// ServerWidget displays metrics from NewRelic's Monetoring service.
// The LastServer may be nil if no server is loaded, yet.
type ServerWidget struct {
	ServerID        int
	Error           error
	LastServer      *Server
	RefreshInterval time.Duration
	NewRelicCreds   APICredentials
}

// The base for all API requests. Needs to be a var so that it can be patched
// during unit tests.
var NewRelicAPIBase = "https://api.newrelic.com/"

// All API points for new relic
const (
	NewRelicServerAPIPoint = "v2/servers/%v.json"
)

// NewServerWidget creates a new ServerWidget
func NewServerWidget(creds APICredentials, id int) ServerWidget {
	return ServerWidget{
		ServerID:        id,
		NewRelicCreds:   creds,
		RefreshInterval: 15 * time.Second,
	}
}

// UpdateServer updates the server's metrics automatically. This is
// non-blocking
func (w ServerWidget) UpdateServer(callback func()) {
	go func() {
		for {
			newServer, err := LoadServer(w.NewRelicCreds, w.ServerID)
			w.Error = err
			w.LastServer = newServer

			callback()
			time.Sleep(w.RefreshInterval)
		}
	}()
}

// A Server represents a simple NewRelic server from the server monetoring
// module.
type Server struct {
	ID             int           `json:"id"`
	LastReportedAt time.Time     `json:"last_reported_at"`
	Reporting      bool          `json:"reporting"`
	Name           string        `json:"name"`
	Host           string        `json:"host"`
	ServerMetrics  ServerMetrics `json:"summary"`
}

// The ServerMetrics of a server.
type ServerMetrics struct {
	CPU                float64 `json:"cpu"`
	CPUStolen          float64 `json:"cpu_stolen"`
	DiskIO             float64 `json:"disk_io"`
	FullestDiskPercent float64 `json:"fullest_disk"`
	FullestDiskFree    int     `json:"fullest_disk_free"`
	MemoryUsed         int     `json:"memory_used"`
	MemoryTotal        int     `json:"memory_total"`
}

func prepareRequest(creds APICredentials) *fluent.Request {
	request := fluent.New()
	return request.Retry(5).SetHeader("X-Api-Key", creds.APIKey)
}

func getAPIEndpoint(url string, args ...interface{}) string {
	return fmt.Sprintf(NewRelicAPIBase+url, args...)
}

// LoadServer fetches and returns the Server and its metrics from the NewRelic
// API.
func LoadServer(creds APICredentials, id int) (*Server, error) {
	url := getAPIEndpoint(NewRelicServerAPIPoint, id)
	request := prepareRequest(creds).Get(url)
	response, err := request.Send()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		msg := "API returned an invalid status code: '%v' (url: %v)"
		return nil, fmt.Errorf(msg, response.Status, url)
	}

	type jsonStruct struct {
		Server *Server `json:"server"`
	}

	decoder := json.NewDecoder(response.Body)
	var jsonResponse jsonStruct
	if err := decoder.Decode(&jsonResponse); err != nil {
		fmt.Println("!!", err)
		return nil, err
	}

	if jsonResponse.Server == nil {
		return nil, fmt.Errorf("API returned invalid response (url: %v)", url)
	}

	return jsonResponse.Server, nil
}
