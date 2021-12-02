package pluginTools

type CardOptions struct {
	Header     LabelData `json:"header"`
	Additional Element   `json:"additional,omitempty"`
	Body       Element   `json:"body"`
}

type Card struct {
	options CardOptions
}

func (c *Card) Type() ElementType            { return ElementTypeCard }
func (c *Card) MarshalJSON() ([]byte, error) { return MarshalJSON(c.Type(), c.options) }

func (c *Card) SetHeaderIcon(icon string) *Card {
	c.options.Header.Icon = icon

	return c
}

func (c *Card) SetAdditionalHeader(el Element) *Card {
	c.options.Additional = el

	return c
}

func NewCard(headerTitle string, body Element) *Card {
	return &Card{
		options: CardOptions{
			Header: LabelData{
				Text: headerTitle,
			},
			Body: body,
		},
	}
}
