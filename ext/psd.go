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

func (p *PSD) GetMXIDs(host string) ([]string, error) {
	if p.url == nil {
		return nil, nil
	}
	cloned := *p.url
	uri := cloned.JoinPath("/mxids/" + host)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(p.login, p.password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusGone { // special case: no such host
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s", resp.Status)
		return nil, err
	}
	datab, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var psd []*PSDTarget
	err = json.Unmarshal(datab, &psd)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(psd))
	for _, p := range psd {
		targets = append(targets, p.Targets...)
	}
	return targets, nil
}
