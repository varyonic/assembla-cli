package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// AssemblaClient handles HTTP communication with the Assembla API.
type AssemblaClient struct {
	apiURL     string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// NewAssemblaClient creates a new API client.
func NewAssemblaClient(apiKey, apiSecret, apiURL string) *AssemblaClient {
	return &AssemblaClient{
		apiURL:    strings.TrimRight(apiURL, "/"),
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			CheckRedirect: stripCredentialsOnHostChange,
		},
	}
}

// stripCredentialsOnHostChange keeps the API credential headers from being
// replayed to a different host across a redirect. Go forwards custom headers on
// redirect and only strips the ones it recognises as sensitive (Authorization,
// Cookie, ...), which would otherwise let a validated host hand the credentials
// to an arbitrary third party.
func stripCredentialsOnHostChange(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}
	if len(via) > 0 && !strings.EqualFold(req.URL.Host, via[0].URL.Host) {
		req.Header.Del("X-Api-Key")
		req.Header.Del("X-Api-Secret")
	}
	return nil
}

func (c *AssemblaClient) request(method, path string, params map[string]string, body interface{}) (*http.Response, error) {
	reqURL := fmt.Sprintf("%s/v1%s", c.apiURL, path)

	// Add query parameters
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		reqURL += "?" + q.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("X-Api-Secret", c.apiSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		handleError(resp)
	}

	return resp, nil
}

// Get performs a GET request and returns the parsed JSON response.
func (c *AssemblaClient) Get(path string, params map[string]string) (interface{}, error) {
	resp, err := c.request("GET", path, params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Post performs a POST request and returns the parsed JSON response.
func (c *AssemblaClient) Post(path string, body interface{}) (interface{}, error) {
	resp, err := c.request("POST", path, nil, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Put performs a PUT request and returns the parsed JSON response.
func (c *AssemblaClient) Put(path string, body interface{}) (interface{}, error) {
	resp, err := c.request("PUT", path, nil, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Delete performs a DELETE request.
func (c *AssemblaClient) Delete(path string) error {
	resp, err := c.request("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetAll performs paginated GET requests and returns all results.
func (c *AssemblaClient) GetAll(path string, perPage int, params map[string]string) ([]interface{}, error) {
	var results []interface{}
	page := 1

	for {
		pageParams := make(map[string]string)
		for k, v := range params {
			pageParams[k] = v
		}
		pageParams["page"] = fmt.Sprintf("%d", page)
		pageParams["per_page"] = fmt.Sprintf("%d", perPage)

		data, err := c.Get(path, pageParams)
		if err != nil {
			return nil, err
		}
		if data == nil {
			break
		}

		arr, ok := data.([]interface{})
		if !ok {
			break
		}
		results = append(results, arr...)
		if len(arr) < perPage {
			break
		}
		page++
	}

	return results, nil
}

func handleError(resp *http.Response) {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var msg string
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if e, ok := parsed["error"]; ok {
			msg = fmt.Sprintf("%v", e)
		} else if e, ok := parsed["errors"]; ok {
			msg = fmt.Sprintf("%v", e)
		} else {
			msg = string(body)
		}
	} else {
		msg = string(body)
	}

	fmt.Fprintf(os.Stderr, "Error %d: %s\n", resp.StatusCode, msg)
	os.Exit(1)
}
