package pluginTools

type SwitchOptions struct {
	Name      string  `json:"name"`
	ElementID string  `json:"id"`
	Checked   bool    `json:"checked,omitempty"`
	Action    *Action `json:"action,omitempty"`
}

type Switch struct {
	options SwitchOptions
}

func (s *Switch) Type() ElementType            { return ElementSwitch }
func (s *Switch) MarshalJSON() ([]byte, error) { return MarshalJSON(s.Type(), s.options) }

func (s *Switch) ID() string {
	return s.options.ElementID
}

func (s *Switch) SetAction(a *Action) *Switch {
	s.options.Action = a

	return s
}

func (s *Switch) SetValue(v bool) *Switch {
	s.options.Checked = v

	return s
}

func NewSwitch(name string) *Switch {
	return &Switch{
		options: SwitchOptions{
			ElementID: name,
			Name:      name,
		},
	}
}
