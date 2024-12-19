package repository

import (
	"fmt"
	"io"
	"net/http"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

func (p *Processor) download(updatesCh chan<- entity.Update) (io.ReadCloser, error) {
	updatesCh <- entity.Update{
		Description: fmt.Sprintf("downloading: %s", p.config.Url),
	}

	resp, err := http.Get(p.config.Url)
	if err != nil {
		updatesCh <- entity.Update{
			Description: "Request error getting tarball",
			Err:         err,
		}
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		updatesCh <- entity.Update{
			Description: fmt.Sprintf("Request error getting tarbal with status code: %d", resp.StatusCode),
			Err:         fmt.Errorf("Status code: %d", resp.StatusCode),
		}
		return nil, fmt.Errorf("Status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
