package processor

import (
	"fmt"
	"time"
)

func New(config *Config, ch chan<- Update) *Processor {
	return &Processor{
		config:        config,
		data:          make(RepositoryData),
		ch:            ch,
		dirCount:      0,
		fileCount:     0,
		dataFileCount: 0,
	}
}

func (p *Processor) Process(stagger time.Duration) (RepositoryData, error) {
	if p.ch != nil {
		defer close(p.ch)
	}

	reader, err := p.download()
	if err != nil {
		p.updateError(err)
		return nil, err
	}
	defer reader.Close()

	if err := p.read(reader, stagger); err != nil {
		p.updateError(err)
		return nil, err
	}

	p.update(fmt.Sprintf("found: %d directories and %d files", p.dirCount, p.fileCount))
	p.update(fmt.Sprintf("prepared: %d out of %d files for report", p.dataFileCount, p.fileCount))

	if !p.config.Stdout {
		if err := p.save(); err != nil {
			p.updateError(err)
			return nil, err
		}
	}

	return p.data, nil
}
