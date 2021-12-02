package pluginTools

type SelectOptions struct {
	Name      string            `json:"name"`
	ElementID string            `json:"id"`
	Value     string            `json:"value,omitempty"`
	Options   map[string]string `json:"options"`
	Error     string            `json:"error,omitempty"`
}

type Select struct {
	options SelectOptions
}

func (s *Select) Type() ElementType            { return ElementTypeSelect }
func (s *Select) MarshalJSON() ([]byte, error) { return MarshalJSON(s.Type(), s.options) }

func (s *Select) ID() string {
	return s.options.ElementID
}

func (s *Select) SetID(id string) *Select {
	s.options.ElementID = id

	return s
}

func (s *Select) SetValue(v string) *Select {
	s.options.Value = v

	return s
}

func (s *Select) AddOption(value string) *Select {
	s.options.Options[value] = value

	return s
}

func (s *Select) AddNamedOption(name string, value string) *Select {
	s.options.Options[value] = name

	return s
}

func (s *Select) SetErrorText(text string) *Select {
	s.options.Error = text

	return s
}

func NewSelect(name string) *Select {
	return &Select{
		options: SelectOptions{
			Name:      name,
			ElementID: name,
			Options:   map[string]string{},
		},
	}
}
