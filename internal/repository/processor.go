package repository

import (
	"time"

	"github.com/iamhectorsosa/octomap/internal/entity"
)

type Processor struct {
	config *entity.Config
	data   map[string]interface{}
	ch     chan<- entity.Update
}

func NewProcessor(config *entity.Config, ch chan<- entity.Update) *Processor {
	return &Processor{
		config: config,
		data:   make(map[string]interface{}),
		ch:     ch,
	}
}

func (p *Processor) Process(stagger time.Duration) error {
	defer close(p.ch)
	reader, err := p.download()
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := p.read(reader, stagger); err != nil {
		return err
	}

	if err := p.save(); err != nil {
		return err
	}

	return nil
}
