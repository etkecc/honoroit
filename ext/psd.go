package ext

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type PSD struct {
	url      *url.URL
	login    string
	password string
}

func NewPSD(baseURL, login, password string) *PSD {
	uri, err := url.Parse(baseURL)
	if err != nil || login == "" || password == "" {
		return &PSD{}
	}
	return &PSD{url: uri, login: login, password: password}
}

// Contains checks if the host is present in the PSD
func (p *PSD) Contains(host string) (bool, error) {
	if p.url == nil {
		return false, nil
	}
	cloned := *p.url
	uri := cloned.JoinPath("/node/" + host)

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
	if resp.StatusCode == http.StatusGone { // special case: no such host
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s", resp.Status)
		return false, err
	}
	return true, nil
}
