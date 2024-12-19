package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

type Processor struct {
	config *entity.Config
	data   map[string]interface{}
}

func NewProcessor(config *entity.Config) *Processor {
	return &Processor{
		config: config,
		data:   make(map[string]interface{}),
	}
}

func (p *Processor) Process(updatesCh chan<- entity.Update, delay time.Duration) error {
	defer close(updatesCh)
	reader, err := p.downloadRepository(updatesCh)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := p.processReader(updatesCh, reader, delay); err != nil {
		return err
	}

	if err := p.saveOutput(updatesCh); err != nil {
		return err
	}

	return nil
}

func (p *Processor) downloadRepository(updatesCh chan<- entity.Update) (io.ReadCloser, error) {
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

func (p *Processor) processReader(updatesCh chan<- entity.Update, reader io.Reader, delay time.Duration) error {
	updatesCh <- entity.Update{Description: "processing repository data"}

	tarReader, err := NewTarGzReader(reader)
	if err != nil {
		return err
	}
	defer tarReader.Close()

	for {
		hdr, err := tarReader.ReadNext()

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Error reading tarball: %v", err)
		}

		if hdr.IsDir || !strings.HasPrefix(hdr.Name, p.config.Dir) {
			continue
		}

		relativePath := strings.TrimPrefix(hdr.Name, p.config.Dir+"/")

		shouldProcess := true
		if len(p.config.Include) > 0 {
			shouldProcess = false
			for _, suffix := range p.config.Include {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = true
					break
				}
			}
		}
		if shouldProcess && len(p.config.Exclude) > 0 {
			for _, suffix := range p.config.Exclude {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = false
					break
				}
			}
		}

		if shouldProcess {
			content, err := tarReader.ReadContent()
			if err != nil {
				return fmt.Errorf("Error reading file: %s - %v", hdr.Name, err)
			}

			pathParts := strings.Split(relativePath, "/")

			current := p.data
			for i, part := range pathParts {
				if i == len(pathParts)-1 {
					current[part] = content
					time.Sleep(delay)
					updatesCh <- entity.Update{
						Description: fmt.Sprintf("mapping: %s", relativePath),
						Err:         nil,
					}
				} else {
					if _, exists := current[part]; !exists {
						current[part] = make(map[string]interface{})
					}

					var ok bool
					current, ok = current[part].(map[string]interface{})
					if !ok {
						return fmt.Errorf("Unexpected structure found on: %s", hdr.Name)
					}
				}
			}
		}
	}
	return nil
}

func (p *Processor) saveOutput(updatesCh chan<- entity.Update) error {
	updatesCh <- entity.Update{
		Description: "saving repository data",
	}

	fileName := fmt.Sprintf("%s%s.json", p.config.Repo, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(p.config.Output, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		updatesCh <- entity.Update{
			Description: fmt.Sprintf("Output file error: %v", err),
			Err:         err,
		}
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(p.data); err != nil {
		updatesCh <- entity.Update{
			Description: fmt.Sprintf("Output encoding error: %v", err),
			Err:         err,
		}
		return err
	}

	updatesCh <- entity.Update{
		Description: fmt.Sprintf("generating report: %s", filePath),
	}

	return nil
}
