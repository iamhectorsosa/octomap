package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	dataOutputFile   = "data.json"
	githubTarballURL = "https://github.com/%s/archive/refs/heads/%s.tar.gz"
)

func main() {
	inputRepo := "iamhectorsosa/f-server"
	inputTargetDir := "internal"
	inputBranch := "main"

	whitelistSuffixes := []string{".sql"}
	blacklistSuffixes := []string{".go"}

	repoParts := strings.Split(inputRepo, "/")
	if len(repoParts) != 2 {
		log.Fatalf("Invalid repository format. Expected 'user/repo', got: %q", inputRepo)
	}

	repo := repoParts[1]
	tarballURL := fmt.Sprintf(githubTarballURL, inputRepo, inputBranch)
	targetDir := fmt.Sprintf("%s-%s/%s", repo, inputBranch, inputTargetDir)

	resp, err := http.Get(tarballURL)
	if err != nil {
		log.Fatalf("Error fetching tarball: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch tarball: %s", resp.Status)
	}

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Error decompressing tarball: %v", err)
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
			log.Fatalf("Error reading tarball: %v", err)
		}

		if strings.HasPrefix(hdr.Name, targetDir) {
			relativePath := strings.TrimPrefix(hdr.Name, targetDir+"/")

			if relativePath == "" {
				continue
			}

			shouldProcess := false
			for _, suffix := range whitelistSuffixes {
				if strings.HasSuffix(relativePath, suffix) {
					shouldProcess = true
					break
				}
			}

			if shouldProcess {
				for _, suffix := range blacklistSuffixes {
					if strings.HasSuffix(relativePath, suffix) {
						shouldProcess = false
						break
					}
				}
			}

			if shouldProcess {
				var buf bytes.Buffer
				if _, err := io.Copy(&buf, tarReader); err != nil {
					log.Printf("Warning: error reading file %s: %v", hdr.Name, err)
					continue
				}

				pathParts := strings.Split(relativePath, "/")

				current := data
				for i, part := range pathParts {
					if i == len(pathParts)-1 {
						current[part] = buf.String()
					} else {
						if _, exists := current[part]; !exists {
							current[part] = make(map[string]interface{})
						}

						var ok bool
						current, ok = current[part].(map[string]interface{})
						if !ok {
							log.Printf("Warning: unexpected structure for %s", hdr.Name)
							break
						}
					}
				}
			}
		}
	}

	f, err := os.Create(dataOutputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Error encoding JSON: %v", err)
	}

	fmt.Printf("Successfully extracted repository structure to %s\n", dataOutputFile)
}
