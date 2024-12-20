package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (p *Processor) save() error {
	fileName := fmt.Sprintf("%s%s.json", p.config.Repo, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(p.config.Output, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %q\n %v", filePath, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(p.data); err != nil {
		return fmt.Errorf("encoding file error: %v", err)
	}

	p.update(fmt.Sprintf("generated report: %s", filePath))
	return nil
}
