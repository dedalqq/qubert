package pluginTools

type TextareaEditOptions struct {
	Action Action `json:"action"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

type TextareaEdit struct {
	options TextareaEditOptions
}

func (t *TextareaEdit) Type() ElementType            { return ElementTextareaEdit }
func (t *TextareaEdit) MarshalJSON() ([]byte, error) { return MarshalJSON(t.Type(), t.options) }

func NewTextareaEdit(name string, value string, cmd string, args ...string) *TextareaEdit {
	return &TextareaEdit{
		options: TextareaEditOptions{
			Action: Action{
				CMD:  cmd,
				Args: args,
			},
			Name:  name,
			Value: value,
		},
	}
}
