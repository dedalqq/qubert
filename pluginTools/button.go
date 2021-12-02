package pluginTools

type ButtonOptions struct {
	LabelData

	Action   Action       `json:"action"`
	Style    ElementStyle `json:"style"`
	Disabled bool         `json:"disabled,omitempty"`
}

type Button struct {
	options ButtonOptions
}

func (b *Button) Type() ElementType            { return ElementTypeButton }
func (b *Button) MarshalJSON() ([]byte, error) { return MarshalJSON(b.Type(), b.options) }

func (b *Button) SetStyle(style ElementStyle) *Button {
	b.options.Style = style

	return b
}

func (b *Button) SetDisabled(v bool) *Button {
	b.options.Disabled = v

	return b
}

func (b *Button) SetLinkStyle() *Button {
	b.options.Style = "link"

	return b
}

func (b *Button) SetImage(icon string) *Button {
	b.options.Icon = icon

	return b
}

func (b *Button) SetStrong(value bool) *Button {
	b.options.Strong = value

	return b
}

func NewButton(text string, action string, args ...string) *Button {
	return &Button{
		options: ButtonOptions{
			LabelData: LabelData{
				Text: text,
			},
			Style: StylePrimary,
			Action: Action{
				CMD:  action,
				Args: args,
			},
		},
	}
}

func NewImageButton(icon string, action string, args ...string) *Button {
	return &Button{
		options: ButtonOptions{
			LabelData: LabelData{
				Icon: icon,
			},
			Style: "link",
			Action: Action{
				CMD:  action,
				Args: args,
			},
		},
	}
}
