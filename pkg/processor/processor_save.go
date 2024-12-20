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
		p.update(fmt.Sprintf("output file error: %v", err), err)
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(p.data); err != nil {
		p.update(fmt.Sprintf("encoding file error: %v", err), err)
		return err
	}

	p.update(fmt.Sprintf("generating report: %s", filePath), nil)
	return nil
}
