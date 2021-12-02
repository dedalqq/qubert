package pluginTools

type InputOptions struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	ElementID string `json:"id"`
	Value     string `json:"value,omitempty"`
	OnInput   Action `json:"oninput,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Input struct {
	options InputOptions
}

func (i *Input) Type() ElementType            { return ElementTypeInput }
func (i *Input) MarshalJSON() ([]byte, error) { return MarshalJSON(i.Type(), i.options) }

func (i *Input) ID() string {
	return i.options.ElementID
}

func (i *Input) SetID(id string) *Input {
	i.options.ElementID = id

	return i
}

func (i *Input) SetTypePassword() *Input {
	i.options.Type = "password"

	return i
}

func (i *Input) SetValue(v string) *Input {
	i.options.Value = v

	return i
}

func (i *Input) SetChangeAction(cmd string, args ...string) *Input {
	i.options.OnInput = Action{
		CMD:  cmd,
		Args: args,
	}

	return i
}

func (i *Input) SetErrorText(text string) *Input {
	i.options.Error = text

	return i
}

func NewInput(name string) *Input {
	return &Input{
		options: InputOptions{
			ElementID: name,
			Name:      name,
			Type:      "text",
		},
	}
}
