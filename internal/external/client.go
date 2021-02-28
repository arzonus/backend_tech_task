package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	GetExternalStatus(ctx context.Context, id string) (Status, error)
}

type Config struct {
	URL            string        `envconfig:"url"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout"`
	RequestTimeout time.Duration `envconfig:"request_timeout"`
}

type client struct {
	cfg        Config
	url        *url.URL
	httpClient *http.Client
}

type GetExternalStatusResponse struct {
	Status string `json:"status"`
}

func (c client) GetExternalStatus(ctx context.Context, id string) (Status, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.RequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.url.String()+"/status/"+id,
		nil,
	)
	if err != nil {
		return StatusUnspecified, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StatusUnspecified, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StatusUnspecified, err
	}
	if resp.StatusCode != http.StatusOK {
		return StatusUnspecified, fmt.Errorf("not 200 status: %s, %s", resp.Status, string(body))
	}

	var r GetExternalStatusResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return StatusUnspecified, err
	}

	status, err := NewStatusString(r.Status)
	if err != nil {
		return StatusUnspecified, err
	}

	return status, nil
}

func NewClient(cfg Config) (Client, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 1 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 5 * time.Second
	}

	return &client{
		cfg: cfg,
		url: u,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: cfg.ConnectTimeout,
				}).DialContext,
			},
		},
	}, nil
}
