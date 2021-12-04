package pluginTools

type SwitchOptions struct {
	Name      string `json:"name"`
	ElementID string `json:"id"`
}

type Switch struct {
	options SwitchOptions
}

func (s *Switch) Type() ElementType            { return ElementSwitch }
func (s *Switch) MarshalJSON() ([]byte, error) { return MarshalJSON(s.Type(), s.options) }

func (s *Switch) ID() string {
	return s.options.ElementID
}

func NewSwitch(name string) *Switch {
	return &Switch{
		options: SwitchOptions{
			ElementID: name,
			Name:      name,
		},
	}
}
