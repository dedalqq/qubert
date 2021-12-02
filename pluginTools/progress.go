package pluginTools

type ProgressOptions struct {
	ID    string `json:"id"`
	Value uint   `json:"value"`
	Max   uint   `json:"max"`
}

type Progress struct {
	options ProgressOptions
}

func (p *Progress) Type() ElementType            { return ElementProgress }
func (p *Progress) MarshalJSON() ([]byte, error) { return MarshalJSON(p.Type(), p.options) }

func (p *Progress) SetID(id string) *Progress {
	p.options.ID = id

	return p
}

func (p *Progress) SetMax(max uint) *Progress {
	p.options.Max = max

	return p
}

func NewProgress(value uint) *Progress {
	return &Progress{
		options: ProgressOptions{
			Max:   100,
			Value: value,
		},
	}
}

type UpdateProgress struct {
	id string

	Value uint `json:"value"`
}

func (p *UpdateProgress) ElementID() string       { return p.id }
func (p *UpdateProgress) UpdateType() ElementType { return ElementProgress }

func NewUpdateProgress(id string, value uint) *UpdateProgress {
	return &UpdateProgress{
		id:    id,
		Value: value,
	}
}
