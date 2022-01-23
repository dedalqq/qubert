package pluginTools

import "fmt"

type TextOption struct {
	Text string `json:"text"`
}

type Text struct {
	options TextOption
}

func (t *Text) Type() ElementType            { return ElementTypeText }
func (t *Text) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func NewText(text string, a ...interface{}) *Text {
	return &Text{
		options: TextOption{
			Text: fmt.Sprintf(text, a...),
		},
	}
}
