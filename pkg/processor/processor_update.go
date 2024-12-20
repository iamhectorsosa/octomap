package processor

func (p *Processor) update(description string, err error) {
	if p.ch != nil {
		p.ch <- Update{
			Description: description,
			Err:         err,
		}
	}
}
