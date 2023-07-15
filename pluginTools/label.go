package pluginTools

import "fmt"

type LabelOption struct {
	LabelData

	Monospace bool `json:"monospace"`
}

type Label struct {
	options LabelOption
}

func (l *Label) Type() ElementType            { return ElementTypeLabel }
func (l *Label) MarshalJSON() ([]byte, error) { return MarshalJSON(l.Type(), l.options) }

func (l *Label) SetText(text string, a ...interface{}) *Label {
	l.options.Text = fmt.Sprintf(text, a...)

	return l
}

func (l *Label) SetImage(icon string) *Label {
	l.options.Icon = icon

	return l
}

func (l *Label) SetStrong(value bool) *Label {
	l.options.Strong = value

	return l
}

func (l *Label) SetMonospace(value bool) *Label {
	l.options.Monospace = value

	return l
}

func NewLabel(text string, a ...interface{}) *Label {
	return &Label{
		options: LabelOption{
			LabelData: LabelData{
				Text: fmt.Sprintf(text, a...),
			},
		},
	}
}
