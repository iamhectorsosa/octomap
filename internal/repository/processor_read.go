package repository

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

func (p *Processor) read(updatesCh chan<- entity.Update, reader io.Reader, delay time.Duration) error {
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
