package redmine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
)

const (
	limitDefault = 100
)

type StatusCode int64

type Settings struct {
	Endpoint string
	APIKey   string
}

// Context struct used for store settings to communicate with Redmine API
type Context struct {
	endpoint string
	apiKey   string
}

// IDName used as embedded struct for other structs within package
type IDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type errorsResult struct {
	Errors []string `json:"errors"`
}

func Init(s Settings) *Context {
	return &Context{
		endpoint: s.Endpoint,
		apiKey:   s.APIKey,
	}
}

// SetAPIKey is used to set Redmine API key
func (r *Context) SetAPIKey(apiKey string) {
	r.apiKey = apiKey
}

// SetEndpoint is used to set Redmine endpoint
func (r *Context) SetEndpoint(endpoint string) {
	r.endpoint = endpoint
}

func (r *Context) Get(out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {

	var er errorsResult

	u := r.endpoint + uri.String()

	// Create request
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Add("X-Redmine-API-Key", r.apiKey)

	// Make request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	dJ := json.NewDecoder(res.Body)

	if StatusCode(res.StatusCode) != statusExpected {
		if err := dJ.Decode(&er); err != nil {
			er.Errors = append(er.Errors, fmt.Sprintf("json decode error: %v", err))
		}
		er.Errors = append(er.Errors, fmt.Sprintf("unexpected status code has been returned (expected: %d, returned: %d, url: %s, method: %s)", statusExpected, res.StatusCode, u, http.MethodGet))

		err = errors.New(strings.Join(er.Errors, "\n"))
	} else {
		if out != nil {

			rawConf := make(map[string]interface{})

			if err := dJ.Decode(&rawConf); err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("json decode error: %v", err)
			}

			dM, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           out,
				TagName:          "json",
			})
			if err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("mapstructure create decoder error: %v", err)
			}

			if err := dM.Decode(rawConf); err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("mapstructure decode error: %v", err)
			}
		}
	}

	return StatusCode(res.StatusCode), err
}

func (r *Context) Post(in interface{}, out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {
	return r.alter(http.MethodPost, in, out, uri, statusExpected)
}

func (r *Context) Put(in interface{}, out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {
	return r.alter(http.MethodPut, in, out, uri, statusExpected)
}

func (r *Context) Del(in interface{}, out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {
	return r.alter(http.MethodDelete, in, out, uri, statusExpected)
}

func (r *Context) alter(method string, in interface{}, out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {

	var er errorsResult

	u := r.endpoint + uri.String()

	s, err := json.Marshal(in)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(method, u, strings.NewReader(string(s)))
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Redmine-API-Key", r.apiKey)

	// Make request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	dJ := json.NewDecoder(res.Body)

	if StatusCode(res.StatusCode) != statusExpected {

		if err := dJ.Decode(&er); err != nil {
			er.Errors = append(er.Errors, fmt.Sprintf("json decode error: %v", err))
		}
		er.Errors = append(er.Errors, fmt.Sprintf("unexpected status code has been returned (expected: %d, returned: %d, url: %s, method: %s)", statusExpected, res.StatusCode, u, method))

		err = errors.New(strings.Join(er.Errors, "\n"))
	} else {
		if out != nil {

			rawConf := make(map[string]interface{})

			if err := dJ.Decode(&rawConf); err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("json decode error: %v", err)
			}

			dM, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				WeaklyTypedInput: true,
				Result:           out,
				TagName:          "json",
			})
			if err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("mapstructure create decoder error: %v", err)
			}

			if err := dM.Decode(rawConf); err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("mapstructure decode error: %v", err)
			}
		}
	}

	return StatusCode(res.StatusCode), err
}

func (r *Context) uploadFile(f io.Reader, out interface{}, uri url.URL, statusExpected StatusCode) (StatusCode, error) {

	var er errorsResult

	u := r.endpoint + uri.String()

	// Create request
	req, err := http.NewRequest(http.MethodPost, u, f)
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Add("X-Redmine-API-Key", r.apiKey)

	// Make request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	dJ := json.NewDecoder(res.Body)

	if StatusCode(res.StatusCode) != statusExpected {
		if err := dJ.Decode(&er); err != nil {
			er.Errors = append(er.Errors, fmt.Sprintf("json decode error: %v", err))
		}
		er.Errors = append(er.Errors, fmt.Sprintf("unexpected status code has been returned (expected: %d, returned: %d, url: %s, method: %s)", statusExpected, res.StatusCode, u, http.MethodPost))

		err = errors.New(strings.Join(er.Errors, "\n"))
	} else {
		if out != nil {
			if err := dJ.Decode(&out); err != nil {
				return StatusCode(res.StatusCode), fmt.Errorf("json decode error: %v", err)
			}
		}
	}

	return StatusCode(res.StatusCode), err
}

func (r *Context) downloadFile(url string, statusExpected StatusCode) (io.ReadCloser, StatusCode, error) {

	var er errorsResult

	// Create request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	// Set headers
	req.Header.Add("X-Redmine-API-Key", r.apiKey)

	// Make request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if StatusCode(res.StatusCode) != statusExpected {
		if err := json.NewDecoder(res.Body).Decode(&er); err != nil {
			er.Errors = append(er.Errors, fmt.Sprintf("json decode error: %v", err))
		}
		er.Errors = append(er.Errors, fmt.Sprintf("unexpected status code has been returned (expected: %d, returned: %d, url: %s, method: %s)", statusExpected, res.StatusCode, url, http.MethodPost))

		res.Body.Close()

		return nil, 0, errors.New(strings.Join(er.Errors, "\n"))
	}

	return res.Body, StatusCode(res.StatusCode), err
}
