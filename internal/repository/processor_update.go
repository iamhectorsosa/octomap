package repository

import "github.com/iamhectorsosa/octomap/internal/entity"

func (p *Processor) update(description string, err error) {
	p.ch <- entity.Update{
		Description: description,
		Err:         err,
	}
}
