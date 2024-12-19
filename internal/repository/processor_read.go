package repository

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func (p *Processor) read(reader io.Reader, stagger time.Duration) error {
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
			return err
		}
		if hdr.IsDir || !strings.HasPrefix(hdr.Name, p.config.Dir) {
			continue
		}

		relativePath := strings.TrimPrefix(hdr.Name, p.config.Dir+"/")

		shouldProcess := len(p.config.Include) == 0

		for _, suffix := range p.config.Include {
			if strings.HasSuffix(relativePath, suffix) {
				shouldProcess = true
				break
			}
		}

		if shouldProcess {
			for _, suffix := range p.config.Exclude {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = false
					break
				}
			}
		}

		if !shouldProcess {
			continue
		}

		content, err := tarReader.ReadContent()
		if err != nil {
			return err
		}

		pathParts := strings.Split(relativePath, "/")
		current := p.data
		for i, part := range pathParts {
			if i == len(pathParts)-1 {
				current[part] = content
				time.Sleep(stagger)
				p.update(fmt.Sprintf("mapping: %s", relativePath), nil)
				break
			}

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
	return nil
}
