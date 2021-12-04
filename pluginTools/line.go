package pluginTools

type LineOptions struct {
	Elements []Element `json:"elements"`
}

type Line struct {
	options LineOptions
}

func (l *Line) Type() ElementType            { return ElementLine }
func (l *Line) MarshalJSON() ([]byte, error) { return MarshalJSON(l.Type(), l.options) }

func (l *Line) Add(e ...Element) *Line {
	l.options.Elements = append(l.options.Elements, e...)

	return l
}

func NewLine(e ...Element) *Line {
	if e == nil {
		e = []Element{}
	}

	return &Line{
		options: LineOptions{
			Elements: e,
		},
	}
}
