package processor

import (
	"time"
)

func New(config *Config, ch chan<- Update) *Processor {
	return &Processor{
		config: config,
		data:   make(RepositoryData),
		ch:     ch,
	}
}

func (p *Processor) Process(stagger time.Duration) (RepositoryData, error) {
	if p.ch != nil {
		defer close(p.ch)
	}

	reader, err := p.download()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	if err := p.read(reader, stagger); err != nil {
		return nil, err
	}

	if !p.config.Stdout {
		if err := p.save(); err != nil {
			return nil, err
		}
	}

	return p.data, nil
}
