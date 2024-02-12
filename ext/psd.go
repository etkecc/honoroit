package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type PSD struct {
	url      *url.URL
	login    string
	password string
}

type PSDTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func NewPSD(baseURL, login, password string) *PSD {
	uri, err := url.Parse(baseURL)
	if err != nil || login == "" || password == "" {
		return &PSD{}
	}
	return &PSD{url: uri, login: login, password: password}
}

func (p *PSD) Contains(identifier string) (bool, error) {
	if p.url == nil {
		return false, nil
	}
	cloned := *p.url
	uri := cloned.JoinPath("/node/" + identifier)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(p.login, p.password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s", resp.Status)
		return false, err
	}
	datab, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var psd []*PSDTarget
	err = json.Unmarshal(datab, &psd)
	if err != nil {
		return false, err
	}

	return len(psd) > 0, nil
}
