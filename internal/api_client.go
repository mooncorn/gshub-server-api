package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mooncorn/gshub-server-api/config"
)

type ServiceConfiguration struct {
	Name     string   `json:"name"`
	NameLong string   `json:"nameLong"`
	Image    string   `json:"image"`
	MinMem   int      `json:"minMem"`
	RecMem   int      `json:"recMem"`
	Env      []Env    `json:"env"`
	Ports    []Port   `json:"ports"`
	Volumes  []Volume `json:"volumes"`
}

type Env struct {
	Name        string  `json:"name"`
	Key         string  `json:"key"`
	Required    bool    `json:"required"`
	Description string  `json:"description"`
	Default     string  `json:"default"`
	Values      []Value `json:"values"`
}

type Value struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Port struct {
	Host      int64  `json:"host"`
	Container int64  `json:"container"`
	Protocol  string `json:"protocol"`
}

type Volume struct {
	Host        string `json:"host"`
	Destination string `json:"destination"`
}

type Service struct {
	ID     uint   `json:"id"`
	NameID string `json:"nameId"`
}

type StartupPayload struct {
	InstanceMemory int                             `json:"instanceMemory"`
	OwnerID        uint                            `json:"ownerId"`
	Cycles         uint                            `json:"cycles"`
	ServiceConfigs map[string]ServiceConfiguration `json:"serviceConfigs"`
	Services       []Service                       `json:"services"`
}

type ApiClient struct {
	baseUrl    string
	instanceId string
	httpClient *http.Client
}

func NewClient() *ApiClient {
	return &ApiClient{
		baseUrl:    os.Getenv("INTERNAL_API_URL"),
		instanceId: config.Env.InstanceId,
		httpClient: &http.Client{},
	}
}

// Gets initialization data for this instance and posts failed burned cycles
func (c *ApiClient) PostStartup(failedBurnedCyclesTotalAmount uint) (*StartupPayload, error) {
	url := fmt.Sprintf("%s/startup/%s", c.baseUrl, c.instanceId)
	payload := map[string]interface{}{"failedBurnedCyclesTotalAmount": failedBurnedCyclesTotalAmount, "publicIp": ""}
	response, err := c.sendRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	var result StartupPayload
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return &result, nil
}

// Posts the number of burned cycles
func (c *ApiClient) PostShutdown(burnedCycles uint) error {
	url := fmt.Sprintf("%s/shutdown/%s", c.baseUrl, c.instanceId)
	payload := map[string]uint{"burnedCyclesAmount": burnedCycles}
	if _, err := c.sendRequest("POST", url, payload); err != nil {
		return err
	}

	return nil
}

// sendRequest is a helper method to send HTTP requests
func (c *ApiClient) sendRequest(method, url string, payload interface{}) ([]byte, error) {
	var reqBody io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}
