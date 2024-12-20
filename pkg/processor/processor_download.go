package processor

import (
	"fmt"
	"io"
	"net/http"
)

func (p *Processor) download() (io.ReadCloser, error) {
	p.update(fmt.Sprintf("downloading: %s", p.config.Url), nil)

	resp, err := http.Get(p.config.Url)
	if err != nil {
		p.update("error getting tarball", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		p.update(
			fmt.Sprintf("error getting tarball, status code: %d", resp.StatusCode),
			fmt.Errorf("Status code: %d", resp.StatusCode),
		)
		return nil, fmt.Errorf("Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
