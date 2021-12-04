package pluginTools

type InputOptions struct {
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	ElementID string      `json:"id"`
	Value     interface{} `json:"value,omitempty"`
	Error     string      `json:"error,omitempty"`
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

type NumberInput struct {
	options InputOptions
}

func (i *NumberInput) Type() ElementType            { return ElementTypeInput }
func (i *NumberInput) MarshalJSON() ([]byte, error) { return MarshalJSON(i.Type(), i.options) }

func (i *NumberInput) ID() string {
	return i.options.ElementID
}

func (i *NumberInput) SetID(id string) *NumberInput {
	i.options.ElementID = id

	return i
}

func (i *NumberInput) SetValue(v int) *NumberInput {
	i.options.Value = v

	return i
}

func (i *NumberInput) SetErrorText(text string) *NumberInput {
	i.options.Error = text

	return i
}

func NewNumberInput(name string) *NumberInput {
	return &NumberInput{
		options: InputOptions{
			ElementID: name,
			Name:      name,
			Type:      "number",
		},
	}
}
