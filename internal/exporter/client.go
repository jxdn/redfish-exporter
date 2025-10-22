package exporter

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type RedfishClient struct {
	baseURL *url.URL
	client  *http.Client
	auth    struct {
		username string
		password string
	}
}

type RedfishCollectorConfig struct {
	BaseURL       string
	Username      string
	Password      string
	InsecureTLS   bool
	ScrapeTimeout time.Duration
	ChassisID     string
}

func NewRedfishClient(baseURL, username, password string, insecure bool, timeout time.Duration) (*RedfishClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is empty")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	return &RedfishClient{
		baseURL: u,
		auth:    struct{ username, password string }{username: username, password: password},
		client:  &http.Client{Transport: tr, Timeout: timeout},
	}, nil
}

func (c *RedfishClient) getJSON(endpoint string, v any) error {
	u := *c.baseURL
	// Ensure we don't drop leading slash when joining
	ep := endpoint
	if !strings.HasPrefix(ep, "/") {
		ep = "/" + ep
	}
	u.Path = path.Join(c.baseURL.Path, ep)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.auth.username, c.auth.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "redfish-exporter/1.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("redfish GET %s failed: %s", endpoint, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// Minimal Redfish structures for Power metrics
type ChassisCollection struct {
	Members []struct {
		Oid string `json:"@odata.id"`
	} `json:"Members"`
}

type PowerResource struct {
	PowerControl []struct {
		Name               string   `json:"Name"`
		PowerConsumedWatts *float64 `json:"PowerConsumedWatts"`
		PowerMetrics       struct {
			AverageConsumedWatts *float64 `json:"AverageConsumedWatts"`
			MinConsumedWatts     *float64 `json:"MinConsumedWatts"`
			MaxConsumedWatts     *float64 `json:"MaxConsumedWatts"`
			IntervalInMin        *float64 `json:"IntervalInMin"`
		} `json:"PowerMetrics"`
	} `json:"PowerControl"`
}

func (c *RedfishClient) ListChassis() ([]string, error) {
	var col ChassisCollection
	if err := c.getJSON("/redfish/v1/Chassis", &col); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(col.Members))
	for _, m := range col.Members {
		ids = append(ids, m.Oid)
	}
	return ids, nil
}

func (c *RedfishClient) GetChassisPower(odataID string) (*PowerResource, error) {
	// odataID is like /redfish/v1/Chassis/<id>
	endpoint := path.Join(odataID, "Power")
	var p PowerResource
	if err := c.getJSON(endpoint, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
