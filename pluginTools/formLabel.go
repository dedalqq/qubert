package pluginTools

type FormLabelOption struct {
	Text string `json:"text"`
	For  string `json:"for"`
}

type FormLabel struct {
	options FormLabelOption
}

func (l *FormLabel) SetFor(id string) *FormLabel {
	l.options.For = id

	return l
}

func (l *FormLabel) Type() ElementType            { return ElementTypeFormLabel }
func (l *FormLabel) MarshalJSON() ([]byte, error) { return MarshalJSON(l.Type(), l.options) }

func NewFormLabel(text string) *FormLabel {
	return &FormLabel{
		options: FormLabelOption{
			Text: text,
		},
	}
}
