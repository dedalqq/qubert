package pluginTools

type BadgeOption struct {
	Text  string       `json:"text"`
	Style ElementStyle `json:"style"`
}

type Badge struct {
	options BadgeOption
}

func (b *Badge) Type() ElementType            { return ElementTypeBadge }
func (b *Badge) MarshalJSON() ([]byte, error) { return MarshalJSON(b.Type(), b.options) }

func (b *Badge) SetStyle(style ElementStyle) *Badge {
	b.options.Style = style

	return b
}

func NewBadge(text string) *Badge {
	return &Badge{
		options: BadgeOption{
			Text:  text,
			Style: StylePrimary,
		},
	}
}
