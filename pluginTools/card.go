package pluginTools

type CardOptions struct {
	Header     *LabelData `json:"header,omitempty"`
	Additional Element    `json:"additional,omitempty"`
	Body       Element    `json:"body"`
}

type Card struct {
	options CardOptions
}

func (c *Card) Type() ElementType            { return ElementTypeCard }
func (c *Card) MarshalJSON() ([]byte, error) { return MarshalJSON(c.Type(), c.options) }

func (c *Card) SetHeaderIcon(icon string) *Card {
	if c.options.Header == nil {
		c.options.Header = &LabelData{}
	}

	c.options.Header.Icon = icon

	return c
}

func (c *Card) SetAdditionalHeader(el Element) *Card {
	c.options.Additional = el

	return c
}

func NewCard(body Element) *Card {
	return &Card{
		options: CardOptions{
			Body: body,
		},
	}
}

func NewCardWithTitle(headerTitle string, body Element) *Card {
	return &Card{
		options: CardOptions{
			Header: &LabelData{
				Text: headerTitle,
			},
			Body: body,
		},
	}
}
