package pluginTools

type InputEditOptions struct {
	Action Action `json:"action"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

type InputEdit struct {
	options InputEditOptions
}

func (i *InputEdit) Type() ElementType            { return ElementInputEdit }
func (i *InputEdit) MarshalJSON() ([]byte, error) { return MarshalJSON(i.Type(), i.options) }

func NewInputEdit(name string, value string, cmd string, args ...string) *InputEdit {
	return &InputEdit{
		options: InputEditOptions{
			Action: Action{
				CMD:  cmd,
				Args: args,
			},
			Name:  name,
			Value: value,
		},
	}
}
