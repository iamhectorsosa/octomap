package processor

func (p *Processor) update(description string) {
	if p.ch != nil {
		p.ch <- Update{Description: description}
	}
}

func (p *Processor) updateError(err error) {
	if p.ch != nil {
		p.ch <- Update{Err: err}
	}
}
