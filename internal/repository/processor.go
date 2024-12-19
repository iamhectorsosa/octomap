package repository

import (
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
	reader, err := p.download(updatesCh)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := p.read(updatesCh, reader, delay); err != nil {
		return err
	}

	if err := p.save(updatesCh); err != nil {
		return err
	}

	return nil
}
