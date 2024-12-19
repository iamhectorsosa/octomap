package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

func (p *Processor) save(updatesCh chan<- entity.Update) error {
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
