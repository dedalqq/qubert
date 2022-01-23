package pluginTools

type LineElement struct {
	Element Element `json:"element"`
	Flex    int     `json:"flex,omitempty"`
}

type LineOptions struct {
	Elements []LineElement `json:"items"`
}

type Line struct {
	options LineOptions
}

func (l *Line) Type() ElementType            { return ElementLine }
func (l *Line) MarshalJSON() ([]byte, error) { return MarshalJSON(l.Type(), l.options) }

func (l *Line) Add(e ...Element) *Line {
	for _, el := range e {
		l.options.Elements = append(l.options.Elements, LineElement{
			Element: el,
		})
	}

	return l
}

func (l *Line) AddWithFlex(flex int, el Element) *Line {
	l.options.Elements = append(l.options.Elements, LineElement{
		Flex:    flex,
		Element: el,
	})

	return l
}

func NewLine(e ...Element) *Line {
	line := &Line{
		options: LineOptions{
			Elements: []LineElement{},
		},
	}

	for _, el := range e {
		line.options.Elements = append(line.options.Elements, LineElement{
			Element: el,
		})
	}

	return line
}
