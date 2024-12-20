package processor

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/iamhectorsosa/octomap/pkg/archive"
)

func (p *Processor) read(reader io.Reader, stagger time.Duration) error {
	tarReader, err := archive.NewTarGzReader(reader)
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

		if hdr.IsDir {
			p.dirCount++
		}
		if hdr.IsFile {
			p.fileCount++
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
				p.dataFileCount++
				p.update(fmt.Sprintf("mapped: %s", relativePath))
				time.Sleep(stagger)
				break
			}

			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}

			var ok bool
			current, ok = current[part].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected structure found on: %s", hdr.Name)
			}

		}
	}
	return nil
}
