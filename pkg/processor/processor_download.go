package processor

import (
	"fmt"
	"io"
	"net/http"
)

func (p *Processor) download() (io.ReadCloser, error) {
	p.update(fmt.Sprintf("downloading: %s", p.config.Url))

	resp, err := http.Get(p.config.Url)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
