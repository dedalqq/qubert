package pluginTools

type SelectEditOptions struct {
	Name       string            `json:"name"`
	Value      string            `json:"value"`
	Options    map[string]string `json:"options"`
	Action     *Action           `json:"action,omitempty"`
	BadgeStyle ElementStyle      `json:"badge-style"`
}

type SelectEdit struct {
	options SelectEditOptions
}

func (s *SelectEdit) Type() ElementType            { return ElementTypeSelectEdit }
func (s *SelectEdit) MarshalJSON() ([]byte, error) { return MarshalJSON(s.Type(), s.options) }

func (s *SelectEdit) SetValue(v string) *SelectEdit {
	s.options.Value = v

	return s
}

func (s *SelectEdit) AddOption(value string) *SelectEdit {
	s.options.Options[value] = value

	return s
}

func (s *SelectEdit) AddNamedOption(name string, value string) *SelectEdit {
	s.options.Options[value] = name

	return s
}

func (s *SelectEdit) SetBadgeStyle(style ElementStyle) *SelectEdit {
	s.options.BadgeStyle = style

	return s
}

func NewSelectEdit(name string, cmd string, args ...string) *SelectEdit {
	return &SelectEdit{
		options: SelectEditOptions{
			Name:    name,
			Options: map[string]string{},
			Action: &Action{
				CMD:  cmd,
				Args: args,
			},
		},
	}
}
