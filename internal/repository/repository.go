package repository

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

func ProcessRepo(cfg *entity.Config, ch chan<- entity.Update, delay time.Duration) {
	defer close(ch)

	resp, err := http.Get(cfg.Url)
	if err != nil {
		ch <- entity.Update{
			Description: "Request error getting tarball",
			Err:         err,
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- entity.Update{
			Description: fmt.Sprintf("Request error getting tarbal with status code: %d", resp.StatusCode),
			Err:         fmt.Errorf("Status code: %d", resp.StatusCode),
		}
		return
	}

	data, err := TarballReader(cfg, resp.Body, ch, delay)
	if err != nil {
		ch <- entity.Update{
			Description: fmt.Sprintf("Tarball error: %v", err),
			Err:         err,
		}
		return
	}

	filePath := fmt.Sprintf("%s%s.json", cfg.Repo, time.Now().Format("20060102_150405"))

	if cfg.Output != "" {
		filePath = fmt.Sprintf("%s/%s", cfg.Output, filePath)
	}

	f, err := os.Create(filePath)
	if err != nil {
		ch <- entity.Update{
			Description: fmt.Sprintf("Output file error: %v", err),
			Err:         err,
		}
		return
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		ch <- entity.Update{
			Description: fmt.Sprintf("Output encoding error: %v", err),
			Err:         err,
		}
		return
	}

	ch <- entity.Update{
		Description: fmt.Sprintf("generating report: %s", filePath),
		Err:         nil,
	}
}

func TarballReader(
	cfg *entity.Config,
	r io.Reader,
	ch chan<- entity.Update,
	delay time.Duration,
) (map[string]interface{}, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error decompressing tarbal: %v", err)
	}
	defer gzipReader.Close()

	data := make(map[string]interface{})

	tarReader := tar.NewReader(gzipReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error reading tarball: %v", err)
		}

		if hdr.Typeflag == tar.TypeDir || !strings.HasPrefix(hdr.Name, cfg.Dir) {
			continue
		}

		relativePath := strings.TrimPrefix(hdr.Name, cfg.Dir+"/")

		shouldProcess := true
		if len(cfg.Include) > 0 {
			shouldProcess = false
			for _, suffix := range cfg.Include {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = true
					break
				}
			}
		}
		if shouldProcess && len(cfg.Exclude) > 0 {
			for _, suffix := range cfg.Exclude {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = false
					break
				}
			}
		}

		if shouldProcess {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				return nil, fmt.Errorf("Error reading file: %s - %v", hdr.Name, err)
			}

			pathParts := strings.Split(relativePath, "/")

			current := data
			for i, part := range pathParts {
				if i == len(pathParts)-1 {
					current[part] = buf.String()
					time.Sleep(delay)
					ch <- entity.Update{
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
						return nil, fmt.Errorf("Unexpected structure found on: %s", hdr.Name)
					}
				}
			}
		}
	}

	return data, nil
}
